package models

import (
	"time"
)

// TenantCashFlow 租户流水模型
type TenantCashFlow struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       int64      `gorm:"index;not null;comment:租户ID" json:"tenant_id"`
	OldMoney       int64      `gorm:"not null;comment:变动前金额" json:"old_money"`
	NewMoney       int64      `gorm:"not null;comment:变动后金额" json:"new_money"`
	ChangeMoney    int64      `gorm:"not null;comment:变动金额" json:"change_money"`
	FlowType       int        `gorm:"not null;comment:流水类型(1=消费,2=充值)" json:"flow_type"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	OrderID        *string    `gorm:"index;type:varchar(30);comment:订单ID" json:"order_id,omitempty"`
	PayChannelID   *int64     `gorm:"index;comment:支付通道ID" json:"pay_channel_id,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`

	// 关联关系
	Tenant     *Tenant     `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Order      *Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	PayChannel *PayChannel `gorm:"foreignKey:PayChannelID" json:"pay_channel,omitempty"`
}

// TableName 指定表名
func (TenantCashFlow) TableName() string {
	return "dvadmin_tenant_cash_flow"
}

// FlowType 流水类型常量
const (
	FlowTypeConsume  = 1 // 消费
	FlowTypeRecharge = 2 // 充值
)
