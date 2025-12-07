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

// OrderDeviceDetail 订单设备详情模型
type OrderDeviceDetail struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID           string     `gorm:"uniqueIndex;type:varchar(30);not null;comment:关联订单" json:"order_id"`
	IPAddress         string     `gorm:"type:varchar(255);index;comment:Ip地址" json:"ip_address"`
	Address           string     `gorm:"type:varchar(32);comment:归属地" json:"address,omitempty"`
	DeviceType        int        `gorm:"default:0;index;comment:设备类型" json:"device_type"`
	DeviceFingerprint string     `gorm:"type:varchar(255);comment:设备指纹" json:"device_fingerprint,omitempty"`
	PID               int        `gorm:"default:-1;comment:代理省ip" json:"pid"`
	CID               int        `gorm:"default:-1;comment:代理城市ip" json:"cid"`
	UserID            string     `gorm:"type:varchar(32);index;comment:用户id" json:"user_id,omitempty"`
	CreateDatetime    *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime    *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
}

// TableName 指定表名
func (OrderDeviceDetail) TableName() string {
	return "dvadmin_order_device_detail"
}

// DeviceType 设备类型常量
const (
	DeviceTypeUnknown = 0 // 未知设备
	DeviceTypeAndroid = 1 // Android
	DeviceTypeIOS     = 2 // IOS
	DeviceTypePC      = 4 // PC
)

// OrderLog 订单日志模型
type OrderLog struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OutOrderNo     string     `gorm:"uniqueIndex;type:varchar(32);not null;comment:外部订单号" json:"out_order_no"`
	SignRaw        string     `gorm:"type:longtext;comment:签名原始数据" json:"sign_raw,omitempty"`
	Sign           string     `gorm:"type:varchar(32);comment:签名数据" json:"sign,omitempty"`
	RequestBody    string     `gorm:"type:longtext;comment:请求参数" json:"request_body,omitempty"`
	RequestMethod  string     `gorm:"type:varchar(8);comment:请求方式" json:"request_method,omitempty"`
	ResponseCode   string     `gorm:"type:varchar(32);comment:响应状态码" json:"response_code,omitempty"`
	JSONResult     string     `gorm:"type:longtext;comment:返回信息" json:"json_result,omitempty"`
	CreateDatetime *time.Time `gorm:"comment:创建时间" json:"create_datetime,omitempty"`
	UpdateDatetime *time.Time `gorm:"comment:修改时间" json:"update_datetime,omitempty"`
}

// TableName 指定表名
func (OrderLog) TableName() string {
	return "dvadmin_order_log"
}
