package service

// OrderError 订单处理错误
type OrderError struct {
	Code    int
	Message string
	Data    interface{}
}

func (e *OrderError) Error() string {
	return e.Message
}

// 订单错误码定义
const (
	ErrCodeAmountInvalid            = 0
	ErrCodeMerchantNotFound         = 7301
	ErrCodeMerchantDisabled         = 7302
	ErrCodeSignInvalid              = 7304
	ErrCodeChannelNotFound          = 7305
	ErrCodeChannelDisabled          = 7306
	ErrCodeMerchantChannelNotFound  = 7307
	ErrCodeMerchantChannelDisabled  = 7308
	ErrCodeChannelTimeInvalid       = 7309
	ErrCodeTenantChannelUnavailable = 7310
	ErrCodeTenantChannelDisabled    = 7311
	ErrCodeAmountZero               = 7312
	ErrCodeAmountOutOfRange         = 7313
	ErrCodeDomainUnavailable        = 7314
	ErrCodeBalanceInsufficient      = 7315
	ErrCodePluginUnavailable        = 7316
	ErrCodePayTypeUnavailable       = 7317
	ErrCodeNoStock                  = 7318
	ErrCodeExtraCheckFailed         = 7319
	ErrCodeCreateFailed             = 7320
	ErrCodeOutOrderNoRequired       = 7321
	ErrCodeOutOrderNoExists         = 7321
	ErrCodeConcurrencyLimit         = 7322
	ErrCodeSystemBusy               = 9999
)

// 错误消息定义
var (
	ErrAmountInvalid       = &OrderError{Code: ErrCodeAmountInvalid, Message: "金额必须大于0"}
	ErrMerchantNotFound    = &OrderError{Code: ErrCodeMerchantNotFound, Message: "商户不存在"}
	ErrMerchantDisabled    = &OrderError{Code: ErrCodeMerchantDisabled, Message: "商户已被禁用,请联系管理员"}
	ErrSignInvalid         = &OrderError{Code: ErrCodeSignInvalid, Message: "签名验证失败"}
	ErrChannelNotFound     = &OrderError{Code: ErrCodeChannelNotFound, Message: "渠道不存在"}
	ErrChannelDisabled     = &OrderError{Code: ErrCodeChannelDisabled, Message: "渠道已被禁用,请联系管理员"}
	ErrChannelTimeInvalid  = &OrderError{Code: ErrCodeChannelTimeInvalid, Message: "通道不在可使用时间"}
	ErrAmountZero          = &OrderError{Code: ErrCodeAmountZero, Message: "金额不能为0"}
	ErrAmountOutOfRange    = &OrderError{Code: ErrCodeAmountOutOfRange, Message: "金额不在范围内"}
	ErrBalanceInsufficient = &OrderError{Code: ErrCodeBalanceInsufficient, Message: "余额不足"}
	ErrPluginUnavailable   = &OrderError{Code: ErrCodePluginUnavailable, Message: "该通道不可用"}
	ErrPayTypeUnavailable  = &OrderError{Code: ErrCodePayTypeUnavailable, Message: "该通道不可用"}
	ErrNoStock             = &OrderError{Code: ErrCodeNoStock, Message: "无库存"}
	ErrExtraCheckFailed    = &OrderError{Code: ErrCodeExtraCheckFailed, Message: "额外参数检查失败"}
	ErrCreateFailed        = &OrderError{Code: ErrCodeCreateFailed, Message: "创建订单失败"}
	ErrOutOrderNoRequired  = &OrderError{Code: ErrCodeOutOrderNoRequired, Message: "商户订单号不能为空"}
	ErrOutOrderNoExists    = &OrderError{Code: ErrCodeOutOrderNoExists, Message: "商户订单号已存在"}
	ErrDomainUnavailable   = &OrderError{Code: ErrCodeDomainUnavailable, Message: "无可用收银台"}
	ErrSystemBusy          = &OrderError{Code: ErrCodeSystemBusy, Message: "系统繁忙,请稍后重试"}
)

// NewOrderError 创建新的订单错误
func NewOrderError(code int, message string) *OrderError {
	return &OrderError{Code: code, Message: message}
}

// NewOrderErrorWithData 创建带数据的订单错误
func NewOrderErrorWithData(code int, message string, data interface{}) *OrderError {
	return &OrderError{Code: code, Message: message, Data: data}
}
