package models

import (
	"time"
)

// Order 订单模型
type Order struct {
	ID             string     `gorm:"primaryKey;type:varchar(30);comment:订单Id" json:"id"`
	OrderNo        string     `gorm:"uniqueIndex;type:varchar(32);not null;comment:本系统订单号" json:"order_no"`
	OutOrderNo     string     `gorm:"uniqueIndex;type:varchar(32);not null;comment:商户订单号" json:"out_order_no"`
	OrderStatus    int        `gorm:"index;not null;comment:订单状态" json:"order_status"`
	Money          int        `gorm:"not null;comment:金额(分)" json:"money"`
	Tax            int        `gorm:"not null;comment:手续费(分)" json:"tax"`
	PayDatetime    *time.Time `gorm:"comment:支付时间" json:"pay_datetime,omitempty"`
	ProductName    string     `gorm:"type:varchar(255);comment:通道名称" json:"product_name,omitempty"`
	ReqExtra       string     `gorm:"type:longtext;comment:额外请求参数" json:"req_extra,omitempty"`
	CreateDatetime *time.Time `gorm:"index;comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	Compatible     int        `gorm:"default:0;comment:系统兼容" json:"compatible"`
	Ver            int64      `gorm:"not null;comment:版本号" json:"ver"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	MerchantID     *int64     `gorm:"index;comment:关联商户" json:"merchant_id,omitempty"`
	WriteoffID     *int64     `gorm:"index;comment:核销" json:"writeoff_id,omitempty"`
	PayChannelID   *int64     `gorm:"index;comment:关联支付通道" json:"pay_channel_id,omitempty"`

	// 关联关系
	Merchant    *Merchant    `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	PayChannel  *PayChannel  `gorm:"foreignKey:PayChannelID" json:"pay_channel,omitempty"`
	OrderDetail *OrderDetail `gorm:"foreignKey:OrderID" json:"order_detail,omitempty"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "dvadmin_order"
}

// OrderDetail 订单详情模型
type OrderDetail struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID        string     `gorm:"uniqueIndex;type:varchar(30);not null;comment:关联订单" json:"order_id"`
	NotifyURL      string     `gorm:"type:longtext;not null;comment:通知地址" json:"notify_url"`
	JumpURL        string     `gorm:"type:longtext;not null;comment:跳转地址" json:"jump_url"`
	ProductID      string     `gorm:"index;type:varchar(255);comment:商品" json:"product_id,omitempty"`
	CookieID       string     `gorm:"type:varchar(255);comment:小号" json:"cookie_id,omitempty"`
	NotifyMoney    int        `gorm:"not null;comment:通知金额(分)" json:"notify_money"`
	TicketNo       string     `gorm:"type:varchar(255);comment:官方流水号" json:"ticket_no,omitempty"`
	QueryNo        string     `gorm:"type:varchar(255);comment:查询订单号" json:"query_no,omitempty"`
	PluginType     string     `gorm:"type:varchar(255);comment:插件支付类型" json:"plugin_type,omitempty"`
	PluginUpstream int        `gorm:"default:-1;comment:插件大类" json:"plugin_upstream"`
	MerchantTax    int        `gorm:"default:0;comment:商户手续费(分)" json:"merchant_tax"`
	Extra          string     `gorm:"type:json;comment:额外数据" json:"extra,omitempty"`
	Remarks        string     `gorm:"type:longtext;comment:备注" json:"remarks,omitempty"`
	BuyerID        string     `gorm:"type:varchar(255);comment:买家ID" json:"buyer_id,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
	CreatorID      *int64     `gorm:"index;comment:创建人" json:"creator_id,omitempty"`
	WriteoffID     *int64     `gorm:"index;comment:核销" json:"writeoff_id,omitempty"`
	DomainID       *int64     `gorm:"index;comment:域名" json:"domain_id,omitempty"`
	PluginID       *int64     `gorm:"index;comment:插件" json:"plugin_id,omitempty"`
}

// TableName 指定表名
func (OrderDetail) TableName() string {
	return "dvadmin_order_detail"
}

// OrderStatus 订单状态常量
const (
	OrderStatusPending   = 0 // 待支付
	OrderStatusPaid      = 1 // 已支付
	OrderStatusFailed    = 2 // 支付失败
	OrderStatusCancelled = 3 // 已取消
	OrderStatusExpired   = 4 // 已过期
)
