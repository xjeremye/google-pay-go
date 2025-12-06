package service

import (
	"fmt"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"gorm.io/gorm"
)

type MerchantService struct{}

// NewMerchantService 创建商户服务
func NewMerchantService() *MerchantService {
	return &MerchantService{}
}

// GetMerchantByID 根据ID获取商户
func (s *MerchantService) GetMerchantByID(id int64) (*models.Merchant, error) {
	var merchant models.Merchant
	err := database.DB.Preload("Parent").
		First(&merchant, id).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("商户不存在")
		}
		return nil, fmt.Errorf("查询商户失败: %w", err)
	}

	return &merchant, nil
}

// GetMerchantPayChannels 获取商户的支付通道列表
func (s *MerchantService) GetMerchantPayChannels(merchantID int64) ([]models.PayChannel, error) {
	var payChannels []models.PayChannel
	
	err := database.DB.Table("dvadmin_pay_channel").
		Joins("INNER JOIN dvadmin_merchant_pay_channel ON dvadmin_pay_channel.id = dvadmin_merchant_pay_channel.pay_channel_id").
		Where("dvadmin_merchant_pay_channel.merchant_id = ? AND dvadmin_merchant_pay_channel.status = ?", merchantID, 1).
		Where("dvadmin_pay_channel.status = ?", true).
		Find(&payChannels).Error
	
	if err != nil {
		return nil, fmt.Errorf("查询支付通道失败: %w", err)
	}

	return payChannels, nil
}

// ValidateMerchant 验证商户是否存在且有效
func (s *MerchantService) ValidateMerchant(merchantID int64) error {
	merchant, err := s.GetMerchantByID(merchantID)
	if err != nil {
		return err
	}

	if merchant == nil {
		return fmt.Errorf("商户不存在")
	}

	return nil
}

