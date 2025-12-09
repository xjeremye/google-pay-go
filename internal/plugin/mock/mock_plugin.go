package mock

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-pay-core/internal/plugin"
)

// MockPlugin 模拟测试插件
// 用于压测时模拟支付宝网关延迟和回调，不调用真实API
type MockPlugin struct {
	*plugin.BasePlugin
	// 模拟延迟配置（毫秒）
	MinDelay int // 最小延迟
	MaxDelay int // 最大延迟
	// 是否模拟回调
	SimulateCallback bool
	// 回调延迟（秒）
	CallbackDelay int
}

// NewMockPlugin 创建模拟插件
func NewMockPlugin(pluginID int64) *MockPlugin {
	return &MockPlugin{
		BasePlugin:       plugin.NewBasePlugin(pluginID),
		MinDelay:         50,   // 默认最小延迟50ms
		MaxDelay:         200,  // 默认最大延迟200ms
		SimulateCallback: true, // 默认启用回调模拟
		CallbackDelay:    5,    // 默认5秒后回调
	}
}

// CreateOrder 创建订单（模拟）
// 模拟支付宝网关延迟，返回模拟支付URL
func (p *MockPlugin) CreateOrder(ctx context.Context, req *plugin.CreateOrderRequest) (*plugin.CreateOrderResponse, error) {
	// 模拟网络延迟（模拟支付宝网关响应时间）
	delay := p.getRandomDelay()
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// 生成模拟支付URL
	// 格式：mock://pay.example.com/pay?order_no=xxx&amount=xxx
	payURL := fmt.Sprintf("mock://pay.example.com/pay?order_no=%s&out_order_no=%s&amount=%d&plugin_type=mock",
		req.OrderNo, req.OutOrderNo, req.Money)

	// 如果配置了模拟回调，异步触发回调
	// 注意：需要从NotifyURL中提取回调地址，并构建正确的回调URL
	if p.SimulateCallback && req.NotifyURL != "" {
		go p.simulateCallback(ctx, req)
	}

	return plugin.NewSuccessResponse(payURL), nil
}

// WaitProduct 等待产品（模拟）
// 返回模拟的产品ID和核销ID
func (p *MockPlugin) WaitProduct(ctx context.Context, req *plugin.WaitProductRequest) (*plugin.WaitProductResponse, error) {
	// 模拟产品选择延迟
	delay := p.getRandomDelay() / 2 // 产品选择通常比支付URL生成快
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// 返回模拟产品ID和核销ID
	// 产品ID格式：MOCK_PRODUCT_{随机数}
	// 核销ID：随机生成一个ID
	productID := fmt.Sprintf("MOCK_PRODUCT_%d", rand.Int63n(1000)+1)
	writeoffID := int64(rand.Intn(1000) + 1)

	return plugin.NewWaitProductSuccessResponse(productID, &writeoffID, "", req.Money), nil
}

// CallbackSubmit 下单回调（模拟）
// 模拟订单创建成功后的回调处理
func (p *MockPlugin) CallbackSubmit(ctx context.Context, req *plugin.CallbackSubmitRequest) error {
	// 模拟回调处理延迟
	delay := p.getRandomDelay() / 3
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// 模拟回调处理逻辑（这里可以记录日志或更新统计）
	// 实际实现中，这里可能会更新订单状态、更新统计等
	// 但因为是模拟，我们只做延迟，不真实操作数据库

	return nil
}

// getRandomDelay 获取随机延迟时间（毫秒）
func (p *MockPlugin) getRandomDelay() int {
	if p.MaxDelay <= p.MinDelay {
		return p.MinDelay
	}
	return p.MinDelay + rand.Intn(p.MaxDelay-p.MinDelay+1)
}

