package service

import (
	"context"
	"time"

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
func (s *NotifyRetryService) retryFailedNotifications(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("通知重试服务异常",
				zap.Any("panic", r))
		}
	}()

	s.notifyService.RetryFailedNotifications(ctx)
}
