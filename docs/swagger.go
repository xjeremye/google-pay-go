package docs

import "github.com/golang-pay-core/internal/response"

// @title           支付系统核心 API
// @version         1.0
// @description     基于 Golang 开发的高并发支付系统核心 API
// @termsOfService  https://example.com/terms/

// @contact.name   API Support
// @contact.url    https://example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   http https

// Response 统一响应结构
// @Description 统一的 API 响应格式
type Response struct {
	Code    int         `json:"code" example:"200"`          // 响应码
	Message string      `json:"message" example:"success"`   // 响应消息
	Data    interface{} `json:"data,omitempty" example:"{}"` // 响应数据
}

// SuccessResponse 成功响应示例
// @Description 成功响应示例
type SuccessResponse struct {
	response.Response
	Data interface{} `json:"data"`
}

// ErrorResponse 错误响应示例
// @Description 错误响应示例
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"参数错误"`
}
