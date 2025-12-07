package models

import (
	"time"
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title          string     `gorm:"type:varchar(50);not null;comment:标题" json:"title"`
	Key            string     `gorm:"type:varchar(20);not null;comment:键" json:"key"`
	Value          string     `gorm:"type:json;comment:值" json:"value"`
	Sort           int        `gorm:"not null;comment:排序" json:"sort"`
	Status         bool       `gorm:"not null;comment:启用状态" json:"status"`
	DataOptions    string     `gorm:"type:json;comment:数据options" json:"data_options,omitempty"`
	FormItemType   int        `gorm:"not null;comment:表单类型" json:"form_item_type"`
	Rule           string     `gorm:"type:json;comment:校验规则" json:"rule,omitempty"`
	Placeholder    string     `gorm:"type:varchar(50);comment:提示信息" json:"placeholder,omitempty"`
	Setting        string     `gorm:"type:json;comment:配置" json:"setting,omitempty"`
	ParentID       *int64     `gorm:"index;comment:父级" json:"parent_id,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "dvadmin_system_config"
}
