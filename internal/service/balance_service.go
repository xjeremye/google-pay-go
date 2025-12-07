package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"go.uber.org/zap"
)

// BalanceService Redis 余额管理服务
type BalanceService struct {
	redis *redis.Client
}

// NewBalanceService 创建余额管理服务
func NewBalanceService() *BalanceService {
	return &BalanceService{
		redis: database.RDB,
	}
}

// GetBalance 获取租户余额（从 Redis）
// 如果 Redis 中没有，从数据库查询并初始化到 Redis
func (s *BalanceService) GetBalance(ctx context.Context, tenantID int64) (int64, error) {
	key := fmt.Sprintf("tenant:balance:%d", tenantID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// 如果 Redis 中没有，从数据库查询
		var balance int64
		if err := database.DB.Table("dvadmin_tenant").
			Where("id = ?", tenantID).
			Select("balance").
			Scan(&balance).Error; err != nil {
			return 0, err
		}
		// 将余额写入 Redis（缓存刷新服务会定期同步，这里只是初始化）
		_ = s.redis.Set(ctx, key, balance, 0).Err()
		return balance, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// GetPreTax 获取租户预占余额（从 Redis）
// 如果 Redis 中没有，默认返回 0（预占余额默认就是 0）
func (s *BalanceService) GetPreTax(ctx context.Context, tenantID int64) (int64, error) {
	key := fmt.Sprintf("tenant:pre_tax:%d", tenantID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// 如果 Redis 中没有，默认返回 0（预占余额默认就是 0）
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// GetTrust 获取租户信任标志（从 Redis）
func (s *BalanceService) GetTrust(ctx context.Context, tenantID int64) (bool, error) {
	key := fmt.Sprintf("tenant:trust:%d", tenantID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// 如果 Redis 中没有，从数据库初始化
		return s.initTrustFromDB(ctx, tenantID)
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

// ReserveBalance 预占余额（原子操作）
// 返回是否成功，以及当前可用余额
func (s *BalanceService) ReserveBalance(ctx context.Context, tenantID int64, amount int64) (bool, int64, error) {
	// 确保余额和信任标志已初始化（预占余额默认为 0，不需要初始化）
	if _, err := s.GetBalance(ctx, tenantID); err != nil {
		return false, 0, fmt.Errorf("获取余额失败: %w", err)
	}
	if _, err := s.GetTrust(ctx, tenantID); err != nil {
		return false, 0, fmt.Errorf("获取信任标志失败: %w", err)
	}

	// 使用 Lua 脚本确保原子性：检查余额、预占余额、更新预占余额
	luaScript := `
		local balanceKey = KEYS[1]
		local preTaxKey = KEYS[2]
		local trustKey = KEYS[3]
		local amount = tonumber(ARGV[1])
		
		-- 获取当前余额和预占余额
		local balanceStr = redis.call('GET', balanceKey)
		local preTaxStr = redis.call('GET', preTaxKey)
		local trustStr = redis.call('GET', trustKey)
		
		-- 如果余额或信任标志不存在，返回错误
		if not balanceStr or not trustStr then
			return {0, 0, 0}  -- 未初始化
		end
		
		local balance = tonumber(balanceStr)
		local preTax = tonumber(preTaxStr) or 0  -- 如果预占余额不存在，默认为 0
		local trust = trustStr == '1'
		
		-- 计算可用余额
		local availableBalance = balance - preTax
		
		-- 检查余额是否足够
		if availableBalance < amount then
			-- 如果不允许负数，拒绝
			if not trust then
				return {0, availableBalance, 0}  -- 余额不足
			end
		end
		
		-- 增加预占余额
		local newPreTax = preTax + amount
		redis.call('INCRBY', preTaxKey, amount)
		
		return {1, availableBalance - amount, newPreTax}  -- 成功，返回新的可用余额和预占余额
	`

	balanceKey := fmt.Sprintf("tenant:balance:%d", tenantID)
	preTaxKey := fmt.Sprintf("tenant:pre_tax:%d", tenantID)
	trustKey := fmt.Sprintf("tenant:trust:%d", tenantID)

	result, err := s.redis.Eval(ctx, luaScript, []string{balanceKey, preTaxKey, trustKey}, amount).Result()
	if err != nil {
		return false, 0, err
	}

	// 解析结果
	results, ok := result.([]interface{})
	if !ok || len(results) != 3 {
		return false, 0, fmt.Errorf("Lua 脚本返回格式错误: %v", result)
	}

	success := false
	if val, ok := results[0].(int64); ok {
		success = val == 1
	} else if val, ok := results[0].(int); ok {
		success = val == 1
	}

	availableBalance := int64(0)
	if val, ok := results[1].(int64); ok {
		availableBalance = val
	} else if val, ok := results[1].(int); ok {
		availableBalance = int64(val)
	}

	return success, availableBalance, nil
}

// ReleasePreTax 释放预占余额（原子操作）
// 如果预占余额不存在，默认为 0，不需要释放
func (s *BalanceService) ReleasePreTax(ctx context.Context, tenantID int64, amount int64) error {
	key := fmt.Sprintf("tenant:pre_tax:%d", tenantID)
	// 使用 Lua 脚本确保原子性，如果预占余额不存在或为 0，不需要操作
	luaScript := `
		local preTaxKey = KEYS[1]
		local amount = tonumber(ARGV[1])
		local preTax = tonumber(redis.call('GET', preTaxKey) or '0')
		local newPreTax = preTax - amount
		if newPreTax < 0 then
			newPreTax = 0
		end
		redis.call('SET', preTaxKey, newPreTax)
		return 1
	`
	_, err := s.redis.Eval(ctx, luaScript, []string{key}, amount).Result()
	return err
}

// DeductBalance 扣减余额并释放预占（已废弃）
// 注意：余额扣减现在在数据库中进行（订单支付成功时），此方法不再使用
// 保留此方法仅用于向后兼容，实际余额扣减在 UpdateOrderStatus 中通过数据库操作完成
// Deprecated: 使用数据库直接扣减余额，使用 ReleasePreTax 释放预占
func (s *BalanceService) DeductBalance(ctx context.Context, tenantID int64, amount int64) error {
	// 只释放预占，不扣减余额（余额扣减在数据库中进行）
	return s.ReleasePreTax(ctx, tenantID, amount)
}

// initTrustFromDB 从数据库初始化信任标志到 Redis
func (s *BalanceService) initTrustFromDB(ctx context.Context, tenantID int64) (bool, error) {
	var trust bool
	err := database.DB.Table("dvadmin_tenant").
		Where("id = ?", tenantID).
		Select("trust").
		Scan(&trust).Error
	if err != nil {
		return false, err
	}

	// 写入 Redis
	key := fmt.Sprintf("tenant:trust:%d", tenantID)
	val := "0"
	if trust {
		val = "1"
	}
	if err := s.redis.Set(ctx, key, val, 0).Err(); err != nil {
		logger.Logger.Warn("初始化信任标志到 Redis 失败",
			zap.Int64("tenant_id", tenantID),
			zap.Error(err))
	}

	return trust, nil
}

// SyncToDB 同步 Redis 余额到数据库（已废弃）
// 注意：现在余额扣减在数据库中进行，不需要从 Redis 同步到数据库
// 缓存刷新服务会定期从数据库同步余额到 Redis（只读同步）
// Deprecated: 余额扣减在数据库中进行，不需要反向同步
func (s *BalanceService) SyncToDB(ctx context.Context, tenantID int64) error {
	// 不再同步，余额扣减在数据库中进行
	return nil
}

// ========== 码商（Writeoff）余额管理 ==========

// GetWriteoffBalance 获取码商余额（从 Redis）
// 返回余额值，如果为 nil 表示无限制
// 如果 Redis 中没有，从缓存服务获取码商信息
func (s *BalanceService) GetWriteoffBalance(ctx context.Context, writeoffID int64) (*int64, error) {
	key := fmt.Sprintf("writeoff:balance:%d", writeoffID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// 如果 Redis 中没有，从缓存服务获取码商信息
		cacheService := NewCacheService()
		writeoff, _, err := cacheService.GetWriteoffWithUser(ctx, writeoffID)
		if err != nil {
			return nil, err
		}
		// 将余额写入 Redis（缓存刷新服务会定期同步，这里只是初始化）
		if writeoff.Balance == nil {
			_ = s.redis.Set(ctx, key, "NULL", 0).Err()
			return nil, nil
		}
		_ = s.redis.Set(ctx, key, *writeoff.Balance, 0).Err()
		return writeoff.Balance, nil
	}
	if err != nil {
		return nil, err
	}

	// 检查是否是特殊标记 "NULL"（表示数据库中的 NULL）
	if val == "NULL" {
		return nil, nil
	}

	balance, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// CheckWriteoffBalance 检查码商余额是否足够
// 如果 balance 为 nil（无限制）或 balance >= amount，返回 true
func (s *BalanceService) CheckWriteoffBalance(ctx context.Context, writeoffID int64, amount int64) (bool, error) {
	balance, err := s.GetWriteoffBalance(ctx, writeoffID)
	if err != nil {
		return false, err
	}

	// 如果余额为 nil，表示无限制，总是返回 true
	if balance == nil {
		return true, nil
	}

	// 检查余额是否足够
	return *balance >= amount, nil
}

// DeductWriteoffBalance 扣减码商余额（原子操作）
func (s *BalanceService) DeductWriteoffBalance(ctx context.Context, writeoffID int64, amount int64) error {
	// 使用 Lua 脚本确保原子性
	luaScript := `
		local balanceKey = KEYS[1]
		local amount = tonumber(ARGV[1])
		
		local balanceStr = redis.call('GET', balanceKey)
		
		-- 如果余额不存在或为 NULL，不允许扣减
		if not balanceStr or balanceStr == 'NULL' then
			return {0, '余额无限制，不允许扣减'}  -- 无限制的余额不允许扣减
		end
		
		local balance = tonumber(balanceStr)
		
		-- 检查余额是否足够
		if balance < amount then
			return {0, '余额不足'}  -- 余额不足
		end
		
		-- 扣减余额
		local newBalance = balance - amount
		redis.call('SET', balanceKey, newBalance)
		
		return {1, newBalance}  -- 成功，返回新余额
	`

	balanceKey := fmt.Sprintf("writeoff:balance:%d", writeoffID)

	result, err := s.redis.Eval(ctx, luaScript, []string{balanceKey}, amount).Result()
	if err != nil {
		return err
	}

	// 解析结果
	results, ok := result.([]interface{})
	if !ok || len(results) != 2 {
		return fmt.Errorf("Lua 脚本返回格式错误: %v", result)
	}

	success := false
	if val, ok := results[0].(int64); ok {
		success = val == 1
	} else if val, ok := results[0].(int); ok {
		success = val == 1
	}

	if !success {
		errorMsg := "未知错误"
		if msg, ok := results[1].(string); ok {
			errorMsg = msg
		}
		return fmt.Errorf(errorMsg)
	}

	return nil
}

// SyncWriteoffBalanceToDB 同步码商余额到数据库（已废弃）
// 注意：码商余额扣减在数据库中进行，不需要从 Redis 同步到数据库
// 缓存刷新服务会定期从数据库同步码商余额到 Redis（只读同步）
// Deprecated: 码商余额扣减在数据库中进行，不需要反向同步
func (s *BalanceService) SyncWriteoffBalanceToDB(ctx context.Context, writeoffID int64) error {
	// 不再同步，码商余额扣减在数据库中进行
	return nil
}
