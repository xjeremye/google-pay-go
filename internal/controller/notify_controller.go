package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-pay-core/internal/alipay"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/mq"
	"github.com/golang-pay-core/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NotifyController 回调控制器
type NotifyController struct {
	orderService  *service.OrderService
	notifyService *service.OrderNotifyService
	mqClient      *mq.RocketMQClient // RocketMQ 客户端（可选）
}

// NewNotifyController 创建回调控制器
func NewNotifyController() *NotifyController {
	// 初始化 RocketMQ 客户端（如果启用）
	mqClient, err := mq.NewRocketMQClient()
	if err != nil {
		logger.Logger.Warn("初始化 RocketMQ 客户端失败，回调将使用同步处理",
			zap.Error(err))
		mqClient = nil
	}

	return &NotifyController{
		orderService:  service.NewOrderService(),
		notifyService: service.NewOrderNotifyService(),
		mqClient:      mqClient,
	}
}

// AlipayNotify 支付宝回调接口
// 参考 Python: /api/pay/order/notify/{plugin_type}/{product_id}/
// @Summary 支付宝回调
// @Description 处理支付宝支付回调通知
// @Tags 支付回调
// @Accept application/x-www-form-urlencoded
// @Produce text/plain
// @Param plugin_type path string true "插件类型" example:"alipay_phone"
// @Param product_id path string true "产品ID" example:"1"
// @Success 200 {string} string "success"
// @Failure 400 {string} string "参数错误"
// @Router /api/pay/order/notify/{plugin_type}/{product_id}/ [post]
func (c *NotifyController) AlipayNotify(ctx *gin.Context) {
	// 获取路径参数
	pluginType := ctx.Param("plugin_type")
	productID := ctx.Param("product_id")

	if pluginType == "" || productID == "" {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}

	// Mock插件特殊处理：不需要验证签名和产品信息
	if pluginType == "mock" || pluginType == "mock_alipay" {
		c.handleMockNotify(ctx, pluginType, productID)
		return
	}

	// 获取产品信息
	product, err := alipay.GetAlipayProductByID(productID)
	if err != nil {
		logger.Logger.Warn("获取产品信息失败",
			zap.String("product_id", productID),
			zap.Error(err))
		ctx.String(http.StatusBadRequest, "产品不存在")
		return
	}

	// 创建支付宝客户端（用于验证签名）
	alipayClient, err := alipay.NewClient(product, "", false)
	if err != nil {
		logger.Logger.Warn("创建支付宝客户端失败",
			zap.String("product_id", productID),
			zap.Error(err))
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 解析回调参数
	// 支付宝回调通常是 POST Form 格式，也可能是 GET Query
	var params map[string]string
	if ctx.Request.Method == "POST" {
		// POST 请求：尝试从 Form 读取
		contentType := ctx.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// Form 格式
			if err := ctx.Request.ParseForm(); err != nil {
				ctx.String(http.StatusBadRequest, "参数解析失败")
				return
			}
			params = make(map[string]string)
			for k, v := range ctx.Request.PostForm {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}
		} else {
			// 可能是 JSON 或其他格式，尝试从 Query 读取
			params = make(map[string]string)
			for k, v := range ctx.Request.URL.Query() {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}
		}
	} else {
		// GET 请求：从 Query 读取
		params = make(map[string]string)
		for k, v := range ctx.Request.URL.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
	}

	if len(params) == 0 {
		ctx.String(http.StatusBadRequest, "参数为空")
		return
	}

	// 验证签名
	if !alipay.VerifyNotify(params, alipayClient.AlipayPublicKey) {
		logger.Logger.Warn("支付宝回调签名验证失败",
			zap.String("product_id", productID),
			zap.Any("params", params))
		ctx.String(http.StatusBadRequest, "签名验证失败")
		return
	}

	// 解析回调数据
	notifyData, err := alipay.ParseNotifyParams(params)
	if err != nil {
		logger.Logger.Warn("解析回调参数失败",
			zap.String("product_id", productID),
			zap.Error(err))
		ctx.String(http.StatusBadRequest, "参数解析失败")
		return
	}

	// 如果启用了 RocketMQ，使用消息队列处理；否则使用 goroutine
	if c.mqClient != nil && c.mqClient.IsEnabled() {
		// 构建支付宝回调消息
		notifyMsg := &mq.AlipayNotifyMessage{
			PluginType: pluginType,
			ProductID:  productID,
			Params:     params, // 保留原始参数（包含签名）
			NotifyData: &mq.AlipayNotifyData{
				OutTradeNo:    notifyData.OutTradeNo,
				TradeNo:       notifyData.TradeNo,
				TradeStatus:   notifyData.TradeStatus,
				TotalAmount:   notifyData.TotalAmount,
				ReceiptAmount: 0,  // 如果回调中有，可以从 params 解析
				BuyerID:       "", // 如果回调中有，可以从 params 解析
				BuyerLogonID:  "", // 如果回调中有，可以从 params 解析
				SellerID:      "", // 如果回调中有，可以从 params 解析
				GmtPayment:    "", // 如果回调中有，可以从 params 解析
			},
			ReceivedAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		// 解析更多字段（如果存在）
		if buyerID, ok := params["buyer_id"]; ok {
			notifyMsg.NotifyData.BuyerID = buyerID
		}
		if buyerLogonID, ok := params["buyer_logon_id"]; ok {
			notifyMsg.NotifyData.BuyerLogonID = buyerLogonID
		}
		if sellerID, ok := params["seller_id"]; ok {
			notifyMsg.NotifyData.SellerID = sellerID
		}
		if gmtPayment, ok := params["gmt_payment"]; ok {
			notifyMsg.NotifyData.GmtPayment = gmtPayment
		}
		if receiptAmount, ok := params["receipt_amount"]; ok {
			var amount float64
			if _, err := fmt.Sscanf(receiptAmount, "%f", &amount); err == nil {
				notifyMsg.NotifyData.ReceiptAmount = int(amount * 100)
			}
		}

		// 发送消息到 RocketMQ
		if err := c.mqClient.SendMessage(ctx.Request.Context(), "alipay-notify", "notify", notifyMsg); err != nil {
			logger.Logger.Error("发送支付宝回调消息失败，降级为同步处理",
				zap.String("product_id", productID),
				zap.String("out_trade_no", notifyData.OutTradeNo),
				zap.Error(err))
			// 降级为同步处理
			go c.handleAlipayNotify(context.Background(), notifyData, productID)
		}
	} else {
		// 未启用 RocketMQ，使用 goroutine（原有逻辑）
		go c.handleAlipayNotify(context.Background(), notifyData, productID)
	}

	// 立即返回 success（支付宝要求）
	ctx.String(http.StatusOK, "success")
}

// handleMockNotify 处理Mock插件回调（用于压测）
func (c *NotifyController) handleMockNotify(ctx *gin.Context, pluginType, productID string) {
	// 解析回调参数（与支付宝回调格式相同）
	var params map[string]string
	if ctx.Request.Method == "POST" {
		contentType := ctx.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			if err := ctx.Request.ParseForm(); err != nil {
				ctx.String(http.StatusBadRequest, "参数解析失败")
				return
			}
			params = make(map[string]string)
			for k, v := range ctx.Request.PostForm {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}
		} else {
			params = make(map[string]string)
			for k, v := range ctx.Request.URL.Query() {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}
		}
	} else {
		params = make(map[string]string)
		for k, v := range ctx.Request.URL.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
	}

	if len(params) == 0 {
		ctx.String(http.StatusBadRequest, "参数为空")
		return
	}

	// Mock插件不需要验证签名，直接解析回调数据
	notifyData, err := alipay.ParseNotifyParams(params)
	if err != nil {
		logger.Logger.Warn("解析Mock回调参数失败",
			zap.String("product_id", productID),
			zap.Error(err))
		ctx.String(http.StatusBadRequest, "参数解析失败")
		return
	}

	// 使用相同的处理逻辑（异步处理）
	go c.handleAlipayNotify(context.Background(), notifyData, productID)

	// 立即返回 success
	ctx.String(http.StatusOK, "success")
}

