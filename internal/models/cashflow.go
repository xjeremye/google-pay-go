package models

import (
	"time"
)

// TenantCashflow 租户资金流水模型
type TenantCashflow struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	Modifier       string     `gorm:"type:varchar(255);comment:修改人" json:"modifier,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	OldMoney       int64      `gorm:"not null;comment:变更前余额" json:"old_money"`
	NewMoney       int64      `gorm:"not null;comment:变更后余额" json:"new_money"`
	ChangeMoney    int64      `gorm:"not null;comment:变更余额" json:"change_money"`
	FlowType       int        `gorm:"not null;default:0;comment:流水类型" json:"flow_type"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	OrderID        *string    `gorm:"index;type:varchar(30);comment:系统订单" json:"order_id,omitempty"`
	PayChannelID   *int64     `gorm:"index;comment:支付通道" json:"pay_channel_id,omitempty"`
	TenantID       int64      `gorm:"index;not null;comment:关联租户" json:"tenant_id"`
}

// TableName 指定表名
func (TenantCashflow) TableName() string {
	return "dvadmin_tenant_cashflow"
}

// WriteoffCashflow 核销（码商）资金流水模型
type WriteoffCashflow struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Remarks        string     `gorm:"type:varchar(255);comment:备注" json:"remarks,omitempty"`
	Modifier       string     `gorm:"type:varchar(255);comment:修改人" json:"modifier,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	OldMoney       int64      `gorm:"not null;comment:变更前余额" json:"old_money"`
	NewMoney       int64      `gorm:"not null;comment:变更后余额" json:"new_money"`
	ChangeMoney    int64      `gorm:"not null;comment:变更余额" json:"change_money"`
	FlowType       int        `gorm:"not null;default:0;comment:流水类型" json:"flow_type"`
	Tax            float64    `gorm:"type:decimal(5,2);not null;default:0.00;comment:费率" json:"tax"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	OrderID        *string    `gorm:"index;type:varchar(30);comment:系统订单" json:"order_id,omitempty"`
	PayChannelID   *int64     `gorm:"index;comment:支付通道" json:"pay_channel_id,omitempty"`
	WriteoffID     int64      `gorm:"index;not null;comment:关联核销" json:"writeoff_id"`
}

// TableName 指定表名
func (WriteoffCashflow) TableName() string {
	return "dvadmin_writeoff_cashflow"
}

// CashflowType 流水类型常量
const (
	CashflowTypeOrderDeduct = 1 // 订单扣减
	CashflowTypeOrderRefund = 2 // 订单退款
	CashflowTypeRecharge    = 3 // 充值
	CashflowTypeWithdraw    = 4 // 提现
)
