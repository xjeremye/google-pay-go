package models

import (
	"time"
)

// PayPlugin 支付插件模型
type PayPlugin struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string     `gorm:"uniqueIndex;type:varchar(64);not null;comment:支付插件名称" json:"name"`
	Description    string     `gorm:"type:longtext;not null;comment:支付插件描述" json:"description"`
	Status         bool       `gorm:"not null;comment:支付插件状态" json:"status"`
	CanDivide      bool       `gorm:"not null;comment:是否可以分账" json:"can_divide"`
	CanTransfer    bool       `gorm:"not null;comment:是否可以转账" json:"can_transfer"`
	SupportDevice  int        `gorm:"not null;comment:支持设备" json:"support_device"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
}

// TableName 指定表名
func (PayPlugin) TableName() string {
	return "dvadmin_pay_plugin"
}

// PayPluginConfig 插件配置模型
type PayPluginConfig struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title          string     `gorm:"type:varchar(50);not null;comment:标题" json:"title"`
	Key            string     `gorm:"type:varchar(20);not null;comment:关键字" json:"key"`
	Value          string     `gorm:"type:json;comment:值" json:"value"`
	Sort           int        `gorm:"not null;comment:排序" json:"sort"`
	Status         bool       `gorm:"not null;comment:启用状态" json:"status"`
	DataOptions    string     `gorm:"type:json;comment:数据options" json:"data_options,omitempty"`
	FormItemType   int        `gorm:"not null;comment:表单类型" json:"form_item_type"`
	Rule           string     `gorm:"type:json;comment:校验规则" json:"rule,omitempty"`
	Placeholder    string     `gorm:"type:varchar(50);comment:提示信息" json:"placeholder,omitempty"`
	Setting        string     `gorm:"type:json;comment:配置" json:"setting,omitempty"`
	ParentID       int64      `gorm:"index;not null;comment:关联插件" json:"parent_id"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
}

// TableName 指定表名
func (PayPluginConfig) TableName() string {
	return "dvadmin_pay_plugin_config"
}

// PayType 支付类型模型
type PayType struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string     `gorm:"uniqueIndex;type:varchar(64);not null;comment:支付方式名称" json:"name"`
	Key            string     `gorm:"uniqueIndex;type:varchar(64);not null;comment:支付方式关键字" json:"key"`
	Status         bool       `gorm:"not null;default:1;comment:支付方式状态" json:"status"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
}

// TableName 指定表名
func (PayType) TableName() string {
	return "dvadmin_pay_type"
}

// PayPluginPayType 插件支付类型关联
type PayPluginPayType struct {
	ID          int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	PayPluginID int64 `gorm:"index;not null;comment:插件ID" json:"pay_plugin_id"`
	PayTypeID   int64 `gorm:"index;not null;comment:支付类型ID" json:"pay_type_id"`
}

// TableName 指定表名
func (PayPluginPayType) TableName() string {
	return "dvadmin_pay_plugin_pay_types"
}
