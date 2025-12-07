package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/response"
	"github.com/golang-pay-core/internal/utils"
)

// PayController 支付控制器
type PayController struct{}

// NewPayController 创建支付控制器
func NewPayController() *PayController {
	return &PayController{}
}

// Auth 鉴权接口
// 参考 Python: 验证鉴权参数，验证通过后返回订单的支付URL
// @Summary 支付鉴权
// @Description 验证收银台的鉴权请求，验证通过后返回订单的支付URL
// @Tags 支付
// @Accept json
// @Produce json
// @Param order_no query string true "订单号" example:"PAY20240101120000001"
// @Param auth_key query string true "鉴权密钥" example:"auth_key_123"
// @Param timestamp query int true "时间戳" example:"1704067200"
// @Param pay_url query string false "支付URL（可选）" example:"https://..."
// @Param sign query string true "签名" example:"ABC123..."
// @Success 200 {object} response.Response{data=object} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "鉴权失败"
// @Router /api/pay/auth [get]
func (c *PayController) Auth(ctx *gin.Context) {
	// 获取参数
	orderNo := ctx.Query("order_no")
	authKey := ctx.Query("auth_key")
	timestampStr := ctx.Query("timestamp")
	payURL := ctx.Query("pay_url")
	sign := ctx.Query("sign")

	// 参数验证
	if orderNo == "" || authKey == "" || timestampStr == "" || sign == "" {
		response.Fail(ctx, http.StatusBadRequest, "参数不完整")
		return
	}

	// 解析时间戳
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		response.Fail(ctx, http.StatusBadRequest, "时间戳格式错误")
		return
	}

	// 验证时间戳（检查是否过期，默认5分钟有效期）
	currentTime := time.Now().Unix()
	if currentTime-timestamp > 300 { // 5分钟
		response.Fail(ctx, http.StatusUnauthorized, "鉴权已过期")
		return
	}

	// 构建鉴权参数
	authParams := map[string]interface{}{
		"order_no":  orderNo,
		"auth_key":  authKey,
		"timestamp": timestamp,
	}
	if payURL != "" {
		authParams["pay_url"] = payURL
	}

	// 验证签名
	signData := make(map[string]interface{})
	for k, v := range authParams {
		if k != "sign" {
			signData[k] = v
		}
	}

	// 生成签名（使用 utils 的签名方法）
	_, expectedSign := utils.GetSign(signData, authKey, nil, nil, 0)
	if sign != expectedSign {
		response.Fail(ctx, http.StatusUnauthorized, "签名验证失败")
		return
	}

	// 查询订单
	var order models.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		response.Fail(ctx, http.StatusNotFound, "订单不存在")
		return
	}

	// 查询订单详情（获取支付URL）
	var orderDetail models.OrderDetail
	if err := database.DB.Where("order_id = ?", order.ID).First(&orderDetail).Error; err != nil {
		response.Fail(ctx, http.StatusNotFound, "订单详情不存在")
		return
	}

	// 从订单详情的 Extra 字段中获取支付URL
	var payURLFromDB string
	if orderDetail.Extra != "" && orderDetail.Extra != "{}" {
		var extraMap map[string]interface{}
		if err := json.Unmarshal([]byte(orderDetail.Extra), &extraMap); err == nil {
			if url, ok := extraMap["pay_url"].(string); ok {
				payURLFromDB = url
			}
		}
	}

	// 如果 Extra 中没有支付URL，检查请求参数中的 pay_url
	if payURLFromDB == "" && payURL != "" {
		payURLFromDB = payURL
	}

	// 如果还是没有支付URL，返回错误
	if payURLFromDB == "" {
		response.Fail(ctx, http.StatusNotFound, "支付URL不存在")
		return
	}

	// 验证域名和鉴权密钥
	// 查询域名信息
	if orderDetail.DomainID == nil {
		response.Fail(ctx, http.StatusUnauthorized, "订单未关联域名")
		return
	}

	var domain models.PayDomain
	if err := database.DB.Where("id = ?", *orderDetail.DomainID).First(&domain).Error; err != nil {
		response.Fail(ctx, http.StatusUnauthorized, "域名不存在")
		return
	}

	// 使用 get_auth_key 方法生成预期的鉴权密钥
	// 参考 Python: get_auth_key(raw, p_key, offset=30)
	// raw 使用订单号，p_key 使用域名的密钥，offset 默认30秒
	expectedAuthKey := utils.GetAuthKey(orderNo, domain.AuthKey, 30)

	// 验证鉴权密钥（允许前后30秒的时间窗口）
	// 因为时间窗口可能变化，需要检查当前窗口和前一个窗口
	validAuthKey1 := expectedAuthKey
	// 检查前一个时间窗口（timestamp - 30）
	prevTimestamp := timestamp - 30
	prevTimeWindow := prevTimestamp / 30
	currentTimeWindow := timestamp / 30
	if prevTimeWindow != currentTimeWindow {
		// 如果时间窗口不同，也检查前一个窗口的密钥
		validAuthKey2 := utils.GetAuthKeyWithTimeWindow(orderNo, domain.AuthKey, prevTimeWindow)
		if authKey != validAuthKey1 && authKey != validAuthKey2 {
			response.Fail(ctx, http.StatusUnauthorized, "鉴权密钥错误")
			return
		}
	} else {
		// 时间窗口相同，只检查当前密钥
		if authKey != validAuthKey1 {
			response.Fail(ctx, http.StatusUnauthorized, "鉴权密钥错误")
			return
		}
	}

	// 验证鉴权状态
	if !domain.AuthStatus {
		response.Fail(ctx, http.StatusUnauthorized, "域名鉴权已禁用")
		return
	}

	// 验证时间戳过期时间（如果域名配置了 auth_timeout）
	if domain.AuthTimeout > 0 {
		expireTime := timestamp + int64(domain.AuthTimeout)
		if currentTime > expireTime {
			response.Fail(ctx, http.StatusUnauthorized, "鉴权已过期")
			return
		}
	}

	// 鉴权成功，返回支付URL
	response.Success(ctx, gin.H{
		"order_no": orderNo,
		"pay_url":  payURLFromDB,
	})
}

