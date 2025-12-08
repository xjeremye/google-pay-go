package plugin

import (
	"context"
)

// OrderContext 订单上下文接口，插件需要从上下文中获取信息
type OrderContext interface {
	GetOutOrderNo() string
	GetNotifyURL() string
	GetMoney() int
	GetJumpURL() string
	GetNotifyMoney() int
	GetExtra() string
	GetCompatible() int
	GetTest() bool
	GetMerchantID() int64
	GetTenantID() int64
	GetChannelID() int64
	GetPluginID() int64
	GetPluginType() string
	GetPluginUpstream() int
	GetDomainID() *int64
	GetDomainURL() string
	GetOrderNo() string
	SetOrderNo(no string)
	SetDomainID(id int64)
	SetDomainURL(url string)
}

// Plugin 插件核心接口
type Plugin interface {
	// CreateOrder 创建订单，返回支付URL
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
	// WaitProduct 等待产品（获取产品ID、核销ID、CookieID等）
	// 参考 Python: BasePluginResponder.wait_product
	WaitProduct(ctx context.Context, req *WaitProductRequest) (*WaitProductResponse, error)
	// CallbackSubmit 下单回调（订单创建成功后调用）
	// 参考 Python: BasePluginResponder.callback_submit
	CallbackSubmit(ctx context.Context, req *CallbackSubmitRequest) error
}

// PluginCapabilities 插件能力接口（可选实现）
type PluginCapabilities interface {
	// CanHandleExtra 是否可以处理额外参数
	CanHandleExtra() bool
	// AutoExtra 是否自动处理额外参数
	AutoExtra() bool
	// ExtraNeedProduct 额外参数是否需要产品
	ExtraNeedProduct() bool
	// ExtraNeedCookie 额外参数是否需要Cookie
	ExtraNeedCookie() bool
	// GetTimeout 获取订单超时时间（秒）
	GetTimeout(ctx context.Context, pluginID int64) int
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	OutOrderNo     string                 `json:"out_order_no"`
	OrderNo        string                 `json:"order_no"`
	OrderID        string                 `json:"order_id"`   // 订单ID（主键）
	DetailID       int64                  `json:"detail_id"`  // 订单详情ID
	ProductID      string                 `json:"product_id"` // 产品ID
	Money          int                    `json:"money"`
	NotifyURL      string                 `json:"notify_url"`
	JumpURL        string                 `json:"jump_url"`
	Extra          string                 `json:"extra"`
	MerchantID     int64                  `json:"merchant_id"`
	TenantID       int64                  `json:"tenant_id"`
	ChannelID      int64                  `json:"channel_id"`
	PluginID       int64                  `json:"plugin_id"`
	PluginType     string                 `json:"plugin_type"`
	PluginUpstream int                    `json:"plugin_upstream"`
	DomainID       *int64                 `json:"domain_id"`
	DomainURL      string                 `json:"domain_url"`
	Domain         map[string]interface{} `json:"domain,omitempty"`
	Channel        map[string]interface{} `json:"channel,omitempty"`
	Plugin         map[string]interface{} `json:"plugin,omitempty"`
	PayType        map[string]interface{} `json:"pay_type,omitempty"`
	Compatible     int                    `json:"compatible"`
	Test           bool                   `json:"test"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	Success      bool                   `json:"success"`
	PayURL       string                 `json:"pay_url,omitempty"`
	ErrorCode    int                    `json:"error_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	ExtraData    map[string]interface{} `json:"extra_data,omitempty"`
}

// IsSuccess 检查响应是否成功
func (r *CreateOrderResponse) IsSuccess() bool {
	return r.Success
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(payURL string) *CreateOrderResponse {
	return &CreateOrderResponse{
		Success: true,
		PayURL:  payURL,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) *CreateOrderResponse {
	return &CreateOrderResponse{
		Success:      false,
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

// WaitProductRequest 等待产品请求
type WaitProductRequest struct {
	OutOrderNo     string                 `json:"out_order_no"`
	Money          int                    `json:"money"`
	NotifyMoney    int                    `json:"notify_money"`
	MerchantID     int64                  `json:"merchant_id"`
	TenantID       int64                  `json:"tenant_id"`
	ChannelID      int64                  `json:"channel_id"`
	PluginID       int64                  `json:"plugin_id"`
	PluginType     string                 `json:"plugin_type"`
	PluginUpstream int                    `json:"plugin_upstream"`
	Channel        map[string]interface{} `json:"channel,omitempty"`
}

// WaitProductResponse 等待产品响应
type WaitProductResponse struct {
	ProductID    string `json:"product_id"`  // 产品ID
	WriteoffID   *int64 `json:"writeoff_id"` // 核销ID
	CookieID     string `json:"cookie_id"`   // Cookie ID
	Money        int    `json:"money"`       // 金额（可能被调整）
	Success      bool   `json:"success"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// NewWaitProductSuccessResponse 创建成功响应
func NewWaitProductSuccessResponse(productID string, writeoffID *int64, cookieID string, money int) *WaitProductResponse {
	return &WaitProductResponse{
		ProductID:  productID,
		WriteoffID: writeoffID,
		CookieID:   cookieID,
		Money:      money,
		Success:    true,
	}
}

// NewWaitProductErrorResponse 创建错误响应
func NewWaitProductErrorResponse(code int, message string) *WaitProductResponse {
	return &WaitProductResponse{
		Success:      false,
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

// CallbackSubmitRequest 下单回调请求
// 参考 Python: callback_plugin_submit 的参数
type CallbackSubmitRequest struct {
	OrderNo        string `json:"order_no"`        // 订单号
	OutOrderNo     string `json:"out_order_no"`    // 商户订单号
	PluginID       int64  `json:"plugin_id"`       // 插件ID
	Tax            int    `json:"tax"`             // 税费
	PluginType     string `json:"plugin_type"`     // 插件类型
	Money          int    `json:"money"`           // 订单金额（分）
	DomainID       *int64 `json:"domain_id"`       // 域名ID
	NotifyMoney    int    `json:"notify_money"`    // 通知金额（分）
	OrderID        string `json:"order_id"`        // 订单数据库ID
	ProductID      string `json:"product_id"`      // 产品ID
	CookieID       string `json:"cookie_id"`       // Cookie ID（可选）
	ChannelID      int64  `json:"channel_id"`      // 支付通道ID
	MerchantID     int64  `json:"merchant_id"`     // 商户ID
	WriteoffID     *int64 `json:"writeoff_id"`     // 核销ID
	TenantID       int64  `json:"tenant_id"`       // 租户ID
	CreateDatetime string `json:"create_datetime"` // 订单创建时间（格式：2006-01-02 15:04:05）
	NotifyURL      string `json:"notify_url"`      // 通知URL
	PluginUpstream int    `json:"plugin_upstream"` // 插件上游类型
}
