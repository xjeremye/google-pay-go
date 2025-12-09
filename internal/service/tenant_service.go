package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"go.uber.org/zap"
)

// TenantService 租户服务
type TenantService struct{}

// NewTenantService 创建租户服务
func NewTenantService() *TenantService {
	return &TenantService{}
}

// DeductTax 扣除租户手续费
// 参考 Python: tenant_success_tax(tenant_id, tax, order_id, channel_id, reorder, remarks="手续费")
func (s *TenantService) DeductTax(ctx context.Context, tenantID int64, tax int, orderID string, channelID int64) error {
	if tax == 0 {
		return nil
	}

	// 查询租户
	var tenant models.Tenant
	if err := database.DB.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		return fmt.Errorf("查询租户失败: %w", err)
	}

	// 记录扣费前的余额
	beforeBalance := tenant.Balance

	// 扣除手续费
	tenant.Balance -= int64(tax)
	if err := database.DB.Save(&tenant).Error; err != nil {
		return fmt.Errorf("扣除租户手续费失败: %w", err)
	}

	// 记录流水
	// 参考 Python: TenantCashFlow.objects.create(...)
	now := time.Now()
	cashFlow := &models.TenantCashFlow{
		TenantID:       tenantID,
		OldMoney:       beforeBalance,
		NewMoney:       tenant.Balance,
		ChangeMoney:    -int64(tax), // 负数表示扣费
		FlowType:       1,           // 1=消费
		Remarks:        "手续费",
		OrderID:        &orderID,
		PayChannelID:   &channelID,
		CreateDatetime: &now,
	}

	if err := database.DB.Create(cashFlow).Error; err != nil {
		logger.Logger.Warn("记录租户流水失败",
			zap.Int64("tenant_id", tenantID),
			zap.String("order_id", orderID),
			zap.Error(err))
		// 不返回错误，因为扣费已经成功
	}

	logger.Logger.Info("租户扣费成功",
		zap.Int64("tenant_id", tenantID),
		zap.String("order_id", orderID),
		zap.Int64("before_balance", beforeBalance),
		zap.Int64("after_balance", tenant.Balance),
		zap.Int("tax", tax))

	return nil
}
