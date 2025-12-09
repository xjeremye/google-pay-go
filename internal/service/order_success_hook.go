package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrderSuccessHookService 订单成功钩子服务
// 参考 Python: notify_order_success 和 @order_success_handle() 装饰器
type OrderSuccessHookService struct {
	pluginManager *plugin.Manager
}

// NewOrderSuccessHookService 创建订单成功钩子服务
func NewOrderSuccessHookService() *OrderSuccessHookService {
	pluginMgr := plugin.NewManager(database.RDB)
	return &OrderSuccessHookService{
		pluginManager: pluginMgr,
	}
}

// OrderSuccessData 订单成功数据
// 参考 Python: create_create_args 返回的数据结构
type OrderSuccessData struct {
	OrderNo        string
	OutOrderNo     string
	Tax            int // 租户手续费
	MerchantTax    int // 商户手续费
	Money          int // 订单金额
	NotifyMoney    int // 通知金额
	RealMoney      int // 实际收入 = notify_money - merchant_tax
	TenantID       int64
	MerchantID     int64
	WriteoffID     *int64
	ChannelID      int64
	PluginID       int64
	ProductID      string
	CreateDatetime time.Time
	PayDatetime    time.Time
	OrderID        string
	OrderBefore    int // 订单之前的状态
}

// NotifyOrderSuccess 触发订单成功钩子
// 参考 Python: notify_order_success(**create_data)
// 这会触发所有注册的回调函数
func (s *OrderSuccessHookService) NotifyOrderSuccess(ctx context.Context, data *OrderSuccessData) error {
	// 1. 调用插件的 callback_success
	if err := s.callbackPluginSuccess(ctx, data); err != nil {
		logger.Logger.Error("插件 callback_success 触发错误",
			zap.String("order_no", data.OrderNo),
			zap.Error(err))
		// 不返回错误，继续执行其他回调
	}

	// 2. 触发统计回调
	s.callbackStatistics(ctx, data)

	return nil
}

// callbackPluginSuccess 调用插件的 callback_success
func (s *OrderSuccessHookService) callbackPluginSuccess(ctx context.Context, data *OrderSuccessData) error {
	// 构建插件上下文（简化版，只包含必要信息）
	pluginCtx := &simpleOrderContextForSuccess{
		pluginID:   data.PluginID,
		pluginType: "", // 需要从订单详情获取
		channelID:  data.ChannelID,
	}

	// 查询订单详情以获取 plugin_type
	var orderDetail models.OrderDetail
	if err := database.DB.Where("order_id = ?", data.OrderID).First(&orderDetail).Error; err != nil {
		return fmt.Errorf("查询订单详情失败: %w", err)
	}
	pluginCtx.pluginType = orderDetail.PluginType

	// 获取插件实例
	pluginInstance, err := s.pluginManager.GetPluginByCtx(ctx, pluginCtx)
	if err != nil {
		return fmt.Errorf("获取插件实例失败: %w", err)
	}

	// 构建回调请求
	callbackReq := &plugin.CallbackSuccessRequest{
		OrderNo:        data.OrderNo,
		OutOrderNo:     data.OutOrderNo,
		PluginID:       data.PluginID,
		Tax:            data.Tax,
		MerchantTax:    data.MerchantTax,
		PluginType:     orderDetail.PluginType,
		Money:          data.Money,
		NotifyMoney:    data.NotifyMoney,
		RealMoney:      data.RealMoney,
		OrderID:        data.OrderID,
		ProductID:      data.ProductID,
		CookieID:       orderDetail.CookieID,
		ChannelID:      data.ChannelID,
		MerchantID:     data.MerchantID,
		WriteoffID:     data.WriteoffID,
		TenantID:       data.TenantID,
		CreateDatetime: data.CreateDatetime.Format("2006-01-02 15:04:05"),
		PayDatetime:    data.PayDatetime.Format("2006-01-02 15:04:05"),
		NotifyURL:      orderDetail.NotifyURL,
		PluginUpstream: orderDetail.PluginUpstream,
	}

	// 调用插件的 callback_success 方法
	return pluginInstance.CallbackSuccess(ctx, callbackReq)
}

