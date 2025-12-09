package mq

import (
	"context"
	"fmt"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/order"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// tenantIDProviderAdapter 租户ID提供者适配器（实现 order.TenantIDProvider 接口）
type tenantIDProviderAdapter struct{}

func (a *tenantIDProviderAdapter) GetTenantIDByMerchantID(ctx context.Context, merchantID int64) (*int64, error) {
	// 查询商户信息获取租户ID（mq 包中直接查询数据库，避免循环依赖）
	var merchant models.Merchant
	if err := database.DB.Select("parent_id").Where("id = ?", merchantID).First(&merchant).Error; err == nil && merchant.ParentID > 0 {
		return &merchant.ParentID, nil
	}
	return nil, fmt.Errorf("无法获取租户ID")
}

// preTaxReleaserAdapter 预占余额释放器适配器（实现 order.PreTaxReleaser 接口）
type preTaxReleaserAdapter struct{}

func (a *preTaxReleaserAdapter) ReleasePreTax(ctx context.Context, tenantID int64, amount int64) error {
	if database.RDB == nil {
		return fmt.Errorf("Redis 未初始化")
	}

	key := fmt.Sprintf("tenant:pre_tax:%d", tenantID)

	// 使用 Lua 脚本确保原子性
	luaScript := `
		local preTaxKey = KEYS[1]
		local amount = tonumber(ARGV[1])
		
		local preTaxStr = redis.call('GET', preTaxKey)
		local preTax = preTaxStr and tonumber(preTaxStr) or 0
		
		-- 计算新的预占余额
		local newPreTax = math.max(0, preTax - amount)
		
		-- 更新预占余额
		redis.call('SET', preTaxKey, newPreTax)
		
		return newPreTax
	`

	result, err := database.RDB.Eval(ctx, luaScript, []string{key}, amount).Result()
	if err != nil {
		return fmt.Errorf("释放预占余额失败: %w", err)
	}

	// 记录日志
	logger.Logger.Debug("释放预占余额成功",
		zap.Int64("tenant_id", tenantID),
		zap.Int64("amount", amount),
		zap.Any("new_pre_tax", result))

	return nil
}

// updateOrderStatusDirectly 直接更新订单状态（避免循环依赖）
// 使用统一的 order.UpdateStatus 实现，避免代码重复
func updateOrderStatusDirectly(ctx context.Context, orderID string, status int, ticketNo string) error {
	// 创建适配器
	tenantIDProvider := &tenantIDProviderAdapter{}
	preTaxReleaser := &preTaxReleaserAdapter{}

	// 使用统一的订单状态更新逻辑
	return order.UpdateStatus(ctx, order.UpdateStatusRequest{
		OrderID:  orderID,
		Status:   status,
		TicketNo: ticketNo,
	}, order.UpdateStatusOptions{
		PreTaxReleaser:        preTaxReleaser,
		TenantIDProvider:      tenantIDProvider,
		HandleWriteoffBalance: false, // mq 包不需要处理码商余额
	})
}

// HandleOrderTimeout 处理订单超时（统一逻辑，对外暴露）
// 参考 Python: timeout_check 和 timeout_order
// 只处理状态为 [0, 2]（生成中、等待支付）的订单，更新订单状态为 7（已关闭）并释放预占余额
// 不处理已退款/已失败的订单（状态 [3, 5]）
func HandleOrderTimeout(ctx context.Context, orderNo string) error {
	// 查询订单（只查询状态为 [0, 2] 的订单）
	// 参考 Python: timeout_order 只处理状态为 [0, 2] 的订单
	var order models.Order
	if err := database.DB.Select("id, order_no, order_status, merchant_id, money").
		Where("order_no = ? AND order_status IN ?", orderNo, []int{
			models.OrderStatusGenerating, // 0 - 生成中
			models.OrderStatusPaying,     // 2 - 等待支付
		}).
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Logger.Info("订单不存在或状态不在超时检查范围内，跳过处理",
				zap.String("order_no", orderNo),
				zap.String("note", "只处理生成中(0)和待支付(2)状态的订单"))
			return nil // 订单不存在或状态不在检查范围内，不返回错误
		}
		logger.Logger.Error("查询订单失败",
			zap.String("order_no", orderNo),
			zap.Error(err))
		return err
	}

	logger.Logger.Info("开始处理订单超时",
		zap.String("order_id", order.ID),
		zap.String("order_no", orderNo),
		zap.Int("current_status", order.OrderStatus))

	// 状态为 [0, 2]，更新订单状态为已关闭（状态 7）并释放预占余额
	// 参考 Python: timeout_order(order_no) - 只处理状态为 [0, 2] 的订单
	if err := updateOrderStatusDirectly(ctx, order.ID, models.OrderStatusClosed, ""); err != nil {
		logger.Logger.Error("更新超时订单状态失败",
			zap.String("order_id", order.ID),
			zap.String("order_no", orderNo),
			zap.Error(err))
		return err
	}
	logger.Logger.Info("订单已超时，状态已更新为已关闭",
		zap.String("order_id", order.ID),
		zap.String("order_no", orderNo),
		zap.Int("old_status", order.OrderStatus),
		zap.Int("new_status", models.OrderStatusClosed))
	// 注意：预占余额的释放已经在 updateOrderStatusDirectly 中处理（OrderStatusCancelled = OrderStatusClosed）

	return nil
}
