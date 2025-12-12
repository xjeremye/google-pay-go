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
// 根据 Python 项目 backend/dvadmin/agent_manager/models/cash.py 保持一致
const (
	// 租户流水类型 (TenantCashFlow)
	TenantCashflowTypeConsume       = 1 // 消费（订单手续费）
	TenantCashflowTypeRecharge      = 2 // 充值
	TenantCashflowTypeOtherRecharge = 3 // 其他充值

	// 核销流水类型 (WriteoffCashFlow)
	WriteoffCashflowTypeRunVolume = 1 // 跑量（订单扣减）
	WriteoffCashflowTypeSubProfit = 7 // 下级收益（上级核销获得）
	WriteoffCashflowTypeRefund    = 8 // 订单退款

	// 向后兼容的别名（保持现有代码可用）
	CashflowTypeOrderDeduct = WriteoffCashflowTypeRunVolume // 订单扣减（核销流水）
	CashflowTypeOrderRefund = WriteoffCashflowTypeRefund    // 订单退款（核销流水）
	CashflowTypeRecharge    = TenantCashflowTypeRecharge    // 充值（租户流水）
	CashflowTypeSubProfit   = WriteoffCashflowTypeSubProfit // 下级收益（核销流水）
)
