package order

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

// PreTaxReleaser 预占余额释放器接口（避免循环依赖）
type PreTaxReleaser interface {
	// ReleasePreTax 释放预占余额
	ReleasePreTax(ctx context.Context, tenantID int64, amount int64) error
}

// TenantIDProvider 租户ID提供者接口（避免循环依赖）
type TenantIDProvider interface {
	// GetTenantIDByMerchantID 根据商户ID获取租户ID
	GetTenantIDByMerchantID(ctx context.Context, merchantID int64) (*int64, error)
}

// UpdateStatusRequest 更新订单状态请求
type UpdateStatusRequest struct {
	OrderID  string
	Status   int
	TicketNo string
}

// UpdateStatusOptions 更新订单状态选项
type UpdateStatusOptions struct {
	PreTaxReleaser   PreTaxReleaser
	TenantIDProvider TenantIDProvider
	// 是否处理码商余额（默认 false）
	HandleWriteoffBalance bool
}

// UpdateStatus 更新订单状态的核心逻辑（统一实现，避免代码重复）
// 此函数可以被 service 和 mq 包复用
func UpdateStatus(ctx context.Context, req UpdateStatusRequest, opts UpdateStatusOptions) error {
	// 先查询订单信息（在事务外，减少事务时间）
	var order models.Order
	if err := database.DB.Select("id, merchant_id, money, order_status, pay_channel_id, writeoff_id").
		Where("id = ?", req.OrderID).
		First(&order).Error; err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	// 检查订单状态是否已经变更（避免重复处理）
	if order.OrderStatus == req.Status {
		logger.Logger.Debug("订单状态未变化，跳过更新",
			zap.String("order_id", req.OrderID),
			zap.Int("current_status", order.OrderStatus),
			zap.Int("target_status", req.Status))
		return nil // 状态未变化，直接返回
	}

	logger.Logger.Info("准备更新订单状态",
		zap.String("order_id", req.OrderID),
		zap.Int("old_status", order.OrderStatus),
		zap.Int("new_status", req.Status))

	// 获取租户ID
	var tenantID *int64
	if order.MerchantID != nil && opts.TenantIDProvider != nil {
		var err error
		tenantID, err = opts.TenantIDProvider.GetTenantIDByMerchantID(ctx, *order.MerchantID)
		if err != nil {
			logger.Logger.Warn("获取租户ID失败，将跳过余额处理",
				zap.String("order_id", req.OrderID),
				zap.Int64("merchant_id", *order.MerchantID),
				zap.Error(err))
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
	if tenantID != nil && opts.PreTaxReleaser != nil {
		// 根据订单状态处理预占余额和余额
		switch req.Status {
		case models.OrderStatusPaid:
			// 订单支付成功：从数据库扣减余额，从 Redis 释放预占
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
				OrderID:        &req.OrderID,
				PayChannelID:   order.PayChannelID,
				TenantID:       *tenantID,
				CreateDatetime: &now,
			}
			if err := tx.Create(cashflow).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("记录租户资金流水失败: %w", err)
			}

			// 从 Redis 释放预占余额
			if err := opts.PreTaxReleaser.ReleasePreTax(ctx, *tenantID, int64(order.Money)); err != nil {
				// 如果 Redis 释放失败，记录日志但不回滚数据库事务（数据库已扣减）
				logger.Logger.Warn("释放预占余额失败",
					zap.Int64("tenant_id", *tenantID),
					zap.Int64("amount", int64(order.Money)),
					zap.Error(err))
			}

			logger.Logger.Info("订单支付成功，已扣减租户余额",
				zap.String("order_id", req.OrderID),
				zap.Int64("tenant_id", *tenantID),
				zap.Int64("old_balance", oldBalance),
				zap.Int64("new_balance", newBalance),
				zap.Int("money", order.Money))

		case models.OrderStatusFailed, models.OrderStatusClosed:
			// 订单失败/取消/过期/关闭：只从 Redis 释放预占，不扣减余额
			// 注意：OrderStatusCancelled 是 OrderStatusClosed 的别名，值相同，所以不需要单独列出
			if err := opts.PreTaxReleaser.ReleasePreTax(ctx, *tenantID, int64(order.Money)); err != nil {
				tx.Rollback()
				return fmt.Errorf("释放预占余额失败: %w", err)
			}
			logger.Logger.Info("订单失败/取消/过期/关闭，已释放预占余额",
				zap.String("order_id", req.OrderID),
				zap.Int64("tenant_id", *tenantID),
				zap.Int("money", order.Money),
				zap.Int("status", req.Status))
		}
	}

	// 处理码商余额（在事务中，确保一致性）
	// 码商余额扣减逻辑与租户相同：订单支付成功时从数据库扣减余额
	if opts.HandleWriteoffBalance && order.WriteoffID != nil {
		switch req.Status {
		case models.OrderStatusPaid:
			// 订单支付成功：从数据库扣减码商余额
			// 先查询变更前的余额（使用 SELECT FOR UPDATE 确保一致性），用于记录流水
			var writeoff models.Writeoff
			if err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("id = ?", *order.WriteoffID).First(&writeoff).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("查询码商余额失败: %w", err)
			}

			// 记录码商资金流水（无论余额是否为 NULL，都需要记录流水以便追踪）
			var oldBalance, newBalance int64
			if writeoff.Balance != nil {
				// 有余额限制：扣减余额并记录实际余额变化
				oldBalance = *writeoff.Balance
				newBalance = oldBalance - int64(order.Money)

				// 使用原子操作扣减余额
				if err := tx.Model(&models.Writeoff{}).
					Where("id = ? AND balance IS NOT NULL", *order.WriteoffID).
					Update("balance", gorm.Expr("balance - ?", order.Money)).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("扣减码商余额失败: %w", err)
				}
			} else {
				// 余额无限制：old_money 和 new_money 都设置为 0（表示无限制）
				// 这样仍然可以记录流水以便追踪订单使用了哪些码商
				oldBalance = 0
				newBalance = 0
			}

			// 记录码商资金流水
			cashflow := &models.WriteoffCashflow{
				OldMoney:       oldBalance,
				NewMoney:       newBalance,
				ChangeMoney:    -int64(order.Money), // 负数表示扣减
				FlowType:       models.CashflowTypeOrderDeduct,
				Tax:            0.00, // 费率，可以根据需要设置
				OrderID:        &req.OrderID,
				PayChannelID:   order.PayChannelID,
				WriteoffID:     *order.WriteoffID,
				CreateDatetime: &now,
			}
			if err := tx.Create(cashflow).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("记录码商资金流水失败: %w", err)
			}
		case models.OrderStatusFailed, models.OrderStatusClosed:
			// 订单失败/取消/过期：码商余额不需要处理（码商没有预占余额的概念）
			// 码商余额只在支付成功时扣减
			// 注意：OrderStatusCancelled 是 OrderStatusClosed 的别名，值相同，所以不需要单独列出
		}
	}

	// 更新订单状态（合并多个字段更新，减少数据库往返）
	updates := map[string]interface{}{
		"order_status":    req.Status,
		"update_datetime": &now,
	}

	if req.Status == models.OrderStatusPaid {
		updates["pay_datetime"] = &now
	}

	rowsAffected := tx.Model(&models.Order{}).
		Where("id = ?", req.OrderID).
		Updates(updates).RowsAffected
	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败：未找到订单或状态未变化")
	}
	if err := tx.Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	logger.Logger.Info("订单状态更新成功",
		zap.String("order_id", req.OrderID),
		zap.Int("old_status", order.OrderStatus),
		zap.Int("new_status", req.Status),
		zap.Int64("rows_affected", rowsAffected))

	// 更新订单详情的 ticket_no（如果提供）
	if req.TicketNo != "" {
		if err := tx.Model(&models.OrderDetail{}).
			Where("order_id = ?", req.OrderID).
			Update("ticket_no", req.TicketNo).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新订单详情 ticket_no 失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 如果订单状态更新为"支付成功，通知未返回"或"支付成功，通知已返回"，触发成功钩子
	// 注意：在事务提交后异步触发，避免影响主流程
	// 为了避免循环依赖，这里只记录日志，实际触发逻辑应该在调用方处理
	if req.Status == models.OrderStatusPaidNoNotify || req.Status == models.OrderStatusPaid {
		logger.Logger.Info("订单状态更新为成功，需要在调用方触发成功钩子",
			zap.String("order_id", req.OrderID),
			zap.Int("status", req.Status))
		// TODO: 通过消息队列或事件系统触发成功钩子，避免循环依赖
	}

	return nil
}
