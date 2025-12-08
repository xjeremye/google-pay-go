package models

import (
	"time"
)

// AlipayProductDay 支付宝产品日统计模型
type AlipayProductDay struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SuccessCount int       `gorm:"not null;default:0;comment:成功订单数" json:"success_count"`
	SubmitCount  int       `gorm:"not null;default:0;comment:总提交订单数" json:"submit_count"`
	SuccessMoney int64     `gorm:"not null;default:0;comment:总收入" json:"success_money"`
	Date         time.Time `gorm:"type:date;not null;comment:日期" json:"date"`
	Ver          int64     `gorm:"not null;comment:版本号" json:"ver"`
	ProductID    *int64    `gorm:"index;comment:关联项目" json:"product_id,omitempty"`
	PayChannelID *int64    `gorm:"index;comment:关联通道" json:"pay_channel_id,omitempty"`
}

// TableName 指定表名
func (AlipayProductDay) TableName() string {
	return "dvadmin_alipay_product_day"
}

// AlipayPublicPoolDay 支付宝公池日统计模型
type AlipayPublicPoolDay struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SuccessCount int       `gorm:"not null;default:0;comment:成功订单数" json:"success_count"`
	SubmitCount  int       `gorm:"not null;default:0;comment:总提交订单数" json:"submit_count"`
	SuccessMoney int64     `gorm:"not null;default:0;comment:总收入" json:"success_money"`
	Date         time.Time `gorm:"type:date;not null;comment:日期" json:"date"`
	Ver          int64     `gorm:"not null;comment:版本号" json:"ver"`
	PoolID       *int64    `gorm:"index;comment:关联项目" json:"pool_id,omitempty"`
	PayChannelID *int64    `gorm:"index;comment:关联通道" json:"pay_channel_id,omitempty"`
}

// TableName 指定表名
func (AlipayPublicPoolDay) TableName() string {
	return "dvadmin_alipay_public_pool_day"
}

// AlipayShenmaDay 支付宝神码日统计模型
type AlipayShenmaDay struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SuccessCount int       `gorm:"not null;default:0;comment:成功订单数" json:"success_count"`
	SubmitCount  int       `gorm:"not null;default:0;comment:总提交订单数" json:"submit_count"`
	SuccessMoney int64     `gorm:"not null;default:0;comment:总收入" json:"success_money"`
	Date         time.Time `gorm:"type:date;not null;comment:日期" json:"date"`
	Ver          int64     `gorm:"not null;comment:版本号" json:"ver"`
	ShenmaID     *int64    `gorm:"index;comment:关联项目" json:"shenma_id,omitempty"`
	PayChannelID *int64    `gorm:"index;comment:关联通道" json:"pay_channel_id,omitempty"`
}

// TableName 指定表名
func (AlipayShenmaDay) TableName() string {
	return "dvadmin_alipay_shenma_day"
}