// simulateCallback 模拟支付回调
// 在指定延迟后，模拟支付宝回调通知，发送HTTP请求到系统回调接口
func (p *MockPlugin) simulateCallback(ctx context.Context, req *plugin.CreateOrderRequest) {
	// 等待回调延迟时间
	time.Sleep(time.Duration(p.CallbackDelay) * time.Second)

	// 如果没有通知URL或产品ID，跳过回调
	if req.NotifyURL == "" || req.ProductID == "" {
		return
	}

	// 构建回调URL：从NotifyURL提取基础URL，然后构建系统回调接口
	// 格式：{base_url}/api/pay/order/notify/mock/{product_id}/
	callbackURL := p.buildCallbackURL(req)
	if callbackURL == "" {
		return
	}

	// 构建回调参数（模拟支付宝回调格式）
	callbackParams := p.buildCallbackParams(req)

	// 发送POST请求到回调URL
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 构建form数据
	formData := url.Values{}
	for k, v := range callbackParams {
		formData.Set(k, v)
	}

	// 发送POST请求
	resp, err := client.PostForm(callbackURL, formData)
	if err != nil {
		// 回调失败不影响主流程，只记录日志（这里简化处理）
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		// 回调失败，记录日志（这里简化处理）
		return
	}
}

// buildCallbackURL 构建回调URL
// 从NotifyURL提取基础URL，然后构建系统回调接口
func (p *MockPlugin) buildCallbackURL(req *plugin.CreateOrderRequest) string {
	// 解析NotifyURL
	notifyURL, err := url.Parse(req.NotifyURL)
	if err != nil {
		return ""
	}

	// 构建回调URL：/api/pay/order/notify/mock/{product_id}/
	callbackPath := fmt.Sprintf("/api/pay/order/notify/mock/%s/", req.ProductID)

	// 组合完整URL
	callbackURL := fmt.Sprintf("%s://%s%s", notifyURL.Scheme, notifyURL.Host, callbackPath)

	return callbackURL
}

// buildCallbackParams 构建回调参数（模拟支付宝回调格式）
func (p *MockPlugin) buildCallbackParams(req *plugin.CreateOrderRequest) map[string]string {
	params := make(map[string]string)

	// 基本参数（模拟支付宝回调）
	params["out_trade_no"] = req.OutOrderNo                                                 // 商户订单号
	params["trade_no"] = fmt.Sprintf("MOCK_%d", time.Now().UnixNano())                      // 模拟支付宝交易号
	params["trade_status"] = "TRADE_SUCCESS"                                                // 交易状态：成功
	params["total_amount"] = fmt.Sprintf("%.2f", float64(req.Money)/100.0)                  // 金额（元）
	params["receipt_amount"] = fmt.Sprintf("%.2f", float64(req.Money)/100.0)                // 实收金额
	params["buyer_id"] = fmt.Sprintf("MOCK_BUYER_%d", rand.Int63n(10000))                   // 模拟买家ID
	params["buyer_logon_id"] = fmt.Sprintf("mock_buyer_%d@example.com", rand.Int63n(10000)) // 模拟买家账号
	params["seller_id"] = fmt.Sprintf("MOCK_SELLER_%d", rand.Int63n(10000))                 // 模拟卖家ID
	params["gmt_payment"] = time.Now().Format("2006-01-02 15:04:05")                        // 支付时间
	params["notify_time"] = time.Now().Format("2006-01-02 15:04:05")                        // 通知时间
	params["notify_type"] = "trade_status_sync"                                             // 通知类型
	params["notify_id"] = fmt.Sprintf("MOCK_NOTIFY_%d", time.Now().UnixNano())              // 通知ID
	params["app_id"] = "MOCK_APP_ID"                                                        // 模拟APP ID
	params["charset"] = "utf-8"                                                             // 字符集
	params["version"] = "1.0"                                                               // 版本
	params["sign_type"] = "RSA2"                                                            // 签名类型
	params["sign"] = "MOCK_SIGNATURE"                                                       // 模拟签名（mock插件不需要真实签名）

	return params
}

// 实现 PluginCapabilities 接口
var _ plugin.PluginCapabilities = (*MockPlugin)(nil)

// CanHandleExtra 是否可以处理额外参数
func (p *MockPlugin) CanHandleExtra() bool {
	return false
}

// AutoExtra 是否自动处理额外参数
func (p *MockPlugin) AutoExtra() bool {
	return false
}

// ExtraNeedProduct 额外参数是否需要产品
func (p *MockPlugin) ExtraNeedProduct() bool {
	return false
}

// ExtraNeedCookie 额外参数是否需要Cookie
func (p *MockPlugin) ExtraNeedCookie() bool {
	return false
}

// GetTimeout 获取订单超时时间（秒）
func (p *MockPlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	// 模拟订单超时时间：10分钟
	return 600
}
