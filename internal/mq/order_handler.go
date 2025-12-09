package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// updateOrderStatusDirectly 直接更新订单状态（避免循环依赖）
// 参考 service.OrderService.UpdateOrderStatus 的逻辑
func updateOrderStatusDirectly(ctx context.Context, orderID string, status int, ticketNo string) error {
	// 先查询订单信息（在事务外，减少事务时间）
	var order models.Order
	if err := database.DB.Select("id, merchant_id, money, order_status").
		Where("id = ?", orderID).
		First(&order).Error; err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	// 检查订单状态是否已经变更（避免重复处理）
	if order.OrderStatus == status {
		return nil // 状态未变化，直接返回
	}

	// 查询商户信息获取租户ID
	var tenantID *int64
	if order.MerchantID != nil {
		var merchant models.Merchant
		if err := database.DB.Select("parent_id").Where("id = ?", *order.MerchantID).First(&merchant).Error; err == nil && merchant.ParentID > 0 {
			tenantID = &merchant.ParentID
		}
	}

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// 处理租户余额（在事务中，确保一致性）
	if tenantID != nil {
		// 根据订单状态处理预占余额和余额
		switch status {
		case models.OrderStatusPaid:
			// 订单支付成功：从数据库扣减余额
			// 先查询变更前的余额（使用 SELECT FOR UPDATE 确保一致性），用于记录流水
			var tenant models.Tenant
			if err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("id = ?", *tenantID).First(&tenant).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("查询租户余额失败: %w", err)
			}

			oldBalance := tenant.Balance
			newBalance := oldBalance - int64(order.Money)

			// 使用原子操作扣减余额
			if err := tx.Model(&models.Tenant{}).
				Where("id = ?", *tenantID).
				Update("balance", gorm.Expr("balance - ?", order.Money)).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("扣减租户余额失败: %w", err)
			}

			// 记录租户资金流水
			cashflow := &models.TenantCashflow{
				OldMoney:       oldBalance,
				NewMoney:       newBalance,
				ChangeMoney:    -int64(order.Money), // 负数表示扣减
				FlowType:       models.CashflowTypeOrderDeduct,
				OrderID:        &orderID,
				PayChannelID:   order.PayChannelID,
				TenantID:       *tenantID,
				CreateDatetime: &now,
			}
			if err := tx.Create(cashflow).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("记录租户资金流水失败: %w", err)
			}

			// 释放预占余额（从 Redis）
			// 直接操作 Redis，避免循环依赖
			if err := releasePreTaxDirectly(ctx, *tenantID, int64(order.Money)); err != nil {
				// 如果 Redis 释放失败，记录日志但不回滚数据库事务（数据库已扣减）
				logger.Logger.Warn("释放预占余额失败",
					zap.Int64("tenant_id", *tenantID),
					zap.Int64("amount", int64(order.Money)),
					zap.Error(err))
			}

			logger.Logger.Info("订单支付成功，已扣减租户余额",
				zap.String("order_id", orderID),
				zap.Int64("tenant_id", *tenantID),
				zap.Int64("old_balance", oldBalance),
				zap.Int64("new_balance", newBalance),
				zap.Int("money", order.Money))

		case models.OrderStatusFailed, models.OrderStatusCancelled:
			// 订单失败/取消/过期：只从 Redis 释放预占，不扣减余额
			if err := releasePreTaxDirectly(ctx, *tenantID, int64(order.Money)); err != nil {
				tx.Rollback()
				return fmt.Errorf("释放预占余额失败: %w", err)
			}
			logger.Logger.Info("订单失败/取消/过期，已释放预占余额",
				zap.String("order_id", orderID),
				zap.Int64("tenant_id", *tenantID),
				zap.Int("money", order.Money))
		}
	}

	// 更新订单状态（合并多个字段更新，减少数据库往返）
	updates := map[string]interface{}{
		"order_status":    status,
		"update_datetime": &now,
	}

	if status == models.OrderStatusPaid {
		updates["pay_datetime"] = &now
	}

	if err := tx.Model(&models.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 更新订单详情的 ticket_no（如果提供）
	if ticketNo != "" {
		if err := tx.Model(&models.OrderDetail{}).
			Where("order_id = ?", orderID).
			Update("ticket_no", ticketNo).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新订单详情 ticket_no 失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// releasePreTaxDirectly 直接释放预占余额（避免循环依赖）
// 参考 service.BalanceService.ReleasePreTax 的逻辑
func releasePreTaxDirectly(ctx context.Context, tenantID int64, amount int64) error {
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
