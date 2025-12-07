package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrderNotifyService 订单回调服务
type OrderNotifyService struct {
	orderService *OrderService
}

// NewOrderNotifyService 创建订单回调服务
func NewOrderNotifyService() *OrderNotifyService {
	return &OrderNotifyService{
		orderService: NewOrderService(),
	}
}

// NotifyMerchant 通知商户（异步执行）
// 参考 Python: 向商户的 notify_url 发送回调通知
// 创建通知任务，如果失败则记录到数据库用于后续重试
func (s *OrderNotifyService) NotifyMerchant(ctx context.Context, order *models.Order, orderDetail *models.OrderDetail) {
	if orderDetail.NotifyURL == "" {
		return
	}

	// 构建通知数据
	notifyData := map[string]interface{}{
		"order_no":     order.OrderNo,
		"out_order_no": order.OutOrderNo,
		"money":        order.Money,
		"status":       order.OrderStatus,
		"ticket_no":    orderDetail.TicketNo,
		"timestamp":    time.Now().Unix(),
	}

	// 序列化为 JSON
	notifyJSON, err := json.Marshal(notifyData)
	if err != nil {
		logger.Logger.Warn("序列化通知数据失败",
			zap.String("order_no", order.OrderNo),
			zap.Error(err))
		return
	}

	// 创建或获取通知任务
	notification, err := s.createOrGetNotification(order.ID)
	if err != nil {
		logger.Logger.Warn("创建通知任务失败",
			zap.String("order_id", order.ID),
			zap.Error(err))
		return
	}

	// 执行通知
	success := s.sendNotification(ctx, notification.ID, orderDetail.NotifyURL, string(notifyJSON))

	// 更新通知状态
	if success {
		// 通知成功，更新状态为成功
		s.updateNotificationStatus(notification.ID, models.NotificationStatusSuccess)
	} else {
		// 通知失败，更新状态为失败（等待重试）
		s.updateNotificationStatus(notification.ID, models.NotificationStatusFailed)
	}
}

// createOrGetNotification 创建或获取通知任务
func (s *OrderNotifyService) createOrGetNotification(orderID string) (*models.MerchantNotification, error) {
	var notification models.MerchantNotification

	// 尝试获取现有通知任务
	if err := database.DB.Where("order_id = ?", orderID).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 不存在，创建新任务
			now := time.Now()
			notification = models.MerchantNotification{
				OrderID:        orderID,
				Status:         models.NotificationStatusPending,
				Ver:            1,
				CreateDatetime: &now,
				UpdateDatetime: &now,
			}
			if err := database.DB.Create(&notification).Error; err != nil {
				return nil, fmt.Errorf("创建通知任务失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("查询通知任务失败: %w", err)
		}
	}

	return &notification, nil
}

