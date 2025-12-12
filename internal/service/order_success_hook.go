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

	// 5.1 核销通道统计（需要从订单详情中获取最终核销手续费和实际扣除金额）
	// 注意：订单成功和退款都使用同一个方法，但传入的参数不同
	if data.WriteoffID != nil && data.ChannelID > 0 {
		s.callbackWriteoffChannelSuccess(ctx, data)
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
	// 记录日志，帮助调试 tax 值
	logger.Logger.Info("通道统计回调",
		zap.String("order_no", data.OrderNo),
		zap.Int64("channel_id", data.ChannelID),
		zap.Int("tax", data.Tax),
		zap.Int("notify_money", data.NotifyMoney))

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
			zap.Int("tax", data.Tax),
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
	// 记录日志，帮助调试
	logger.Logger.Info("更新租户日统计",
		zap.String("order_no", data.OrderNo),
		zap.Int64("tenant_id", data.TenantID),
		zap.Int64("notify_money", int64(data.NotifyMoney)),
		zap.Int("tax", data.Tax))

	if data.TenantID == 0 {
		logger.Logger.Warn("租户ID为0，跳过租户统计",
			zap.String("order_no", data.OrderNo))
		return
	}

	statsService := NewStatisticsService()
	stats := &models.TenantDayStatistics{
		TenantID: &data.TenantID,
	}

	if err := statsService.SuccessBaseDayStatistics(ctx, stats, int64(data.NotifyMoney), int64(data.Tax), data.CreateDatetime, nil); err != nil {
		logger.Logger.Error("租户统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Int64("tenant_id", data.TenantID),
			zap.Int("tax", data.Tax),
			zap.Error(err))
	} else {
		logger.Logger.Info("租户统计更新成功",
			zap.String("order_no", data.OrderNo),
			zap.Int64("tenant_id", data.TenantID),
			zap.Int("tax", data.Tax),
			zap.Int64("notify_money", int64(data.NotifyMoney)))
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

	// 记录日志，帮助调试 tax 值
	logger.Logger.Info("核销统计回调",
		zap.String("order_no", data.OrderNo),
		zap.Int64("writeoff_id", *data.WriteoffID),
		zap.Int("money", data.Money),
		zap.Int("tax", data.Tax),
		zap.Int("notify_money", data.NotifyMoney))

	statsService := NewStatisticsService()
	stats := &models.WriteOffDayStatistics{
		WriteoffID: data.WriteoffID,
	}

	if err := statsService.SuccessBaseDayStatistics(ctx, stats, int64(data.Money), int64(data.Tax), data.CreateDatetime, nil); err != nil {
		logger.Logger.Error("核销统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Int("tax", data.Tax),
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

// callbackWriteoffChannelSuccess 核销通道统计回调
// 参考 Python: callback_writeoff_channel_tax_success (订单成功) 和 callback_writeoff_tax_refund (订单退款)
// 根据文档：
// - 订单成功时：success_money += real_money, total_tax += parent_tax_money
// - 订单退款时：success_money -= flow.money, total_tax -= flow.tax（使用负数）
func (s *OrderSuccessHookService) callbackWriteoffChannelSuccess(ctx context.Context, data *OrderSuccessData) {
	if data.WriteoffID == nil || data.ChannelID == 0 {
		return
	}

	// 查询订单详情以获取最终核销手续费和实际扣除金额
	// 这些值在 status_updater.go 中已经计算并记录到 WriteoffCashflow 中
	// 订单成功时：flow_type=1（跑量流水）
	// 订单退款时：需要查找订单成功时的流水记录（flow_type=1）
	var cashflow models.WriteoffCashflow
	flowType := models.WriteoffCashflowTypeRunVolume // 1=跑量流水

	// 根据文档：订单退款时，只有 order_before in [4, 6] 的订单才能退款（成功订单）
	// 订单成功时：直接查找跑量流水
	// 订单退款时：也需要查找跑量流水（订单成功时的流水记录）
	if err := database.DB.Where("order_id = ? AND writeoff_id = ? AND flow_type = ?", data.OrderID, *data.WriteoffID, flowType).
		First(&cashflow).Error; err != nil {
		logger.Logger.Warn("查询核销流水失败，跳过核销通道统计",
			zap.String("order_no", data.OrderNo),
			zap.String("order_id", data.OrderID),
			zap.Int64("writeoff_id", *data.WriteoffID),
			zap.Error(err))
		return
	}

	// real_money = -cashflow.ChangeMoney（因为 ChangeMoney 是负数，表示扣减）
	// 根据文档：success_money 记录的是实际扣除金额（real_money）
	realMoney := -cashflow.ChangeMoney

	// parent_tax_money = int(最终核销费率 × 订单金额 / 100)
	// 根据文档：total_tax 记录的是最终核销手续费（parent_tax_money）
	// 从 cashflow.Tax 获取费率，然后计算手续费
	// 但更简单的方式是：parent_tax_money = 订单金额 - real_money
	parentTaxMoney := int64(data.Money) - realMoney

	// 判断是订单成功还是退款
	// 根据文档：订单退款时，使用负数更新统计
	// 如果 OrderBefore 是成功状态（4=支付成功，6=支付成功通知已返回），且当前是退款状态，则使用负数
	// 但这里我们通过查询订单状态来判断
	var order models.Order
	if err := database.DB.Select("order_status").Where("id = ?", data.OrderID).First(&order).Error; err == nil {
		// 如果是退款状态，使用负数
		if order.OrderStatus == models.OrderStatusRefunded {
			// 订单退款：使用负数回退统计
			realMoney = -realMoney
			// 计算 parent_tax_money：从 cashflow.Tax（费率）和订单金额计算
			// parent_tax_money = int(cashflow.Tax * 订单金额 / 100)
			parentTaxMoney = int64(cashflow.Tax * float64(data.Money) / 100.0)
			parentTaxMoney = -parentTaxMoney // 使用负数

			logger.Logger.Info("核销通道统计回调（退款）",
				zap.String("order_no", data.OrderNo),
				zap.Int64("writeoff_id", *data.WriteoffID),
				zap.Int64("channel_id", data.ChannelID),
				zap.Int64("real_money", -realMoney),            // 显示原始值
				zap.Int64("parent_tax_money", -parentTaxMoney), // 显示原始值
				zap.Int("money", data.Money))
		} else {
			// 订单成功：使用正数
			logger.Logger.Info("核销通道统计回调（成功）",
				zap.String("order_no", data.OrderNo),
				zap.Int64("writeoff_id", *data.WriteoffID),
				zap.Int64("channel_id", data.ChannelID),
				zap.Int64("real_money", realMoney),
				zap.Int64("parent_tax_money", parentTaxMoney),
				zap.Int("money", data.Money))
		}
	}

	statsService := NewStatisticsService()
	stats := &models.WriteOffChannelDayStatistics{
		WriteoffID:   data.WriteoffID,
		PayChannelID: &data.ChannelID,
	}

	if err := statsService.SuccessBaseDayStatistics(ctx, stats, realMoney, parentTaxMoney, data.CreateDatetime, nil); err != nil {
		logger.Logger.Error("核销通道统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Int64("writeoff_id", *data.WriteoffID),
			zap.Int64("channel_id", data.ChannelID),
			zap.Error(err))
	}
}

// callbackDaySuccess 全局日统计回调
// 参考 Python: callback_day_success
func (s *OrderSuccessHookService) callbackDaySuccess(ctx context.Context, data *OrderSuccessData) {
	// 查询订单设备详情以获取设备类型
	var deviceDetail models.OrderDeviceDetail
	deviceType := models.DeviceTypeUnknown // 默认未知设备
	if err := database.DB.Where("order_id = ?", data.OrderID).First(&deviceDetail).Error; err == nil {
		deviceType = deviceDetail.DeviceType
	}

	// 构建设备统计更新字段
	deviceUpdateFields := map[string]interface{}{}
	switch deviceType {
	case models.DeviceTypeAndroid:
		deviceUpdateFields["android_count"] = gorm.Expr("android_count + 1")
	case models.DeviceTypeIOS:
		deviceUpdateFields["ios_count"] = gorm.Expr("ios_count + 1")
	case models.DeviceTypePC:
		deviceUpdateFields["pc_count"] = gorm.Expr("pc_count + 1")
	default:
		deviceUpdateFields["unknown_count"] = gorm.Expr("unknown_count + 1")
	}

	// 记录日志，帮助调试
	logger.Logger.Info("更新全局日统计",
		zap.String("order_no", data.OrderNo),
		zap.Int64("notify_money", int64(data.NotifyMoney)),
		zap.Int("tax", data.Tax),
		zap.Int("device_type", deviceType),
		zap.Int("order_tax_from_data", data.Tax))

	statsService := NewStatisticsService()
	stats := &models.DayStatistics{}

	// 更新成功统计（包含设备统计和手续费）
	// 注意：tax 是租户手续费，也就是系统总利润
	if err := statsService.SuccessBaseDayStatistics(ctx, stats, int64(data.NotifyMoney), int64(data.Tax), data.CreateDatetime, deviceUpdateFields); err != nil {
		logger.Logger.Error("全局日统计失败",
			zap.String("order_no", data.OrderNo),
			zap.Int("tax", data.Tax),
			zap.Error(err))
	} else {
		logger.Logger.Info("全局日统计更新成功",
			zap.String("order_no", data.OrderNo),
			zap.Int("tax", data.Tax),
			zap.Int64("notify_money", int64(data.NotifyMoney)))
	}
}