// Cashier 收银台页面
// 参考 Python: 收银台页面，显示订单信息并跳转到支付URL
// @Summary 收银台
// @Description 收银台页面，显示订单信息并自动跳转到支付URL
// @Tags 支付
// @Produce html
// @Param order_no query string true "订单号" example:"PAY20240101120000001"
// @Param expire_time query int false "过期时间戳" example:"1704067500"
// @Success 200 {string} string "HTML页面"
// @Failure 400 {string} string "参数错误"
// @Failure 404 {string} string "订单不存在"
// @Router /cashier [get]
func (c *PayController) Cashier(ctx *gin.Context) {
	// 获取参数
	orderNo := ctx.Query("order_no")
	expireTimeStr := ctx.Query("expire_time")

	if orderNo == "" {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "参数错误",
			"message": "订单号不能为空",
		})
		return
	}

	// 查询订单
	var order models.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "订单不存在",
			"message": "未找到订单：" + orderNo,
		})
		return
	}

	// 查询订单详情
	var orderDetail models.OrderDetail
	if err := database.DB.Where("order_id = ?", order.ID).First(&orderDetail).Error; err != nil {
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "订单详情不存在",
			"message": "未找到订单详情",
		})
		return
	}

	// 检查订单状态
	if order.OrderStatus != models.OrderStatusPending {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "订单状态错误",
			"message": "订单已处理，无法支付",
		})
		return
	}

	// 检查过期时间
	if expireTimeStr != "" {
		expireTime, err := strconv.ParseInt(expireTimeStr, 10, 64)
		if err == nil {
			currentTime := time.Now().Unix()
			if currentTime > expireTime {
				ctx.HTML(http.StatusBadRequest, "error.html", gin.H{
					"title":   "订单已过期",
					"message": "订单已过期，请重新下单",
				})
				return
			}
		}
	}

	// 查询域名信息（用于判断是否需要鉴权）
	var domain *models.PayDomain
	if orderDetail.DomainID != nil {
		var d models.PayDomain
		if err := database.DB.Where("id = ?", *orderDetail.DomainID).First(&d).Error; err == nil {
			domain = &d
		}
	}

	// 获取支付URL
	var payURL string
	if domain != nil && domain.AuthStatus && domain.AuthKey != "" {
		// 需要鉴权，通过鉴权逻辑获取支付URL
		// 参考 Python: 收银台调用鉴权接口获取支付URL
		payURL = c.getPayURLWithAuth(ctx, orderNo, domain, &orderDetail)
	} else {
		// 不需要鉴权，直接从订单详情获取支付URL
		payURL = c.getPayURLFromOrderDetail(&orderDetail)
	}

	if payURL == "" {
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "支付URL不存在",
			"message": "无法获取支付链接，请联系客服",
		})
		return
	}

	// 格式化金额（分转元）
	amount := float64(order.Money) / 100.0

	// 渲染收银台页面
	ctx.HTML(http.StatusOK, "cashier.html", gin.H{
		"title":      "收银台",
		"order_no":   orderNo,
		"amount":     amount,
		"pay_url":    payURL,
		"expireTime": expireTimeStr,
	})
}