// callbackStatistics 触发统计回调
// 参考 Python: 各种 @order_success_handle() 装饰的回调函数
func (s *OrderSuccessHookService) callbackStatistics(ctx context.Context, data *OrderSuccessData) {
	// 1. 通道统计
	s.callbackPayChannelSuccess(ctx, data)

	// 2. 商户统计
	s.callbackMerchantSuccess(ctx, data)

	// 3. 租户统计
	s.callbackTenantSuccess(ctx, data)

	// 4. 租户扣费
	s.callbackTenantTaxSuccess(ctx, data)

	// 5. 核销统计
	if data.WriteoffID != nil {
		s.callbackWriteoffSuccess(ctx, data)
	}

	// 6. 全局统计
	s.callbackDaySuccess(ctx, data)
}

// simpleOrderContextForSuccess 简单的订单上下文（用于获取插件）
type simpleOrderContextForSuccess struct {
	pluginID   int64
	pluginType string
	channelID  int64
}

func (o *simpleOrderContextForSuccess) GetOutOrderNo() string   { return "" }
func (o *simpleOrderContextForSuccess) GetNotifyURL() string    { return "" }
func (o *simpleOrderContextForSuccess) GetMoney() int           { return 0 }
func (o *simpleOrderContextForSuccess) GetJumpURL() string      { return "" }
func (o *simpleOrderContextForSuccess) GetNotifyMoney() int     { return 0 }
func (o *simpleOrderContextForSuccess) GetExtra() string        { return "" }
func (o *simpleOrderContextForSuccess) GetCompatible() int      { return 0 }
func (o *simpleOrderContextForSuccess) GetTest() bool           { return false }
func (o *simpleOrderContextForSuccess) GetMerchantID() int64    { return 0 }
func (o *simpleOrderContextForSuccess) GetTenantID() int64      { return 0 }
func (o *simpleOrderContextForSuccess) GetChannelID() int64     { return o.channelID }
func (o *simpleOrderContextForSuccess) GetPluginID() int64      { return o.pluginID }
func (o *simpleOrderContextForSuccess) GetPluginType() string   { return o.pluginType }
func (o *simpleOrderContextForSuccess) GetPluginUpstream() int  { return 0 }
func (o *simpleOrderContextForSuccess) GetDomainID() *int64     { return nil }
func (o *simpleOrderContextForSuccess) GetDomainURL() string    { return "" }
func (o *simpleOrderContextForSuccess) GetOrderNo() string      { return "" }
func (o *simpleOrderContextForSuccess) SetOrderNo(no string)    {}
func (o *simpleOrderContextForSuccess) SetDomainID(id int64)    {}
func (o *simpleOrderContextForSuccess) SetDomainURL(url string) {}

