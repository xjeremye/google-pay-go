package mq

// CallbackSubmitMessage 插件回调提交消息
type CallbackSubmitMessage struct {
	OrderNo        string `json:"order_no"`
	OutOrderNo     string `json:"out_order_no"`
	PluginID       int64  `json:"plugin_id"`
	Tax            int    `json:"tax"`
	PluginType     string `json:"plugin_type"`
	Money          int    `json:"money"`
	DomainID       *int64 `json:"domain_id"`
	NotifyMoney    int    `json:"notify_money"`
	OrderID        string `json:"order_id"`
	ProductID      string `json:"product_id"`
	CookieID       string `json:"cookie_id"`
	ChannelID      int64  `json:"channel_id"`
	MerchantID     int64  `json:"merchant_id"`
	WriteoffID     *int64 `json:"writeoff_id"`
	TenantID       int64  `json:"tenant_id"`
	CreateDatetime string `json:"create_datetime"`
	NotifyURL      string `json:"notify_url"`
	PluginUpstream int    `json:"plugin_upstream"`
}

// OrderNotifyMessage 订单通知消息
type OrderNotifyMessage struct {
	OrderID    string `json:"order_id"`
	OrderNo    string `json:"order_no"`
	OutOrderNo string `json:"out_order_no"`
	Money      int    `json:"money"`
	Status     int    `json:"status"`
	TicketNo   string `json:"ticket_no"`
	NotifyURL  string `json:"notify_url"`
	Timestamp  int64  `json:"timestamp"`
	RetryCount int    `json:"retry_count"` // 重试次数
}

// DayStatisticsMessage 日统计数据更新消息
type DayStatisticsMessage struct {
	ProductID      string `json:"product_id"`
	ChannelID      int64  `json:"channel_id"`
	TenantID       int64  `json:"tenant_id"`
	WriteoffID     *int64 `json:"writeoff_id"`
	Money          int    `json:"money"`
	Date           string `json:"date"`            // 日期格式：2006-01-02
	StatisticsType string `json:"statistics_type"` // submit, success
	ExtraArg       int    `json:"extra_arg"`       // 通道的 extra_arg（用于判断公池/神码模式）
}

// AlipayNotifyMessage 支付宝回调消息
type AlipayNotifyMessage struct {
	PluginType string            `json:"plugin_type"` // 插件类型（如 alipay_phone）
	ProductID  string            `json:"product_id"`  // 产品ID
	Params     map[string]string `json:"params"`      // 回调参数（包含签名，用于验证）
	NotifyData *AlipayNotifyData `json:"notify_data"` // 解析后的回调数据
	ReceivedAt string            `json:"received_at"` // 接收时间（格式：2006-01-02 15:04:05）
}

// AlipayNotifyData 支付宝回调数据（解析后）
type AlipayNotifyData struct {
	OutTradeNo    string `json:"out_trade_no"`   // 商户订单号
	TradeNo       string `json:"trade_no"`       // 支付宝交易号
	TradeStatus   string `json:"trade_status"`   // 交易状态
	TotalAmount   int    `json:"total_amount"`   // 交易金额（分）
	ReceiptAmount int    `json:"receipt_amount"` // 实收金额（分）
	BuyerID       string `json:"buyer_id"`       // 买家支付宝用户ID
	BuyerLogonID  string `json:"buyer_logon_id"` // 买家支付宝账号
	SellerID      string `json:"seller_id"`      // 卖家支付宝用户ID
	GmtPayment    string `json:"gmt_payment"`    // 支付时间
}

// CacheRefreshMessage 缓存刷新触发消息（替代定时器）
type CacheRefreshMessage struct {
	Full        bool     `json:"full"`         // 是否全量刷新
	Targets     []string `json:"targets"`      // 目标列表，见 cache_refresh CacheTarget* 常量
	TenantIDs   []int64  `json:"tenant_ids"`   // 精确刷新租户余额
	WriteoffIDs []int64  `json:"writeoff_ids"` // 精确刷新码商余额
}

// BalanceSyncMessage 后台调整余额后触发的稳定同步消息
type BalanceSyncMessage struct {
	TenantIDs   []int64 `json:"tenant_ids"`   // 需要刷新余额/信任标志的租户
	WriteoffIDs []int64 `json:"writeoff_ids"` // 需要刷新余额的码商
	Full        bool    `json:"full"`         // 是否强制全量刷新余额
}
