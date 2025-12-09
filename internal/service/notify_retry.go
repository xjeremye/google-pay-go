package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"go.uber.org/zap"
)

// NotifyRetryService 通知重试服务
// 定时扫描失败的通知并重试
type NotifyRetryService struct {
	notifyService *OrderNotifyService
	stopChan      chan struct{}
}

// NewNotifyRetryService 创建通知重试服务
func NewNotifyRetryService() *NotifyRetryService {
	return &NotifyRetryService{
		notifyService: NewOrderNotifyService(),
		stopChan:      make(chan struct{}),
	}
}

// Start 启动重试服务
// 参考 Python: 定时任务扫描失败的通知并重试
func (s *NotifyRetryService) Start(ctx context.Context) {
	// 每30秒执行一次重试检查
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	logger.Logger.Info("通知重试服务已启动")

	// 立即执行一次
	s.retryFailedNotifications(ctx)

	for {
		select {
		case <-ticker.C:
			// 定时执行重试
			s.retryFailedNotifications(ctx)
		case <-s.stopChan:
			logger.Logger.Info("通知重试服务已停止")
			return
		case <-ctx.Done():
			logger.Logger.Info("通知重试服务已停止（上下文取消）")
			return
		}
	}
}

// Stop 停止重试服务
func (s *NotifyRetryService) Stop() {
	close(s.stopChan)
}

// retryFailedNotifications 执行重试逻辑
// 使用分布式锁确保只有一个实例执行重试任务
func (s *NotifyRetryService) retryFailedNotifications(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("通知重试服务异常",
				zap.Any("panic", r))
		}
	}()

	// 使用分布式锁，确保只有一个实例执行重试任务
	lockKey := "lock:notify_retry"
	lockTTL := 25 * time.Second // 锁的过期时间（略小于定时任务间隔30秒）

	// 尝试获取分布式锁
	acquired, err := acquireDistributedLock(ctx, lockKey, lockTTL)
	if err != nil {
		// Redis 不可用时，记录警告但继续执行（容错处理）
		logger.Logger.Warn("获取分布式锁失败，继续执行（Redis可能不可用）",
			zap.String("lock_key", lockKey),
			zap.Error(err))
		// 如果 Redis 不可用，仍然执行任务（单机模式）
		s.notifyService.RetryFailedNotifications(ctx)
		return
	}

	if !acquired {
		// 锁已被其他实例获取，跳过此次执行
		logger.Logger.Debug("通知重试任务正在其他实例执行，跳过此次执行",
			zap.String("lock_key", lockKey))
		return
	}

	// 成功获取锁，执行重试任务
	defer func() {
		// 释放锁（即使任务执行失败也要释放）
		if err := releaseDistributedLock(ctx, lockKey); err != nil {
			logger.Logger.Warn("释放分布式锁失败",
				zap.String("lock_key", lockKey),
				zap.Error(err))
		}
	}()

	logger.Logger.Debug("成功获取分布式锁，开始执行通知重试任务",
		zap.String("lock_key", lockKey))
	s.notifyService.RetryFailedNotifications(ctx)
}

// acquireDistributedLock 获取分布式锁（使用 Redis SET NX EX）
// 返回是否成功获取锁
func acquireDistributedLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if database.RDB == nil {
		return false, fmt.Errorf("Redis 未初始化")
	}

	// 使用 SET NX EX 命令原子性地设置锁
	// NX: 只有当 key 不存在时才设置
	// EX: 设置过期时间（秒）
	result, err := database.RDB.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("获取分布式锁失败: %w", err)
	}

	return result, nil
}

// releaseDistributedLock 释放分布式锁
func releaseDistributedLock(ctx context.Context, key string) error {
	if database.RDB == nil {
		return fmt.Errorf("Redis 未初始化")
	}

	// 删除锁
	_, err := database.RDB.Del(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("释放分布式锁失败: %w", err)
	}

	return nil
}
