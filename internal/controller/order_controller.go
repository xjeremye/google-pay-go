package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-pay-core/internal/response"
	"github.com/golang-pay-core/internal/service"
)

type OrderController struct {
	orderService   *service.OrderService
	merchantService *service.MerchantService
	payChannelService *service.PayChannelService
}

// NewOrderController 创建订单控制器
func NewOrderController() *OrderController {
	return &OrderController{
		orderService:      service.NewOrderService(),
		merchantService:    service.NewMerchantService(),
		payChannelService: service.NewPayChannelService(),
	}
}

// CreateOrder 创建订单
// @Summary 创建订单
// @Description 创建支付订单
// @Tags 订单
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "订单信息"
// @Success 200 {object} response.Response{data=object} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/orders [post]
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req service.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 获取商户ID（从中间件中获取）
	merchantID, exists := ctx.Get("merchant_id")
	if exists {
		if id, ok := merchantID.(int64); ok {
			req.MerchantID = &id
		}
	}

	// 验证商户
	if req.MerchantID != nil {
		if err := c.merchantService.ValidateMerchant(*req.MerchantID); err != nil {
			response.Fail(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}

	// 验证支付通道
	if req.PayChannelID != nil {
		if err := c.payChannelService.ValidatePayChannel(*req.PayChannelID, req.Money); err != nil {
			response.Fail(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}

	// 创建订单
	order, err := c.orderService.CreateOrder(&req)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, order)
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