// sendNotification 发送通知并记录历史
func (s *OrderNotifyService) sendNotification(ctx context.Context, notificationID int64, notifyURL, requestBody string) bool {
	// 发送 POST 请求到商户的通知地址
	req, err := http.NewRequestWithContext(ctx, "POST", notifyURL, strings.NewReader(requestBody))
	if err != nil {
		logger.Logger.Warn("创建通知请求失败",
			zap.Int64("notification_id", notificationID),
			zap.String("notify_url", notifyURL),
			zap.Error(err))
		// 记录失败历史
		s.recordNotificationHistory(notificationID, notifyURL, "POST", requestBody, 0, fmt.Sprintf("创建请求失败: %v", err))
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(requestBody))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.Warn("发送通知失败",
			zap.Int64("notification_id", notificationID),
			zap.String("notify_url", notifyURL),
			zap.Error(err))
		// 记录失败历史
		s.recordNotificationHistory(notificationID, notifyURL, "POST", requestBody, 0, fmt.Sprintf("请求失败: %v", err))
		return false
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 记录通知历史
	s.recordNotificationHistory(notificationID, notifyURL, "POST", requestBody, resp.StatusCode, bodyStr)

	if resp.StatusCode == http.StatusOK {
		logger.Logger.Info("商户通知成功",
			zap.Int64("notification_id", notificationID),
			zap.String("notify_url", notifyURL),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", bodyStr))
		return true
	} else {
		logger.Logger.Warn("商户通知失败",
			zap.Int64("notification_id", notificationID),
			zap.String("notify_url", notifyURL),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", bodyStr))
		return false
	}
}

// recordNotificationHistory 记录通知历史
func (s *OrderNotifyService) recordNotificationHistory(notificationID int64, url, method, requestBody string, responseCode int, jsonResult string) {
	now := time.Now()
	history := &models.MerchantNotificationHistory{
		NotificationID: notificationID,
		URL:            url,
		RequestMethod:  method,
		RequestBody:    requestBody,
		ResponseCode:   responseCode,
		JSONResult:     jsonResult,
		CreateDatetime: &now,
		UpdateDatetime: &now,
	}

	if err := database.DB.Create(history).Error; err != nil {
		logger.Logger.Warn("记录通知历史失败",
			zap.Int64("notification_id", notificationID),
			zap.Error(err))
	}
}

// updateNotificationStatus 更新通知状态
func (s *OrderNotifyService) updateNotificationStatus(notificationID int64, status int) {
	now := time.Now()
	if err := database.DB.Model(&models.MerchantNotification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"status":          status,
			"update_datetime": &now,
			"ver":             gorm.Expr("ver + ?", 1),
		}).Error; err != nil {
		logger.Logger.Warn("更新通知状态失败",
			zap.Int64("notification_id", notificationID),
			zap.Int("status", status),
			zap.Error(err))
	}
}

// RetryFailedNotifications 重试失败的通知
// 参考 Python: 定时任务扫描失败的通知并重试
// 使用指数退避策略，更灵活且可扩展
func (s *OrderNotifyService) RetryFailedNotifications(ctx context.Context) {
	// 查询需要重试的通知（状态为失败或重试中）

	now := time.Now()
	var notifications []models.MerchantNotification

	// 查询所有失败状态的通知
	if err := database.DB.Where("status IN ?", []int{
		models.NotificationStatusFailed,
		models.NotificationStatusRetrying,
	}).Find(&notifications).Error; err != nil {
		logger.Logger.Warn("查询失败通知失败", zap.Error(err))
		return
	}

	for _, notification := range notifications {
		// 查询订单和订单详情
		var order models.Order
		if err := database.DB.Where("id = ?", notification.OrderID).First(&order).Error; err != nil {
			continue
		}

		var orderDetail models.OrderDetail
		if err := database.DB.Where("order_id = ?", order.ID).First(&orderDetail).Error; err != nil {
			continue
		}

		if orderDetail.NotifyURL == "" {
			continue
		}

		// 计算重试次数（通过历史记录数量）
		var historyCount int64
		database.DB.Model(&models.MerchantNotificationHistory{}).
			Where("notification_id = ?", notification.ID).
			Count(&historyCount)

		// 最大重试次数（5次）
		maxRetries := 5
		if historyCount >= int64(maxRetries) {
			s.updateNotificationStatus(notification.ID, models.NotificationStatusMaxRetry)
			continue
		}

		// 计算下次重试时间
		if notification.UpdateDatetime == nil {
			continue
		}

		// 使用指数退避策略计算重试间隔
		retryInterval := s.calculateRetryInterval(int(historyCount))
		nextRetryTime := notification.UpdateDatetime.Add(retryInterval)
		if now.Before(nextRetryTime) {
			// 还没到重试时间
			continue
		}

		// 执行重试
		logger.Logger.Info("开始重试商户通知",
			zap.Int64("notification_id", notification.ID),
			zap.String("order_id", order.ID),
			zap.Int64("retry_count", historyCount))

		// 更新状态为重试中
		s.updateNotificationStatus(notification.ID, models.NotificationStatusRetrying)

		// 构建通知数据
		notifyData := map[string]interface{}{
			"order_no":     order.OrderNo,
			"out_order_no": order.OutOrderNo,
			"money":        order.Money,
			"status":       order.OrderStatus,
			"ticket_no":    orderDetail.TicketNo,
			"timestamp":    time.Now().Unix(),
		}

		notifyJSON, _ := json.Marshal(notifyData)

		// 发送通知
		success := s.sendNotification(ctx, notification.ID, orderDetail.NotifyURL, string(notifyJSON))

		// 更新状态
		if success {
			s.updateNotificationStatus(notification.ID, models.NotificationStatusSuccess)
		} else {
			s.updateNotificationStatus(notification.ID, models.NotificationStatusFailed)
		}
	}
}

// calculateRetryInterval 计算重试间隔（指数退避策略）
// 使用指数退避算法：base * (2 ^ retryCount)
// 基础间隔为1分钟，每次重试间隔翻倍
func (s *OrderNotifyService) calculateRetryInterval(retryCount int) time.Duration {
	// 基础间隔：1分钟
	baseInterval := 1 * time.Minute

	// 指数退避：base * (2 ^ retryCount)
	multiplier := 1 << uint(retryCount) // 使用位运算计算 2^retryCount

	return baseInterval * time.Duration(multiplier)
}
