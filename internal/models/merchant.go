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
