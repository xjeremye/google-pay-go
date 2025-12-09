package service

import (
	"context"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/mq"
	"github.com/golang-pay-core/internal/plugin"
	"go.uber.org/zap"
)

// OrderTimeoutService 订单超时检查服务
// 定时扫描超时的订单并更新状态
type OrderTimeoutService struct {
	orderService *OrderService
	stopChan     chan struct{}
}

// NewOrderTimeoutService 创建订单超时检查服务
func NewOrderTimeoutService() *OrderTimeoutService {
	return &OrderTimeoutService{
		orderService: NewOrderService(),
		stopChan:     make(chan struct{}),
	}
}

// Start 启动超时检查服务
// 每30秒执行一次超时检查
func (s *OrderTimeoutService) Start(ctx context.Context) {
	// 每30秒执行一次超时检查
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	logger.Logger.Info("订单超时检查服务已启动（每30秒检查一次）")

	// 立即执行一次
	s.checkExpiredOrders(ctx)

	for {
		select {
		case <-ticker.C:
			// 定时执行超时检查
			s.checkExpiredOrders(ctx)
		case <-s.stopChan:
			logger.Logger.Info("订单超时检查服务已停止")
			return
		case <-ctx.Done():
			logger.Logger.Info("订单超时检查服务已停止（上下文取消）")
			return
		}
	}
}

// Stop 停止超时检查服务
func (s *OrderTimeoutService) Stop() {
	close(s.stopChan)
}

// checkExpiredOrders 检查并处理超时的订单（兜底机制）
// 如果延迟消息失败或服务重启，定时扫描会作为兜底
func (s *OrderTimeoutService) checkExpiredOrders(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("订单超时检查服务异常",
				zap.Any("panic", r))
		}
	}()

	now := time.Now()

	// 查询需要检查的订单（生成中或支付中状态）
	// 只查询最近2小时内的订单，避免扫描过多数据
	// 注意：这是兜底机制，延迟消息是主要方式
	twoHoursAgo := now.Add(-2 * time.Hour)
	var orders []struct {
		ID             string
		OrderNo        string
		OrderStatus    int
		CreateDatetime *time.Time
		PluginID       *int64
	}

	// 查询订单及其详情（包含插件ID）
	// 只查询状态为 [0, 2]（生成中、等待支付）的订单
	// 参考 Python: timeout_order 只处理状态为 [0, 2] 的订单
	// 参考 Python: 从插件配置获取超时时间，而不是从域名配置
	if err := database.DB.Table("dvadmin_order").
		Select("dvadmin_order.id, dvadmin_order.order_no, dvadmin_order.order_status, dvadmin_order.create_datetime, dvadmin_order_detail.plugin_id").
		Joins("LEFT JOIN dvadmin_order_detail ON dvadmin_order.id = dvadmin_order_detail.order_id").
		Where("dvadmin_order.order_status IN ?", []int{
			models.OrderStatusGenerating, // 0 - 生成中
			models.OrderStatusPaying,     // 2 - 等待支付
		}).
		Where("dvadmin_order.create_datetime >= ?", twoHoursAgo).
		Where("dvadmin_order_detail.plugin_id IS NOT NULL").
		Scan(&orders).Error; err != nil {
		logger.Logger.Error("查询待检查订单失败", zap.Error(err))
		return
	}

	if len(orders) == 0 {
		return
	}

	logger.Logger.Debug("开始检查订单超时（兜底机制）",
		zap.Int("order_count", len(orders)))

	// 检查每个订单是否超时
	expiredCount := 0
	for _, order := range orders {
		if order.CreateDatetime == nil || order.PluginID == nil {
			continue
		}

		// 获取订单的超时时间（从插件配置获取）
		// 参考 Python: get_plugin_out_time(ctx.plugin.id)
		// 使用 BasePlugin 的 GetTimeout 方法，它会从插件配置中获取 out_time
		basePlugin := plugin.NewBasePlugin(*order.PluginID)
		timeoutSeconds := basePlugin.GetTimeout(ctx, *order.PluginID)

		expireTime := order.CreateDatetime.Add(time.Duration(timeoutSeconds) * time.Second)

		// 检查是否已过期
		if now.After(expireTime) {
			// 订单已过期，使用统一的超时处理函数（与延迟消息使用相同的逻辑）
			// 参考 Python: timeout_check 的逻辑
			logger.Logger.Info("发现过期订单，开始处理",
				zap.String("order_id", order.ID),
				zap.String("order_no", order.OrderNo),
				zap.Int("order_status", order.OrderStatus),
				zap.Time("create_time", *order.CreateDatetime),
				zap.Time("expire_time", expireTime),
				zap.Int("timeout_seconds", timeoutSeconds),
				zap.Duration("overdue_time", now.Sub(expireTime)))

			if err := mq.HandleOrderTimeout(ctx, order.OrderNo); err != nil {
				logger.Logger.Error("处理订单超时失败",
					zap.String("order_id", order.ID),
					zap.String("order_no", order.OrderNo),
					zap.Int("order_status", order.OrderStatus),
					zap.Error(err))
				// 即使失败也计入 expiredCount，因为确实过期了
				expiredCount++
			} else {
				expiredCount++
				logger.Logger.Info("订单已超时，处理完成（通过定时扫描兜底）",
					zap.String("order_id", order.ID),
					zap.String("order_no", order.OrderNo),
					zap.Int("order_status", order.OrderStatus),
					zap.Time("create_time", *order.CreateDatetime),
					zap.Time("expire_time", expireTime),
					zap.Int("timeout_seconds", timeoutSeconds))
			}
		}
	}

	if expiredCount > 0 {
		logger.Logger.Info("订单超时检查完成（兜底机制）",
			zap.Int("total_checked", len(orders)),
			zap.Int("expired_count", expiredCount),
			zap.String("note", "建议检查延迟消息是否正常工作"))
	}
}
