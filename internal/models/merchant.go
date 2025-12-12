package models

import (
	"time"
)

// Merchant 商户模型
type Merchant struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Telegram       string     `gorm:"type:varchar(255);comment:Telegram群的id" json:"telegram,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	SystemUserID   *int64     `gorm:"uniqueIndex;comment:绑定的系统用户" json:"system_user_id,omitempty"`
	ParentID       int64      `gorm:"index;not null;comment:上级租户" json:"parent_id"`

	// 关联关系
	Parent      *Tenant              `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Orders      []Order              `gorm:"foreignKey:MerchantID" json:"orders,omitempty"`
	PayChannels []MerchantPayChannel `gorm:"foreignKey:MerchantID" json:"pay_channels,omitempty"`
}

// TableName 指定表名
func (Merchant) TableName() string {
	return "dvadmin_merchant"
}

// MerchantPayChannel 商户支付通道关联
type MerchantPayChannel struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID     int64      `gorm:"index;not null;comment:商户ID" json:"merchant_id"`
	PayChannelID   int64      `gorm:"index;not null;comment:支付通道ID" json:"pay_channel_id"`
	Status         int        `gorm:"not null;default:1;comment:状态" json:"status"`
	Tax            float64    `gorm:"type:decimal(5,2);not null;default:0.00;comment:费率(百分比)" json:"tax"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`

	// 关联关系
	Merchant   *Merchant   `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	PayChannel *PayChannel `gorm:"foreignKey:PayChannelID" json:"pay_channel,omitempty"`
}

// TableName 指定表名
func (MerchantPayChannel) TableName() string {
	return "dvadmin_merchant_pay_channel"
}

// Tenant 租户模型
type Tenant struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	Balance        int64      `gorm:"not null;default:0;comment:金额" json:"balance"`
	PreTax         int        `gorm:"not null;default:0;comment:占用金额" json:"pre_tax"`
	Trust          bool       `gorm:"not null;default:0;comment:允许负数拉单" json:"trust"`
	SystemUserID   *int64     `gorm:"uniqueIndex;comment:绑定的系统用户" json:"system_user_id,omitempty"`

	// 关联关系
	Merchants []Merchant `gorm:"foreignKey:ParentID" json:"merchants,omitempty"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "dvadmin_tenant"
}

// MerchantPre 商户预付款模型
// 根据文档：预付款用于统计商户的预付款金额（不是余额）
// 预付款表示商户已使用的金额，订单成功时减少，订单退款时增加
type MerchantPre struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantID     int64      `gorm:"uniqueIndex;not null;comment:关联商户" json:"merchant_id"`
	PrePay         int64      `gorm:"not null;default:0;comment:预付金额" json:"pre_pay"`
	Ver            int64      `gorm:"not null;comment:版本号" json:"ver"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`

	// 关联关系
	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

// TableName 指定表名
func (MerchantPre) TableName() string {
	return "dvadmin_merchant_pre"
}
