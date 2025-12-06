package service

import (
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"gorm.io/gorm"
)

type PayChannelService struct{}

// NewPayChannelService 创建支付通道服务
func NewPayChannelService() *PayChannelService {
	return &PayChannelService{}
}

// GetPayChannelByID 根据ID获取支付通道
func (s *PayChannelService) GetPayChannelByID(id int64) (*models.PayChannel, error) {
	var payChannel models.PayChannel
	err := database.DB.First(&payChannel, id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("支付通道不存在")
		}
		return nil, fmt.Errorf("查询支付通道失败: %w", err)
	}

	return &payChannel, nil
}

// GetAvailablePayChannels 获取可用的支付通道列表
func (s *PayChannelService) GetAvailablePayChannels() ([]models.PayChannel, error) {
	var payChannels []models.PayChannel
	
	now := time.Now()
	currentTime := now.Format("150405") // HHMMSS
	
	err := database.DB.Where("status = ?", true).
		Where("start_time <= ? AND end_time >= ?", currentTime, currentTime).
		Find(&payChannels).Error
	
	if err != nil {
		return nil, fmt.Errorf("查询支付通道失败: %w", err)
	}

	return payChannels, nil
}

// ValidatePayChannel 验证支付通道是否可用
func (s *PayChannelService) ValidatePayChannel(payChannelID int64, amount int) error {
	payChannel, err := s.GetPayChannelByID(payChannelID)
	if err != nil {
		return err
	}

	if !payChannel.Status {
		return fmt.Errorf("支付通道已禁用")
	}

	// 检查时间范围
	now := time.Now()
	currentTime := now.Format("150405")
	if payChannel.StartTime > currentTime || payChannel.EndTime < currentTime {
		return fmt.Errorf("支付通道不在服务时间内")
	}

	// 检查金额范围
	if amount < payChannel.MinMoney || amount > payChannel.MaxMoney {
		return fmt.Errorf("金额超出支付通道限制范围")
	}

	return nil
}

