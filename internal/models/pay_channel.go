package models

import (
	"time"
)

// PayChannel 支付通道模型
type PayChannel struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string     `gorm:"uniqueIndex;type:varchar(64);not null;comment:通道名称" json:"name"`
	Status         bool       `gorm:"not null;default:1;comment:通道状态" json:"status"`
	MaxMoney       int        `gorm:"not null;comment:单笔最大金额(分)" json:"max_money"`
	MinMoney       int        `gorm:"not null;comment:单笔最小金额(分)" json:"min_money"`
	FloatMaxMoney  int        `gorm:"not null;comment:浮动单笔最大金额(分)" json:"float_max_money"`
	FloatMinMoney  int        `gorm:"not null;comment:浮动单笔最小金额(分)" json:"float_min_money"`
	Settled        bool       `gorm:"not null;default:0;comment:固定金额模式" json:"settled"`
	Moneys         string     `gorm:"type:json;comment:固定金额列表" json:"moneys,omitempty"`
	StartTime      string     `gorm:"type:varchar(8);not null;comment:启用时间" json:"start_time"`
	EndTime        string     `gorm:"type:varchar(8);not null;comment:结束时间" json:"end_time"`
	ExtraArg       *int       `gorm:"comment:额外参数" json:"extra_arg,omitempty"`
	BanIP          string     `gorm:"type:json;comment:封禁IP列表" json:"ban_ip,omitempty"`
	Logo           string     `gorm:"type:longtext;comment:图标" json:"logo,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	PluginID       int64      `gorm:"index;not null;comment:支付插件" json:"plugin_id"`

	// 关联关系
	Orders              []Order              `gorm:"foreignKey:PayChannelID" json:"orders,omitempty"`
	MerchantPayChannels []MerchantPayChannel `gorm:"foreignKey:PayChannelID" json:"merchant_pay_channels,omitempty"`
	ChannelTaxes        []PayChannelTax      `gorm:"foreignKey:PayChannelID" json:"channel_taxes,omitempty"`
}

// TableName 指定表名
func (PayChannel) TableName() string {
	return "dvadmin_pay_channel"
}

// PayChannelTax 支付通道费率（租户级别）
type PayChannelTax struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	PayChannelID   int64      `gorm:"index;not null;comment:支付通道ID" json:"pay_channel_id"`
	TenantID       int64      `gorm:"index;not null;comment:租户ID" json:"tenant_id"`
	Tax            float64    `gorm:"type:decimal(5,2);not null;comment:费率(百分比)" json:"tax"`
	Status         bool       `gorm:"not null;comment:状态" json:"status"`
	Mark           string     `gorm:"uniqueIndex;type:varchar(100);not null;comment:标志(通道id-租户id)" json:"mark"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`

	// 关联关系
	PayChannel *PayChannel `gorm:"foreignKey:PayChannelID" json:"pay_channel,omitempty"`
	Tenant     *Tenant     `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// TableName 指定表名
func (PayChannelTax) TableName() string {
	return "dvadmin_pay_channel_tax"
}

// PayChannelStatus 支付通道状态
const (
	PayChannelStatusDisabled = false // 禁用
	PayChannelStatusEnabled  = true  // 启用
)
