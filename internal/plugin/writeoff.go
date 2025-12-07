package plugin

import (
	"fmt"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
)

// getWriteoffIDsForPlugin 获取可用的核销ID列表（供插件使用）
// 参考 Python: get_writeoff_ids(tenant_id, money, pay_channel_id=None)
func getWriteoffIDsForPlugin(tenantID int64, money int, payChannelID *int64) ([]int64, error) {
	var writeoffIDs []int64

	// 构建查询条件
	// Python: WriteOff.objects.filter(
	//     Q(balance__isnull=True) | Q(balance__gte=money),
	//     system_user__status=True,
	//     parent_id=tenant_id,
	//     system_user__is_active=True,
	// ).values("id", "balance")
	query := database.DB.Model(&models.Writeoff{}).
		Joins("JOIN dvadmin_system_users ON dvadmin_writeoff.system_user_id = dvadmin_system_users.id").
		Where("dvadmin_writeoff.parent_id = ?", tenantID).
		Where("dvadmin_system_users.status = ?", true).
		Where("dvadmin_system_users.is_active = ?", true).
		Where("dvadmin_writeoff.balance IS NULL OR dvadmin_writeoff.balance >= ?", money)

	if err := query.Pluck("dvadmin_writeoff.id", &writeoffIDs).Error; err != nil {
		return nil, fmt.Errorf("查询核销ID失败: %w", err)
	}

	// 过滤掉被禁用的支付通道关联
	// Python: if WriteoffPayChannel.objects.filter(pay_channel_id=pay_channel_id, writeoff_id=writeoff['id'], status=False).exists():
	if payChannelID != nil {
		var disabledWriteoffIDs []int64
		if err := database.DB.Model(&models.WriteoffPayChannel{}).
			Where("pay_channel_id = ? AND status = ?", *payChannelID, false).
			Pluck("writeoff_id", &disabledWriteoffIDs).Error; err != nil {
			return nil, fmt.Errorf("查询禁用的核销通道关联失败: %w", err)
		}

		// 过滤掉被禁用的核销ID
		disabledMap := make(map[int64]bool)
		for _, id := range disabledWriteoffIDs {
			disabledMap[id] = true
		}
		enabledWriteoffIDs := make([]int64, 0)
		for _, id := range writeoffIDs {
			if !disabledMap[id] {
				enabledWriteoffIDs = append(enabledWriteoffIDs, id)
			}
		}
		writeoffIDs = enabledWriteoffIDs
	}

	return writeoffIDs, nil
}

