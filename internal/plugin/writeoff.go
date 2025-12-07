package plugin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
)

// checkWriteoffBalance 检查码商余额是否足够（使用 Redis）
// 如果 balance 为 nil（无限制）或 balance >= amount，返回 true
func checkWriteoffBalance(ctx context.Context, redisClient *redis.Client, writeoffID int64, amount int64) (bool, error) {
	key := fmt.Sprintf("writeoff:balance:%d", writeoffID)
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		// 如果 Redis 中没有，从数据库初始化并检查
		var balance *int64
		if err := database.DB.Table("dvadmin_writeoff").
			Where("id = ?", writeoffID).
			Select("balance").
			Scan(&balance).Error; err != nil {
			return false, err
		}

		// 初始化到 Redis
		if balance == nil {
			_ = redisClient.Set(ctx, key, "NULL", 0).Err()
			return true, nil // 无限制
		}
		_ = redisClient.Set(ctx, key, *balance, 0).Err()
		return *balance >= amount, nil
	}
	if err != nil {
		return false, err
	}

	// 检查是否是特殊标记 "NULL"（表示数据库中的 NULL）
	if val == "NULL" {
		return true, nil // 无限制
	}

	balance, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return false, err
	}
	return balance >= amount, nil
}

// getWriteoffIDsForPlugin 获取可用的核销ID列表（供插件使用）
// 参考 Python: get_writeoff_ids(tenant_id, money, pay_channel_id=None)
// 使用 Redis 检查码商余额，提高性能
func getWriteoffIDsForPlugin(tenantID int64, money int, payChannelID *int64) ([]int64, error) {
	ctx := context.Background()
	redisClient := database.RDB

	// 先查询所有符合条件的核销ID（不检查余额）
	var allWriteoffIDs []int64
	query := database.DB.Model(&models.Writeoff{}).
		Joins("JOIN dvadmin_system_users ON dvadmin_writeoff.system_user_id = dvadmin_system_users.id").
		Where("dvadmin_writeoff.parent_id = ?", tenantID).
		Where("dvadmin_system_users.status = ?", true).
		Where("dvadmin_system_users.is_active = ?", true)

	if err := query.Pluck("dvadmin_writeoff.id", &allWriteoffIDs).Error; err != nil {
		return nil, fmt.Errorf("查询核销ID失败: %w", err)
	}

	// 使用 Redis 检查每个码商的余额
	writeoffIDs := make([]int64, 0)
	for _, writeoffID := range allWriteoffIDs {
		// 检查余额是否足够（使用 Redis）
		ok, err := checkWriteoffBalance(ctx, redisClient, writeoffID, int64(money))
		if err != nil {
			// 如果检查失败，降级到数据库查询
			var balance *int64
			if err := database.DB.Table("dvadmin_writeoff").
				Where("id = ?", writeoffID).
				Select("balance").
				Scan(&balance).Error; err == nil {
				if balance == nil || *balance >= int64(money) {
					writeoffIDs = append(writeoffIDs, writeoffID)
				}
			}
			continue
		}
		if ok {
			writeoffIDs = append(writeoffIDs, writeoffID)
		}
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
