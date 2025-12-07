package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/response"
	"github.com/golang-pay-core/internal/service"
	"go.uber.org/zap"
)

type OrderController struct {
	orderService      *service.OrderService
	merchantService   *service.MerchantService
	payChannelService *service.PayChannelService
}

// NewOrderController 创建订单控制器
func NewOrderController() *OrderController {
	return &OrderController{
		orderService:      service.NewOrderService(),
		merchantService:   service.NewMerchantService(),
		payChannelService: service.NewPayChannelService(),
	}
}

// CreateOrder 创建订单（支持 POST 和 GET）
// @Summary 创建订单
// @Description 创建支付订单，支持以下方式：1. POST JSON 2. POST Form 3. GET Query String
// @Tags 订单
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param mchId query int false "商户ID" example:"1"
// @Param channelId query int false "渠道ID" example:"1"
// @Param mchOrderNo query string false "商户订单号" example:"ORD20240101001"
// @Param amount query int false "金额（分）" example:"10000"
// @Param notifyUrl query string false "通知地址" example:"https://example.com/notify"
// @Param jumpUrl query string false "跳转地址" example:"https://example.com/jump"
// @Param extra query string false "额外参数" example:"{}"
// @Param compatible query int false "兼容模式 0/1" example:"0"
// @Param test query bool false "测试模式" example:"false"
// @Param sign query string false "签名" example:"ABC123..."
// @Param request body CreateOrderRequest false "订单信息（POST 方式）"
// @Success 200 {object} response.Response{data=object} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/orders [post]
// @Router /api/v1/orders [get]
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req service.CreateOrderRequest
	var rawSignData map[string]interface{}
	var requestBody string

	// 记录请求方法
	req.RequestMethod = ctx.Request.Method

	// 根据请求方法选择不同的参数绑定方式
	if ctx.Request.Method == "GET" {
		// GET 请求：从 Query String 读取参数
		rawSignData = c.parseQueryParams(ctx, &req)
		// 将查询参数转换为JSON字符串
		if bodyJSON, err := json.Marshal(rawSignData); err == nil {
			requestBody = string(bodyJSON)
		}
	} else {
		// POST 请求：尝试从 JSON Body 或 Form 读取参数
		contentType := ctx.GetHeader("Content-Type")
		if contentType == "application/json" || contentType == "application/json; charset=utf-8" {
			// JSON 格式：读取原始请求体
			bodyBytes, err := io.ReadAll(ctx.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// 恢复请求体供 ShouldBindJSON 使用
				ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			if err := ctx.ShouldBindJSON(&req); err != nil {
				response.Fail(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
				return
			}
			// 构建原始签名数据
			rawSignData = make(map[string]interface{})
			rawSignData["mchId"] = req.MerchantID
			rawSignData["channelId"] = req.ChannelID
			rawSignData["mchOrderNo"] = req.OutOrderNo
			rawSignData["amount"] = req.Money
			rawSignData["notifyUrl"] = req.NotifyURL
			rawSignData["jumpUrl"] = req.JumpURL
			rawSignData["extra"] = req.Extra
			rawSignData["compatible"] = req.Compatible
			rawSignData["test"] = req.Test
			rawSignData["sign"] = req.Sign
		} else {
			// Form 格式（application/x-www-form-urlencoded 或 multipart/form-data）
			rawSignData = c.parseFormParams(ctx, &req)
			// 将表单参数转换为JSON字符串
			if bodyJSON, err := json.Marshal(rawSignData); err == nil {
				requestBody = string(bodyJSON)
			}
		}
	}

	req.RawSignData = rawSignData
	req.RequestBody = requestBody

	// 将签名原始数据转换为JSON字符串（用于日志）
	// 参考 Python: ctx.sign_raw 应该是签名原始数据的字符串表示
	// 注意：这里先保存完整的 rawSignData（包含 sign 字段），
	// 后续在 validateSign 中会生成正确的 signRaw（用于签名的原始数据，不包含 sign）
	// 但为了保持与 Python 代码一致，这里先保存完整数据
	if signRawJSON, err := json.Marshal(rawSignData); err == nil {
		req.SignRaw = string(signRawJSON)
	}

	// 创建订单
	orderResp, orderErr := c.orderService.CreateOrder(ctx.Request.Context(), &req)
	if orderErr != nil {
		// 订单事务失败，不插入日志
		// 返回业务错误码和消息
		response.FailWithCode(ctx, orderErr.Code, orderErr.Message)
		return
	}

	// 订单事务成功，异步插入一条订单日志（包含请求和响应信息）
	go c.recordOrderLogSuccess(ctx.Request.Context(), &req, orderResp)
	response.Success(ctx, orderResp)
}

// recordOrderLogSuccess 记录订单日志的成功响应
// 参考 Python: 成功情况下记录响应到 order_log
// 重要：订单事务成功后异步插入一条订单日志（包含请求和响应信息）
// 注意：order_log 只创建一次，不更新，不需要检查订单是否存在（因为事务已成功）
func (c *OrderController) recordOrderLogSuccess(ctx context.Context, req *service.CreateOrderRequest, orderResp *service.CreateOrderResponse) {
	if req.OutOrderNo == "" {
		return
	}

	// 构建成功响应（参考 response.Success 的格式）
	successResponse := response.Response{
		Code:    200,
		Message: "success",
		Data:    orderResp,
	}

	// 将成功响应转换为JSON字符串
	responseJSON, err := json.Marshal(successResponse)
	if err != nil {
		return
	}

	// 创建 order_log（包含请求和响应信息，只创建不更新）
	// 不需要检查订单是否存在，因为只有在事务成功后才调用此方法
	now := time.Now()
	orderLog := &models.OrderLog{
		OutOrderNo:     req.OutOrderNo,
		SignRaw:        req.SignRaw, // JSON 格式的原始签名数据
		Sign:           "",          // 如果 validateSign 已执行，Service 层会设置这个值
		RequestBody:    req.RequestBody,
		RequestMethod:  req.RequestMethod,
		ResponseCode:   "200",
		JSONResult:     string(responseJSON),
		CreateDatetime: &now,
	}

	if err := database.DB.Create(orderLog).Error; err != nil {
		// 创建失败，记录警告但不影响主流程
		logger.Logger.Warn("创建订单日志失败",
			zap.String("out_order_no", req.OutOrderNo),
			zap.Error(err))
	}
}

// parseQueryParams 解析 GET 请求的查询参数
func (c *OrderController) parseQueryParams(ctx *gin.Context, req *service.CreateOrderRequest) map[string]interface{} {
	rawSignData := make(map[string]interface{})

	// 从 Query String 读取参数
	if mchId := ctx.Query("mchId"); mchId != "" {
		if id, err := strconv.Atoi(mchId); err == nil {
			req.MerchantID = id
			rawSignData["mchId"] = id
		}
	}

	if channelId := ctx.Query("channelId"); channelId != "" {
		if id, err := strconv.Atoi(channelId); err == nil {
			req.ChannelID = id
			rawSignData["channelId"] = id
		}
	}

	if mchOrderNo := ctx.Query("mchOrderNo"); mchOrderNo != "" {
		req.OutOrderNo = mchOrderNo
		rawSignData["mchOrderNo"] = mchOrderNo
	}

	if amount := ctx.Query("amount"); amount != "" {
		if money, err := strconv.Atoi(amount); err == nil {
			req.Money = money
			rawSignData["amount"] = money
		}
	}

	if notifyUrl := ctx.Query("notifyUrl"); notifyUrl != "" {
		req.NotifyURL = notifyUrl
		rawSignData["notifyUrl"] = notifyUrl
	}

	if jumpUrl := ctx.Query("jumpUrl"); jumpUrl != "" {
		req.JumpURL = jumpUrl
		rawSignData["jumpUrl"] = jumpUrl
	}

	if extra := ctx.Query("extra"); extra != "" {
		req.Extra = extra
		rawSignData["extra"] = extra
	}

	if compatible := ctx.Query("compatible"); compatible != "" {
		if comp, err := strconv.Atoi(compatible); err == nil {
			req.Compatible = comp
			rawSignData["compatible"] = comp
		}
	}

	if test := ctx.Query("test"); test != "" {
		req.Test = (test == "true" || test == "1")
		rawSignData["test"] = req.Test
	}

	if sign := ctx.Query("sign"); sign != "" {
		req.Sign = sign
		rawSignData["sign"] = sign
	}

	return rawSignData
}

// parseFormParams 解析 POST 请求的 Form 参数
func (c *OrderController) parseFormParams(ctx *gin.Context, req *service.CreateOrderRequest) map[string]interface{} {
	rawSignData := make(map[string]interface{})

	// 从 Form 读取参数
	if mchId := ctx.PostForm("mchId"); mchId != "" {
		if id, err := strconv.Atoi(mchId); err == nil {
			req.MerchantID = id
			rawSignData["mchId"] = id
		}
	}

	if channelId := ctx.PostForm("channelId"); channelId != "" {
		if id, err := strconv.Atoi(channelId); err == nil {
			req.ChannelID = id
			rawSignData["channelId"] = id
		}
	}

	if mchOrderNo := ctx.PostForm("mchOrderNo"); mchOrderNo != "" {
		req.OutOrderNo = mchOrderNo
		rawSignData["mchOrderNo"] = mchOrderNo
	}

	if amount := ctx.PostForm("amount"); amount != "" {
		if money, err := strconv.Atoi(amount); err == nil {
			req.Money = money
			rawSignData["amount"] = money
		}
	}

	if notifyUrl := ctx.PostForm("notifyUrl"); notifyUrl != "" {
		req.NotifyURL = notifyUrl
		rawSignData["notifyUrl"] = notifyUrl
	}

	if jumpUrl := ctx.PostForm("jumpUrl"); jumpUrl != "" {
		req.JumpURL = jumpUrl
		rawSignData["jumpUrl"] = jumpUrl
	}

	if extra := ctx.PostForm("extra"); extra != "" {
		req.Extra = extra
		rawSignData["extra"] = extra
	}

	if compatible := ctx.PostForm("compatible"); compatible != "" {
		if comp, err := strconv.Atoi(compatible); err == nil {
			req.Compatible = comp
			rawSignData["compatible"] = comp
		}
	}

	if test := ctx.PostForm("test"); test != "" {
		req.Test = (test == "true" || test == "1")
		rawSignData["test"] = req.Test
	}

	if sign := ctx.PostForm("sign"); sign != "" {
		req.Sign = sign
		rawSignData["sign"] = sign
	}

	return rawSignData
}

// GetOrder 获取订单信息
// @Summary 获取订单信息
// @Description 根据订单号获取订单详情
// @Tags 订单
// @Produce json
// @Param order_no path string true "订单号" example:"PAY20240101120000001"
// @Success 200 {object} response.Response{data=object} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "订单不存在"
// @Router /api/v1/orders/{order_no} [get]
func (c *OrderController) GetOrder(ctx *gin.Context) {
	orderNo := ctx.Param("order_no")
	if orderNo == "" {
		response.Fail(ctx, http.StatusBadRequest, "订单号不能为空")
		return
	}

	order, err := c.orderService.GetOrderByOrderNo(orderNo)
	if err != nil {
		response.Fail(ctx, http.StatusNotFound, err.Error())
		return
	}

	response.Success(ctx, order)
}

// QueryOrder 查询订单（根据商户订单号）
// @Summary 查询订单
// @Description 根据商户订单号查询订单
// @Tags 订单
// @Produce json
// @Param out_order_no query string true "商户订单号" example:"ORD20240101001"
// @Success 200 {object} response.Response{data=object} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "订单不存在"
// @Router /api/v1/orders/query [get]
func (c *OrderController) QueryOrder(ctx *gin.Context) {
	outOrderNo := ctx.Query("out_order_no")
	if outOrderNo == "" {
		response.Fail(ctx, http.StatusBadRequest, "商户订单号不能为空")
		return
	}

	// 获取商户ID
	merchantID, exists := ctx.Get("merchant_id")
	if !exists {
		response.Fail(ctx, http.StatusUnauthorized, "未认证")
		return
	}

	var id int64
	if idStr, ok := merchantID.(int64); ok {
		id = idStr
	} else if idStr, ok := merchantID.(string); ok {
		var err error
		id, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(ctx, http.StatusBadRequest, "商户ID格式错误")
			return
		}
	} else {
		response.Fail(ctx, http.StatusBadRequest, "无效的商户ID")
		return
	}

	order, err := c.orderService.GetOrderByOutOrderNo(outOrderNo, id)
	if err != nil {
		response.Fail(ctx, http.StatusNotFound, err.Error())
		return
	}

	response.Success(ctx, order)
}

// CreateOrderRequest 创建订单请求（用于文档）
type CreateOrderRequest struct {
	OutOrderNo   string `json:"out_order_no" binding:"required" example:"ORD20240101001"`
	Money        int    `json:"money" binding:"required,min=1" example:"10000"`
	Tax          int    `json:"tax" example:"100"`
	ProductName  string `json:"product_name" example:"支付宝扫码"`
	ReqExtra     string `json:"req_extra" example:"{}"`
	NotifyURL    string `json:"notify_url" binding:"required" example:"https://example.com/notify"`
	JumpURL      string `json:"jump_url" example:"https://example.com/jump"`
	ProductID    string `json:"product_id" example:"PROD001"`
	NotifyMoney  int    `json:"notify_money" binding:"required" example:"10000"`
	MerchantTax  int    `json:"merchant_tax" example:"50"`
	Extra        string `json:"extra" example:"{}"`
	PayChannelID *int64 `json:"pay_channel_id" example:"1"`
}
