package models

import (
	"time"
)

// Writeoff 核销模型
type Writeoff struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	Modifier       string     `gorm:"type:varchar(255);comment:修改人" json:"modifier,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	Balance        *int64     `gorm:"comment:金额" json:"balance,omitempty"`
	White          string     `gorm:"type:json;default:'[]';comment:白名单" json:"white,omitempty"`
	Telegram       string     `gorm:"type:varchar(255);comment:Telegram群的id" json:"telegram,omitempty"`
	Ver            int64      `gorm:"not null;comment:版本号" json:"ver"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	ParentID       int64      `gorm:"index;not null;comment:上级租户" json:"parent_id"`
	ParentWriteoffID *int64   `gorm:"index;comment:上级核销" json:"parent_writeoff_id,omitempty"`
	SystemUserID   *int64     `gorm:"uniqueIndex;comment:绑定的系统用户" json:"system_user_id,omitempty"`

	// 关联关系（注释掉，避免循环依赖）
	// SystemUser *SystemUser `gorm:"foreignKey:SystemUserID" json:"system_user,omitempty"`
	// Parent     *Tenant     `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

// TableName 指定表名
func (Writeoff) TableName() string {
	return "dvadmin_writeoff"
}

// WriteoffPayChannel 核销支付通道关联表
type WriteoffPayChannel struct {
	WriteoffID   int64 `gorm:"primaryKey;comment:核销" json:"writeoff_id"`
	PayChannelID int64 `gorm:"primaryKey;comment:支付通道" json:"pay_channel_id"`
	Status       bool  `gorm:"not null;default:1;comment:状态" json:"status"`
}

// TableName 指定表名
func (WriteoffPayChannel) TableName() string {
	return "dvadmin_writeoff_pay_channel"
}

