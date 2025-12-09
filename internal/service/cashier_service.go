package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CashierService 收银台服务
type CashierService struct {
	orderService *OrderService
}

// NewCashierService 创建收银台服务
func NewCashierService() *CashierService {
	return &CashierService{
		orderService: NewOrderService(),
	}
}

// RecordCashierVisit 记录用户访问收银台
// 收集用户信息（IP、设备指纹、设备类型等）并保存到订单设备详情表
// 参考 Python: 用户进入收银台时记录设备信息
// 优化：只查询必要的字段，减少数据库查询压力
func (s *CashierService) RecordCashierVisit(ctx context.Context, orderNo string, clientIP string, userAgent string, deviceFingerprint string, userID string) error {
	// 查询订单（只查询必要的字段：id 和 order_status）
	var order struct {
		ID          string `gorm:"column:id"`
		OrderStatus int    `gorm:"column:order_status"`
	}
	if err := database.DB.Table("dvadmin_order").
		Select("id, order_status").
		Where("order_no = ?", orderNo).
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("订单不存在: %s", orderNo)
		}
		return fmt.Errorf("查询订单失败: %w", err)
	}

	// 检查订单状态（只有生成中、等待支付和支付中状态的订单才记录访问）
	if order.OrderStatus != models.OrderStatusGenerating && order.OrderStatus != models.OrderStatusPaying {
		// 订单已处理，不需要记录访问
		return nil
	}

	// 检测设备类型
	deviceType := utils.DetectDeviceType(userAgent)

	// 解析 IP 地址归属地（异步查询，避免阻塞）
	ipLocation, err := utils.GetIPLocation(clientIP)
	if err != nil {
		logger.Logger.Warn("解析 IP 归属地失败",
			zap.String("order_no", orderNo),
			zap.String("ip", clientIP),
			zap.Error(err))
		// 失败时使用默认值
		ipLocation = &utils.IPLocationInfo{
			Address: "",
			PID:     -1,
			CID:     -1,
		}
	} else {
		logger.Logger.Debug("解析 IP 归属地成功",
			zap.String("order_no", orderNo),
			zap.String("ip", clientIP),
			zap.String("address", ipLocation.Address),
			zap.Int("pid", ipLocation.PID),
			zap.Int("cid", ipLocation.CID))
	}

	now := time.Now()

	// 优化：使用 MySQL 的 INSERT ... ON DUPLICATE KEY UPDATE 实现 UPSERT
	// 减少一次查询，提高性能（order_id 有唯一索引）
	// 如果记录已存在，更新；如果不存在，创建
	// 归属地信息更新逻辑：
	// 1. 如果 IP 地址变化，更新归属地信息
	// 2. 如果归属地信息为空（之前查询失败），现在有值了，更新归属地信息
	// 3. 如果归属地信息已经有值，且 IP 地址没变，不更新（保持原值）
	sql := `INSERT INTO dvadmin_order_device_detail (order_id, ip_address, device_type, device_fingerprint, user_id, create_datetime, update_datetime, address, pid, cid)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				ip_address = VALUES(ip_address),
				device_type = VALUES(device_type),
				device_fingerprint = IF(VALUES(device_fingerprint) != '', VALUES(device_fingerprint), device_fingerprint),
				user_id = IF(VALUES(user_id) != '', VALUES(user_id), user_id),
				address = IF(
					ip_address != VALUES(ip_address) OR 
					(address = '' OR address IS NULL) AND VALUES(address) != '',
					VALUES(address),
					address
				),
				pid = IF(
					ip_address != VALUES(ip_address) OR 
					(pid = -1 AND VALUES(pid) != -1),
					VALUES(pid),
					pid
				),
				cid = IF(
					ip_address != VALUES(ip_address) OR 
					(cid = -1 AND VALUES(cid) != -1),
					VALUES(cid),
					cid
				),
				update_datetime = VALUES(update_datetime)`

	result := database.DB.Exec(sql,
		order.ID,
		clientIP,
		deviceType,
		deviceFingerprint,
		userID,
		&now,
		&now,
		ipLocation.Address, // address
		ipLocation.PID,     // pid
		ipLocation.CID,     // cid
	)
	if result.Error != nil {
		logger.Logger.Warn("创建/更新订单设备详情失败",
			zap.String("order_no", orderNo),
			zap.String("order_id", order.ID),
			zap.Error(result.Error))
		return fmt.Errorf("创建/更新订单设备详情失败: %w", result.Error)
	}

	logger.Logger.Debug("订单设备详情已保存",
		zap.String("order_no", orderNo),
		zap.String("order_id", order.ID),
		zap.String("ip", clientIP),
		zap.String("address", ipLocation.Address),
		zap.Int("pid", ipLocation.PID),
		zap.Int("cid", ipLocation.CID),
		zap.Int64("rows_affected", result.RowsAffected))

	return nil
}

// GetOrderDeviceDetail 获取订单设备详情
func (s *CashierService) GetOrderDeviceDetail(ctx context.Context, orderNo string) (*models.OrderDeviceDetail, error) {
	// 查询订单
	var order models.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, fmt.Errorf("订单不存在: %w", err)
	}

	// 查询设备详情
	var deviceDetail models.OrderDeviceDetail
	if err := database.DB.Where("order_id = ?", order.ID).First(&deviceDetail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 设备详情不存在，返回 nil
		}
		return nil, fmt.Errorf("查询订单设备详情失败: %w", err)
	}

	return &deviceDetail, nil
}