// handleAlipayNotify 处理支付宝回调
func (c *NotifyController) handleAlipayNotify(ctx context.Context, notifyData *alipay.NotifyData, productID string) {
	// 查询订单（通过 out_trade_no，即商户订单号）
	var order models.Order
	if err := database.DB.Where("out_order_no = ?", notifyData.OutTradeNo).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Logger.Warn("订单不存在",
				zap.String("out_order_no", notifyData.OutTradeNo))
			return
		}
		logger.Logger.Error("查询订单失败",
			zap.String("out_order_no", notifyData.OutTradeNo),
			zap.Error(err))
		return
	}

	// 查询订单详情
	var orderDetail models.OrderDetail
	if err := database.DB.Where("order_id = ?", order.ID).First(&orderDetail).Error; err != nil {
		logger.Logger.Error("查询订单详情失败",
			zap.String("order_id", order.ID),
			zap.Error(err))
		return
	}

	// 验证产品ID是否匹配
	if orderDetail.ProductID != productID {
		logger.Logger.Warn("产品ID不匹配",
			zap.String("order_id", order.ID),
			zap.String("order_product_id", orderDetail.ProductID),
			zap.String("notify_product_id", productID))
		return
	}

	// 验证金额是否匹配（如果回调中有金额）
	if notifyData.TotalAmount > 0 && notifyData.TotalAmount != order.Money {
		logger.Logger.Warn("金额不匹配",
			zap.String("order_id", order.ID),
			zap.Int("order_money", order.Money),
			zap.Int("notify_amount", notifyData.TotalAmount))
		// 金额不匹配，但不阻止处理（可能是部分退款等情况）
	}

	// 检查订单状态，避免重复处理
	// 如果订单已经是成功状态，且回调也是成功，则跳过
	if order.OrderStatus == models.OrderStatusPaid {
		if notifyData.TradeStatus == "TRADE_SUCCESS" || notifyData.TradeStatus == "TRADE_FINISHED" {
			logger.Logger.Info("订单已处理，跳过重复回调",
				zap.String("order_id", order.ID),
				zap.String("trade_status", notifyData.TradeStatus))
			// 仍然更新 ticket_no（如果还没有）
			if notifyData.TradeNo != "" && orderDetail.TicketNo == "" {
				database.DB.Model(&models.OrderDetail{}).
					Where("order_id = ?", order.ID).
					Update("ticket_no", notifyData.TradeNo)
			}
			return
		}
	}

	// 根据交易状态更新订单
	// 参考 Python: 根据 trade_status 判断订单状态
	var newStatus int
	switch notifyData.TradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		// 交易成功
		newStatus = models.OrderStatusPaid
	case "TRADE_CLOSED":
		// 交易关闭（可能是退款或取消）
		newStatus = models.OrderStatusFailed
	default:
		// 其他状态不处理
		logger.Logger.Info("未处理的交易状态",
			zap.String("order_id", order.ID),
			zap.String("trade_status", notifyData.TradeStatus))
		return
	}

	// 更新订单状态（包含 ticket_no 更新）
	// UpdateOrderStatus 方法会在事务中更新 ticket_no，这里不需要重复更新
	if err := c.orderService.UpdateOrderStatus(ctx, order.ID, newStatus, notifyData.TradeNo); err != nil {
		logger.Logger.Error("更新订单状态失败",
			zap.String("order_id", order.ID),
			zap.Int("status", newStatus),
			zap.Error(err))
		return
	}

	// 通知商户（异步执行）
	if newStatus == models.OrderStatusPaid {
		// 重新查询订单和详情（获取最新状态）
		var updatedOrder models.Order
		var updatedDetail models.OrderDetail
		if err := database.DB.Where("id = ?", order.ID).First(&updatedOrder).Error; err == nil {
			if err := database.DB.Where("order_id = ?", order.ID).First(&updatedDetail).Error; err == nil {
				go c.notifyService.NotifyMerchant(context.Background(), &updatedOrder, &updatedDetail)
			}
		}
	}

	logger.Logger.Info("支付宝回调处理成功",
		zap.String("order_id", order.ID),
		zap.String("out_order_no", notifyData.OutTradeNo),
		zap.String("trade_no", notifyData.TradeNo),
		zap.String("trade_status", notifyData.TradeStatus),
		zap.Int("new_status", newStatus))
}