// getPayURLWithAuth 通过鉴权逻辑获取支付URL
// 参考 Python: 收银台调用鉴权接口的逻辑
func (c *PayController) getPayURLWithAuth(ctx *gin.Context, orderNo string, domain *models.PayDomain, orderDetail *models.OrderDetail) string {
	// 先从订单详情获取支付URL（优先使用已存储的支付URL）
	payURL := c.getPayURLFromOrderDetail(orderDetail)
	if payURL != "" {
		return payURL
	}

	// 如果订单详情中没有支付URL，通过鉴权接口获取
	// 生成鉴权参数
	timestamp := time.Now().Unix()
	authKey := utils.GetAuthKey(orderNo, domain.AuthKey, 30)

	// 构建鉴权参数
	authParams := map[string]interface{}{
		"order_no":  orderNo,
		"auth_key":  authKey,
		"timestamp": timestamp,
	}

	// 生成签名
	signData := make(map[string]interface{})
	for k, v := range authParams {
		if k != "sign" {
			signData[k] = v
		}
	}
	_, sign := utils.GetSign(signData, authKey, nil, nil, 0)

	// 构建鉴权URL（使用当前请求的 Host）
	host := ctx.Request.Host
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	authURL := scheme + "://" + host + "/api/v1/pay/auth?order_no=" + orderNo + "&auth_key=" + authKey + "&timestamp=" + strconv.FormatInt(timestamp, 10) + "&sign=" + sign

	// 调用鉴权接口
	req, err := http.NewRequestWithContext(ctx.Request.Context(), "GET", authURL, nil)
	if err != nil {
		return ""
	}

	// 复制请求头
	for k, v := range ctx.Request.Header {
		req.Header[k] = v
	}

	// 发送请求
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// 解析响应
	if resp.StatusCode == http.StatusOK {
		var result struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				OrderNo string `json:"order_no"`
				PayURL  string `json:"pay_url"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.Code == 200 {
			return result.Data.PayURL
		}
	}

	return ""
}

// getPayURLFromOrderDetail 从订单详情获取支付URL
func (c *PayController) getPayURLFromOrderDetail(orderDetail *models.OrderDetail) string {
	if orderDetail.Extra == "" || orderDetail.Extra == "{}" {
		return ""
	}

	var extraMap map[string]interface{}
	if err := json.Unmarshal([]byte(orderDetail.Extra), &extraMap); err != nil {
		return ""
	}

	if payURL, ok := extraMap["pay_url"].(string); ok {
		return payURL
	}

	return ""
}