// callbackPayChannelSuccess 通道统计回调
// 参考 Python: callback_pay_channel_success
func (s *OrderSuccessHookService) callbackPayChannelSuccess(ctx context.Context, data *OrderSuccessData) {
	statsService := NewStatisticsService()
	stats := &models.PayChannelDayStatistics{
		PayChannelID: &data.ChannelID,
		TenantID:     &data.TenantID,
		MerchantID:   &data.MerchantID,
		WriteoffID:   data.WriteoffID,
	}

	updateFields := map[string]interface{}{
		"real_money": gorm.Expr("real_money + ?", data.RealMoney),
	}

	if err := statsService.SuccessBaseDayStatistics(ctx, stats, int64(data.NotifyMoney), int64(data.Tax), data.CreateDatetime, updateFields); err != nil {
		logger.Logger.Error("通道统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Error(err))
	}
}

// callbackMerchantSuccess 商户统计回调
// 参考 Python: callback_merchant_success
func (s *OrderSuccessHookService) callbackMerchantSuccess(ctx context.Context, data *OrderSuccessData) {
	statsService := NewStatisticsService()
	stats := &models.MerchantDayStatistics{
		MerchantID: &data.MerchantID,
	}

	updateFields := map[string]interface{}{
		"real_money": gorm.Expr("real_money + ?", data.RealMoney),
	}

	if err := statsService.SuccessBaseDayStatistics(ctx, stats, int64(data.NotifyMoney), int64(data.MerchantTax), data.CreateDatetime, updateFields); err != nil {
		logger.Logger.Error("商户统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Error(err))
		return
	}

	// 更新商户预付款（扣除实际收入）
	// 参考 Python: update_merchant_pre(-real_money, merchant_id)
	// 注意：这里需要实现商户预付款更新逻辑，暂时先记录日志
	logger.Logger.Info("商户统计成功，需要更新商户预付款",
		zap.String("order_no", data.OrderNo),
		zap.Int64("merchant_id", data.MerchantID),
		zap.Int("real_money", data.RealMoney))
	// TODO: 实现商户预付款更新
}

// callbackTenantSuccess 租户统计回调
// 参考 Python: callback_tenant_success
func (s *OrderSuccessHookService) callbackTenantSuccess(ctx context.Context, data *OrderSuccessData) {
	statsService := NewStatisticsService()
	stats := &models.TenantDayStatistics{
		TenantID: &data.TenantID,
	}

	if err := statsService.SuccessBaseDayStatistics(ctx, stats, int64(data.NotifyMoney), int64(data.Tax), data.CreateDatetime, nil); err != nil {
		logger.Logger.Error("租户统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Error(err))
	}
}

// callbackTenantTaxSuccess 租户扣费回调
// 参考 Python: callback_tenant_tax_success
func (s *OrderSuccessHookService) callbackTenantTaxSuccess(ctx context.Context, data *OrderSuccessData) {
	// 如果订单之前的状态不是7（已关闭），删除预扣
	if data.OrderBefore != 7 {
		// TODO: 实现删除预扣逻辑（take_up_tax）
		logger.Logger.Info("删除租户预扣",
			zap.String("order_no", data.OrderNo),
			zap.Int64("tenant_id", data.TenantID),
			zap.Int("tax", data.Tax))
	}

	// 扣除手续费
	// 参考 Python: tenant_success_tax(tenant_id, tax, order_id, channel_id, reorder)
	tenantService := NewTenantService()
	if err := tenantService.DeductTax(ctx, data.TenantID, data.Tax, data.OrderID, data.ChannelID); err != nil {
		logger.Logger.Error("租户扣费失败",
			zap.String("order_no", data.OrderNo),
			zap.Int64("tenant_id", data.TenantID),
			zap.Int("tax", data.Tax),
			zap.Error(err))
	}
}

// callbackWriteoffSuccess 核销统计回调
// 参考 Python: callback_writeoff_success
func (s *OrderSuccessHookService) callbackWriteoffSuccess(ctx context.Context, data *OrderSuccessData) {
	if data.WriteoffID == nil {
		return
	}

	statsService := NewStatisticsService()
	stats := &models.WriteOffDayStatistics{
		WriteoffID: data.WriteoffID,
	}

	if err := statsService.SuccessBaseDayStatistics(ctx, stats, int64(data.Money), int64(data.Tax), data.CreateDatetime, nil); err != nil {
		logger.Logger.Error("核销统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Error(err))
		return
	}

	// 更新核销预付款
	// 参考 Python: update_writeoff_pre(-money, writeoff_id)
	logger.Logger.Info("核销统计成功，需要更新核销预付款",
		zap.String("order_no", data.OrderNo),
		zap.Int64("writeoff_id", *data.WriteoffID),
		zap.Int("money", data.Money))
	// TODO: 实现核销预付款更新
}

// callbackDaySuccess 全局日统计回调
// 参考 Python: callback_day_success
func (s *OrderSuccessHookService) callbackDaySuccess(ctx context.Context, data *OrderSuccessData) {
	statsService := NewStatisticsService()
	stats := &models.DayStatistics{}

	if err := statsService.SuccessBaseDayStatistics(ctx, stats, int64(data.NotifyMoney), int64(data.Tax), data.CreateDatetime, nil); err != nil {
		logger.Logger.Error("全局日统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Error(err))
	}
}
