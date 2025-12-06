package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"github.com/golang-pay-core/internal/utils"
	"gorm.io/gorm"
)

// isValidJSON 检查字符串是否为有效的 JSON
func isValidJSON(s string) bool {
	if s == "" {
		return false
	}
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	OutOrderNo  string                 `json:"mchOrderNo" binding:"required"`   // 商户订单号
	MerchantID  int                    `json:"mchId" binding:"required"`        // 商户ID
	ChannelID   int                    `json:"channelId" binding:"required"`    // 渠道ID
	Money       int                    `json:"amount" binding:"required,min=1"` // 金额（分）
	NotifyURL   string                 `json:"notifyUrl" binding:"required"`    // 通知地址
	JumpURL     string                 `json:"jumpUrl"`                         // 跳转地址
	Extra       string                 `json:"extra"`                           // 额外参数
	Compatible  int                    `json:"compatible"`                      // 兼容模式 0/1
	Test        bool                   `json:"test"`                            // 测试模式
	Sign        string                 `json:"sign" binding:"required"`         // 签名
	RawSignData map[string]interface{} `json:"-"`                               // 原始签名数据（内部使用）
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	// Compatible == 1 时的响应格式
	TradeNo string `json:"trade_no,omitempty"`
	PayURL  string `json:"payurl,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Code    int    `json:"code,omitempty"`

	// Compatible == 0 时的响应格式
	MchOrderNo string `json:"mchOrderNo,omitempty"`
	PayOrderID string `json:"payOrderId,omitempty"`
	PayURL2    string `json:"payUrl,omitempty"`
	Sign       string `json:"sign,omitempty"`
}

// OrderCreateContext 订单创建上下文
type OrderCreateContext struct {
	// 请求参数
	OutOrderNo  string
	NotifyURL   string
	Money       int
	JumpURL     string
	NotifyMoney int
	Extra       string
	Compatible  int
	Test        bool

	// 验证后填充的字段
	MerchantID     int64
	TenantID       int64
	ChannelID      int64
	PluginID       int64
	PluginType     string
	PluginUpstream int
	DomainID       *int64
	DomainURL      string
	Tax            int
	SignKey        string

	// 关联对象
	Merchant   *models.Merchant
	Tenant     *models.Tenant
	Channel    *models.PayChannel
	Plugin     *models.PayPlugin
	PayType    *models.PayType
	User       *SystemUser
	TenantUser *SystemUser

	// 订单信息
	OrderNo string
	OrderID string
}

// 实现 plugin.OrderContext 接口
func (o *OrderCreateContext) GetOutOrderNo() string   { return o.OutOrderNo }
func (o *OrderCreateContext) GetNotifyURL() string    { return o.NotifyURL }
func (o *OrderCreateContext) GetMoney() int           { return o.Money }
func (o *OrderCreateContext) GetJumpURL() string      { return o.JumpURL }
func (o *OrderCreateContext) GetNotifyMoney() int     { return o.NotifyMoney }
func (o *OrderCreateContext) GetExtra() string        { return o.Extra }
func (o *OrderCreateContext) GetCompatible() int      { return o.Compatible }
func (o *OrderCreateContext) GetTest() bool           { return o.Test }
func (o *OrderCreateContext) GetMerchantID() int64    { return o.MerchantID }
func (o *OrderCreateContext) GetTenantID() int64      { return o.TenantID }
func (o *OrderCreateContext) GetChannelID() int64     { return o.ChannelID }
func (o *OrderCreateContext) GetPluginID() int64      { return o.PluginID }
func (o *OrderCreateContext) GetPluginType() string   { return o.PluginType }
func (o *OrderCreateContext) GetPluginUpstream() int  { return o.PluginUpstream }
func (o *OrderCreateContext) GetDomainID() *int64     { return o.DomainID }
func (o *OrderCreateContext) GetDomainURL() string    { return o.DomainURL }
func (o *OrderCreateContext) GetOrderNo() string      { return o.OrderNo }
func (o *OrderCreateContext) SetOrderNo(no string)    { o.OrderNo = no }
func (o *OrderCreateContext) SetDomainID(id int64)    { o.DomainID = &id }
func (o *OrderCreateContext) SetDomainURL(url string) { o.DomainURL = url }

// OrderService 订单服务（重构版）
type OrderService struct {
	cacheService  *CacheService
	pluginService *PluginService
	pluginManager *plugin.Manager
	redis         *redis.Client
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	pluginMgr := plugin.NewManager(database.RDB)
	pluginSvc := NewPluginService()

	// 设置插件信息提供者（实现 PluginInfoProvider 接口）
	pluginMgr.SetInfoProvider(&pluginInfoProviderAdapter{service: pluginSvc})

	return &OrderService{
		cacheService:  NewCacheService(),
		pluginService: pluginSvc,
		pluginManager: pluginMgr,
		redis:         database.RDB,
	}
}

// pluginInfoProviderAdapter 适配器，将 PluginService 适配为 PluginInfoProvider
type pluginInfoProviderAdapter struct {
	service *PluginService
}

func (a *pluginInfoProviderAdapter) GetPlugin(ctx context.Context, pluginID int64) (interface{}, error) {
	return a.service.GetPlugin(ctx, pluginID)
}

func (a *pluginInfoProviderAdapter) GetPluginUpstream(ctx context.Context, pluginID int64) (int, error) {
	return a.service.GetPluginUpstream(ctx, pluginID)
}

func (a *pluginInfoProviderAdapter) GetPluginPayTypes(ctx context.Context, pluginID int64) ([]interface{}, error) {
	payTypes, err := a.service.GetPluginPayTypes(ctx, pluginID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(payTypes))
	for i, pt := range payTypes {
		result[i] = map[string]interface{}{
			"id":     pt.ID,
			"name":   pt.Name,
			"key":    pt.Key,
			"status": pt.Status,
		}
	}
	return result, nil
}

// CreateOrder 创建订单（主入口）
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, *OrderError) {
	startTime := time.Now()

	// 1. 基础验证
	if req.Money <= 0 {
		return nil, ErrAmountInvalid
	}

	// 2. 创建上下文
	orderCtx := &OrderCreateContext{
		OutOrderNo:  req.OutOrderNo,
		NotifyURL:   req.NotifyURL,
		Money:       req.Money,
		JumpURL:     req.JumpURL,
		NotifyMoney: req.Money,
		Extra:       req.Extra,
		Compatible:  req.Compatible,
		Test:        req.Test,
	}

	// 3. 执行验证链
	if err := s.validateMerchant(ctx, orderCtx, int64(req.MerchantID)); err != nil {
		return nil, err
	}
	if err := s.validateTenant(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.validateSign(ctx, orderCtx, req.RawSignData); err != nil {
		return nil, err
	}
	if err := s.validateOutOrderNo(ctx, orderCtx); err != nil {
		return nil, err
	}
	if err := s.validateChannel(ctx, orderCtx, int64(req.ChannelID)); err != nil {
		return nil, err
	}
	if err := s.validatePlugin(ctx, orderCtx); err != nil {
		return nil, err
	}

	// 4. 创建订单和详情
	if err := s.createOrderAndDetail(ctx, orderCtx); err != nil {
		return nil, err
	}

	// 5. 检查余额（简化版，实际需要查询租户余额表）
	// TODO: 实现余额检查

	// 6. 生成支付URL（使用插件系统）
	payURL, err := s.generatePayURL(ctx, orderCtx)
	if err != nil {
		return nil, err
	}

	// 7. 设置缓存
	s.setCache(ctx, orderCtx)

	// 8. 构建响应
	response := s.buildResponse(orderCtx, payURL)

	// 记录耗时
	elapsed := time.Since(startTime)
	if elapsed > 1*time.Second {
		// TODO: 记录慢查询日志
		fmt.Printf("[Order] Create order %s took %v\n", orderCtx.OrderNo, elapsed)
	}

	return response, nil
}

// validateMerchant 验证商户
func (s *OrderService) validateMerchant(ctx context.Context, orderCtx *OrderCreateContext, merchantID int64) *OrderError {
	merchant, user, err := s.cacheService.GetMerchantWithUser(ctx, merchantID)
	if err != nil {
		return ErrMerchantNotFound
	}

	if user == nil {
		return ErrMerchantDisabled
	}

	if !user.Status {
		return ErrMerchantDisabled
	}

	orderCtx.Merchant = merchant
	orderCtx.MerchantID = merchantID
	orderCtx.User = user
	orderCtx.SignKey = user.Key

	return nil
}

// validateTenant 验证租户
func (s *OrderService) validateTenant(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	if orderCtx.Merchant == nil {
		return ErrMerchantNotFound
	}

	tenant, tenantUser, err := s.cacheService.GetTenantWithUser(ctx, orderCtx.Merchant.ParentID)
	if err != nil {
		return NewOrderError(ErrCodeMerchantDisabled, "商户上级已被禁用,请联系管理员")
	}

	if tenantUser == nil || !tenantUser.Status {
		return NewOrderError(ErrCodeMerchantDisabled, "商户上级已被禁用,请联系管理员")
	}

	orderCtx.Tenant = tenant
	orderCtx.TenantID = tenant.ID
	orderCtx.TenantUser = tenantUser

	return nil
}

// validateSign 验证签名
func (s *OrderService) validateSign(ctx context.Context, orderCtx *OrderCreateContext, rawSignData map[string]interface{}) *OrderError {
	if orderCtx.SignKey == "" {
		return ErrSignInvalid
	}

	sign, ok := rawSignData["sign"].(string)
	if !ok || sign == "" {
		return ErrSignInvalid
	}

	_, actualSign := utils.GetSign(rawSignData, orderCtx.SignKey, nil, nil, orderCtx.Compatible)
	if sign != actualSign {
		// 如果是开发、测试模式则输出正确签名
		if config.Cfg.App.Mode == "debug" || config.Cfg.App.Mode == "test" {
			fmt.Println("sign", sign)
			fmt.Println("actualSign", actualSign)
		}
		return ErrSignInvalid
	}

	return nil
}

// validateOutOrderNo 验证商户订单号（使用 Redis SetNX 做幂等控制）
func (s *OrderService) validateOutOrderNo(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	if orderCtx.OutOrderNo == "" {
		return ErrOutOrderNoRequired
	}

	key := fmt.Sprintf("out_order_no:%s", orderCtx.OutOrderNo)
	ok, err := s.redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
	if err != nil {
		return ErrSystemBusy
	}

	if !ok {
		return ErrOutOrderNoExists
	}

	return nil
}

// validateChannel 验证渠道
func (s *OrderService) validateChannel(ctx context.Context, orderCtx *OrderCreateContext, channelID int64) *OrderError {
	channel, err := s.cacheService.GetPayChannel(ctx, channelID)
	if err != nil {
		return ErrChannelNotFound
	}

	if !channel.Status {
		return ErrChannelDisabled
	}

	// 检查时间范围
	if err := s.checkChannelTime(channel); err != nil {
		return err
	}

	// 检查金额范围
	if err := s.checkChannelAmount(channel, orderCtx); err != nil {
		return err
	}

	// 应用浮动加价
	s.applyFloatAmount(channel, orderCtx)

	orderCtx.Channel = channel
	orderCtx.ChannelID = channelID
	orderCtx.PluginID = channel.PluginID

	return nil
}

// validatePlugin 验证插件
func (s *OrderService) validatePlugin(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	if orderCtx.PluginID == 0 {
		return ErrPluginUnavailable
	}

	// 获取插件信息
	pluginInfo, err := s.pluginService.GetPlugin(ctx, orderCtx.PluginID)
	if err != nil {
		return ErrPluginUnavailable
	}

	if !pluginInfo.Status {
		return ErrPluginUnavailable
	}

	orderCtx.Plugin = pluginInfo

	// 获取插件上游类型
	upstream, err := s.pluginService.GetPluginUpstream(ctx, orderCtx.PluginID)
	if err != nil {
		return ErrPluginUnavailable
	}
	orderCtx.PluginUpstream = upstream

	// 获取支付类型（关键：获取 PayType 的 key，如 alipay_wap）
	payTypes, err := s.pluginService.GetPluginPayTypes(ctx, orderCtx.PluginID)
	if err != nil || len(payTypes) == 0 {
		return NewOrderError(ErrCodePayTypeUnavailable, "该通道不可用")
	}

	// 使用第一个支付类型
	payType := payTypes[0]
	if !payType.Status {
		return NewOrderError(ErrCodePayTypeUnavailable, "该通道不可用")
	}

	orderCtx.PayType = &payType
	// 关键：使用 PayType 的 Key 作为 PluginType（如 alipay_wap）
	// 插件管理器将使用这个 key 来创建对应的插件实例
	orderCtx.PluginType = payType.Key

	if orderCtx.PluginType == "" {
		return NewOrderError(ErrCodePayTypeUnavailable, "支付类型标识不能为空")
	}

	return nil
}

// checkChannelTime 检查渠道可用时间
func (s *OrderService) checkChannelTime(channel *models.PayChannel) *OrderError {
	if channel.StartTime == "00:00:00" && channel.EndTime == "00:00:00" {
		return nil // 全天可用
	}

	layout := "15:04:05"
	startTime, err1 := time.Parse(layout, channel.StartTime)
	endTime, err2 := time.Parse(layout, channel.EndTime)
	if err1 != nil || err2 != nil {
		return nil // 时间格式错误，跳过检查
	}

	now := time.Now()
	currentTime, _ := time.Parse(layout, now.Format(layout))

	// 支持跨零点时段
	if startTime.Before(endTime) {
		if currentTime.Before(startTime) || currentTime.After(endTime) {
			return NewOrderError(ErrCodeChannelTimeInvalid,
				fmt.Sprintf("通道不在可使用时间[%s-%s]", channel.StartTime, channel.EndTime))
		}
	} else if startTime.After(endTime) { // 跨0点情况
		if currentTime.Before(startTime) && currentTime.After(endTime) {
			return NewOrderError(ErrCodeChannelTimeInvalid,
				fmt.Sprintf("通道不在可使用时间[%s-%s]", channel.StartTime, channel.EndTime))
		}
	}

	return nil
}

// checkChannelAmount 检查渠道金额限制
func (s *OrderService) checkChannelAmount(channel *models.PayChannel, orderCtx *OrderCreateContext) *OrderError {
	// 检查固定金额模式
	if channel.Settled && channel.Moneys != "" {
		var moneys []int
		if err := json.Unmarshal([]byte(channel.Moneys), &moneys); err == nil {
			valid := false
			for _, m := range moneys {
				if m == orderCtx.Money {
					valid = true
					break
				}
			}
			if !valid {
				return NewOrderError(ErrCodeAmountOutOfRange,
					fmt.Sprintf("金额%d不在范围内,可用:%v", orderCtx.Money, moneys))
			}
		}
	}

	// 检查金额范围
	if channel.MinMoney > 0 || channel.MaxMoney > 0 {
		if orderCtx.Money < channel.MinMoney || orderCtx.Money > channel.MaxMoney {
			return NewOrderError(ErrCodeAmountOutOfRange,
				fmt.Sprintf("金额%d不在范围[%d,%d]内", orderCtx.Money, channel.MinMoney, channel.MaxMoney))
		}
	}

	return nil
}

// applyFloatAmount 应用浮动加价
func (s *OrderService) applyFloatAmount(channel *models.PayChannel, orderCtx *OrderCreateContext) {
	if channel.FloatMinMoney > 0 || channel.FloatMaxMoney > 0 {
		delta := channel.FloatMinMoney
		if channel.FloatMaxMoney > channel.FloatMinMoney {
			delta = channel.FloatMinMoney + rand.Intn(channel.FloatMaxMoney-channel.FloatMinMoney+1)
		}
		orderCtx.Money += delta
	}

	if orderCtx.Money <= 0 {
		// 如果金额变为0或负数，恢复原金额
		orderCtx.Money = orderCtx.NotifyMoney
	}
}

// createOrderAndDetail 创建订单和订单详情
func (s *OrderService) createOrderAndDetail(ctx context.Context, orderCtx *OrderCreateContext) *OrderError {
	// 生成订单号
	orderCtx.OrderNo = utils.GenerateOrderNo()
	orderCtx.OrderID = utils.GenerateID()

	now := time.Now()

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建订单
	order := &models.Order{
		ID:             orderCtx.OrderID,
		OrderNo:        orderCtx.OrderNo,
		OutOrderNo:     orderCtx.OutOrderNo,
		OrderStatus:    models.OrderStatusPending,
		Money:          orderCtx.Money,
		Tax:            orderCtx.Tax,
		CreateDatetime: &now,
		Compatible:     orderCtx.Compatible,
		Ver:            1,
		MerchantID:     &orderCtx.MerchantID,
		PayChannelID:   &orderCtx.ChannelID,
	}

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("创建订单失败: %v", err))
	}

	// 创建订单详情
	// 处理 Extra 字段：如果为空字符串，设置为 NULL 或有效的 JSON
	extraValue := orderCtx.Extra
	if extraValue == "" {
		// 空字符串时，设置为有效的空 JSON 对象
		extraValue = "{}"
	} else {
		// 验证是否为有效的 JSON，如果不是则包装为 JSON 字符串
		if !isValidJSON(extraValue) {
			// 如果不是有效的 JSON，将其作为字符串值包装在 JSON 中
			extraValue = fmt.Sprintf(`{"value":%q}`, extraValue)
		}
	}

	orderDetail := &models.OrderDetail{
		OrderID:        orderCtx.OrderID,
		NotifyURL:      orderCtx.NotifyURL,
		JumpURL:        orderCtx.JumpURL,
		NotifyMoney:    orderCtx.NotifyMoney,
		CreateDatetime: &now,
		Extra:          extraValue,
		PluginID:       &orderCtx.PluginID,
		DomainID:       orderCtx.DomainID,
	}

	if err := tx.Create(orderDetail).Error; err != nil {
		tx.Rollback()
		return NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("创建订单详情失败: %v", err))
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("提交事务失败: %v", err))
	}

	return nil
}

// generatePayURL 生成支付URL（使用插件系统）
func (s *OrderService) generatePayURL(ctx context.Context, orderCtx *OrderCreateContext) (string, *OrderError) {
	// 获取插件实例
	pluginInstance, err := s.pluginManager.GetPluginByCtx(ctx, orderCtx)
	if err != nil {
		return "", NewOrderError(ErrCodePluginUnavailable, fmt.Sprintf("插件不可用: %v", err))
	}

	// 构建插件请求
	createReq := &plugin.CreateOrderRequest{
		OutOrderNo:     orderCtx.OutOrderNo,
		OrderNo:        orderCtx.OrderNo,
		Money:          orderCtx.Money,
		NotifyURL:      orderCtx.NotifyURL,
		JumpURL:        orderCtx.JumpURL,
		Extra:          orderCtx.Extra,
		MerchantID:     orderCtx.MerchantID,
		TenantID:       orderCtx.TenantID,
		ChannelID:      orderCtx.ChannelID,
		PluginID:       orderCtx.PluginID,
		PluginType:     orderCtx.PluginType,
		PluginUpstream: orderCtx.PluginUpstream,
		DomainID:       orderCtx.DomainID,
		DomainURL:      orderCtx.DomainURL,
		Compatible:     orderCtx.Compatible,
		Test:           orderCtx.Test,
	}

	// 添加关联对象（转换为 map）
	if orderCtx.Channel != nil {
		channelMap := map[string]interface{}{
			"id":        orderCtx.Channel.ID,
			"name":      orderCtx.Channel.Name,
			"status":    orderCtx.Channel.Status,
			"plugin_id": orderCtx.Channel.PluginID,
		}
		createReq.Channel = channelMap
	}

	if orderCtx.Plugin != nil {
		pluginMap := map[string]interface{}{
			"id":          orderCtx.Plugin.ID,
			"name":        orderCtx.Plugin.Name,
			"status":      orderCtx.Plugin.Status,
			"description": orderCtx.Plugin.Description,
		}
		createReq.Plugin = pluginMap
	}

	if orderCtx.PayType != nil {
		payTypeMap := map[string]interface{}{
			"id":     orderCtx.PayType.ID,
			"name":   orderCtx.PayType.Name,
			"key":    orderCtx.PayType.Key,
			"status": orderCtx.PayType.Status,
		}
		createReq.PayType = payTypeMap
	}

	// 调用插件创建订单
	createResp, err := pluginInstance.CreateOrder(ctx, createReq)
	if err != nil {
		return "", NewOrderError(ErrCodeCreateFailed, fmt.Sprintf("创建订单失败: %v", err))
	}

	// 检查响应
	if !createResp.IsSuccess() {
		return "", NewOrderError(createResp.ErrorCode, createResp.ErrorMessage)
	}

	return createResp.PayURL, nil
}

// setCache 设置缓存
func (s *OrderService) setCache(ctx context.Context, orderCtx *OrderCreateContext) {
	// 缓存订单号映射
	s.redis.Set(ctx, fmt.Sprintf("order_no:%s", orderCtx.OrderNo), orderCtx.OrderID, 24*time.Hour)
	s.redis.Set(ctx, fmt.Sprintf("out_order_no_map:%s", orderCtx.OutOrderNo), orderCtx.OrderNo, 24*time.Hour)
}

// buildResponse 构建响应
func (s *OrderService) buildResponse(orderCtx *OrderCreateContext, payURL string) *CreateOrderResponse {
	response := &CreateOrderResponse{}

	if orderCtx.Compatible == 1 {
		// 兼容模式
		response.TradeNo = orderCtx.OrderNo
		response.PayURL = payURL
		response.Msg = "订单创建成功"
		response.Code = 1
	} else {
		// 标准模式
		response.MchOrderNo = orderCtx.OutOrderNo
		response.PayOrderID = orderCtx.OrderNo
		response.PayURL2 = payURL

		// 生成响应签名
		dataMap := map[string]interface{}{
			"mchOrderNo": response.MchOrderNo,
			"payOrderId": response.PayOrderID,
			"payUrl":     response.PayURL2,
		}
		// 响应签名使用所有字段（useList 为 nil 表示使用所有字段）
		response.Sign = utils.GenerateResponseSign(dataMap, orderCtx.SignKey, orderCtx.Compatible)
	}

	return response
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *OrderService) GetOrderByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	err := database.DB.Preload("OrderDetail").
		Preload("Merchant").
		Preload("PayChannel").
		Where("order_no = ?", orderNo).
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	return &order, nil
}

// GetOrderByOutOrderNo 根据商户订单号获取订单
func (s *OrderService) GetOrderByOutOrderNo(outOrderNo string, merchantID int64) (*models.Order, error) {
	var order models.Order
	err := database.DB.Preload("OrderDetail").
		Where("out_order_no = ? AND merchant_id = ?", outOrderNo, merchantID).
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}

	return &order, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(orderID string, status int, ticketNo string) error {
	now := time.Now()

	// 先更新版本号
	if err := database.DB.Exec("UPDATE dvadmin_order SET ver = ver + 1 WHERE id = ?", orderID).Error; err != nil {
		return fmt.Errorf("更新版本号失败: %w", err)
	}

	updates := map[string]interface{}{
		"order_status":    status,
		"update_datetime": &now,
	}

	if status == models.OrderStatusPaid {
		updates["pay_datetime"] = &now
	}

	if ticketNo != "" {
		if err := database.DB.Model(&models.OrderDetail{}).
			Where("order_id = ?", orderID).
			Update("ticket_no", ticketNo).Error; err != nil {
			return fmt.Errorf("更新订单详情失败: %w", err)
		}
	}

	if err := database.DB.Model(&models.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	return nil
}
