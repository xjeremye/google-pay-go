package alipay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"github.com/golang-pay-core/internal/service"
	"go.uber.org/zap"
)

// MockPlugin 支付宝模拟插件
// 用于压测时模拟支付宝网关延迟和回调，不调用真实API
// 继承自 BasePlugin，保持与支付宝插件一致的行为
type MockPlugin struct {
	*BasePlugin
	// 模拟延迟配置（毫秒）
	MinDelay int // 最小延迟
	MaxDelay int // 最大延迟
	// 是否模拟回调
	SimulateCallback bool
	// 回调延迟（秒）
	CallbackDelay int
}

// NewMockPlugin 创建支付宝模拟插件
func NewMockPlugin(pluginID int64) *MockPlugin {
	return &MockPlugin{
		BasePlugin:       NewBasePlugin(pluginID),
		MinDelay:         50,   // 默认最小延迟50ms
		MaxDelay:         200,  // 默认最大延迟200ms
		SimulateCallback: true, // 默认启用回调模拟
		CallbackDelay:    5,    // 默认5秒后回调
	}
}

// CreateOrder 创建订单（模拟）
// 不真正调用支付宝API，而是模拟延迟后返回模拟支付URL
// 如果配置了模拟回调，会异步触发回调
func (p *MockPlugin) CreateOrder(ctx context.Context, req *plugin.CreateOrderRequest) (*plugin.CreateOrderResponse, error) {
	// 模拟网络延迟（模拟支付宝网关响应时间）
	delay := p.getRandomDelay()
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// 生成模拟支付URL（格式类似支付宝）
	// 格式：mock://pay.example.com/pay?order_no=xxx&amount=xxx
	payURL := fmt.Sprintf("mock://pay.example.com/pay?order_no=%s&out_order_no=%s&amount=%d&plugin_type=alipay_mock",
		req.OrderNo, req.OutOrderNo, req.Money)

	// 如果配置了模拟回调，异步触发回调
	if p.SimulateCallback && req.NotifyURL != "" && req.ProductID != "" {
		go p.simulateCallback(ctx, req)
	}

	return plugin.NewSuccessResponse(payURL), nil
}

// WaitProduct 等待产品
// 继承自 BasePlugin，使用真实的支付宝产品选择逻辑
func (p *MockPlugin) WaitProduct(ctx context.Context, req *plugin.WaitProductRequest) (*plugin.WaitProductResponse, error) {
	// 使用 BasePlugin 的实现，保持与真实支付宝插件一致
	return p.BasePlugin.WaitProduct(ctx, req)
}

// CallbackSubmit 下单回调
// 继承自 BasePlugin，使用真实的回调处理逻辑（更新订单备注、日统计等）
func (p *MockPlugin) CallbackSubmit(ctx context.Context, req *plugin.CallbackSubmitRequest) error {
	// 使用 BasePlugin 的实现，保持与真实支付宝插件一致
	return p.BasePlugin.CallbackSubmit(ctx, req)
}

// getRandomDelay 获取随机延迟时间（毫秒）
func (p *MockPlugin) getRandomDelay() int {
	if p.MaxDelay <= p.MinDelay {
		return p.MinDelay
	}
	return p.MinDelay + rand.Intn(p.MaxDelay-p.MinDelay+1)
}

// simulateCallback 模拟支付回调
// 在指定延迟后，模拟支付宝官方回调到我们自己的系统回调接口（不是回调给商户）
// 完全模拟真实支付宝回调格式，用于测试和压测
// 回调流程：模拟支付宝 -> 我们的系统回调接口 -> 处理订单状态 -> 通知商户（如果配置了 NotifyURL）
func (p *MockPlugin) simulateCallback(ctx context.Context, req *plugin.CreateOrderRequest) {
	// 检查必需参数：产品ID（用于构建回调路径）和 DomainURL（用于构建回调基础URL）
	if req.ProductID == "" {
		logger.Logger.Debug("跳过模拟回调：缺少产品ID",
			zap.String("order_no", req.OrderNo),
			zap.String("product_id", req.ProductID))
		return
	}
	if req.DomainURL == "" {
		logger.Logger.Debug("跳过模拟回调：缺少域名URL",
			zap.String("order_no", req.OrderNo),
			zap.String("domain_url", req.DomainURL))
		return
	}

	logger.Logger.Info("开始模拟支付宝回调到系统回调接口",
		zap.String("order_no", req.OrderNo),
		zap.String("out_order_no", req.OutOrderNo),
		zap.String("product_id", req.ProductID),
		zap.String("domain_url", req.DomainURL),
		zap.Int("callback_delay_seconds", p.CallbackDelay))

	// 等待回调延迟时间
	time.Sleep(time.Duration(p.CallbackDelay) * time.Second)

	// 构建回调URL：模拟支付宝回调到我们自己的系统回调接口
	// 格式：{base_url}/api/pay/order/notify/alipay_mock/{product_id}/
	callbackURL := p.buildCallbackURL(ctx, req)
	if callbackURL == "" {
		logger.Logger.Warn("构建回调URL失败，跳过模拟回调",
			zap.String("order_no", req.OrderNo),
			zap.String("domain_url", req.DomainURL),
			zap.String("product_id", req.ProductID))
		return
	}

	logger.Logger.Debug("构建回调URL成功",
		zap.String("order_no", req.OrderNo),
		zap.String("callback_url", callbackURL))

	// 构建回调参数（模拟支付宝回调格式）
	callbackParams := p.buildCallbackParams(req)

	// 发送POST请求到回调URL（重试机制：最多重试3次）
	maxRetries := 3
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// 重试前等待一段时间（指数退避）
			retryDelay := time.Duration(attempt-1) * time.Second
			logger.Logger.Info("重试模拟回调",
				zap.String("order_no", req.OrderNo),
				zap.Int("attempt", attempt),
				zap.Duration("retry_delay", retryDelay))
			time.Sleep(retryDelay)
		}

		// 创建HTTP客户端
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
			lastErr = err
			logger.Logger.Warn("发送模拟回调请求失败",
				zap.String("order_no", req.OrderNo),
				zap.String("out_order_no", req.OutOrderNo),
				zap.String("callback_url", callbackURL),
				zap.Int("attempt", attempt),
				zap.Int("max_retries", maxRetries),
				zap.Error(err))
			continue
		}

		// 读取响应体（用于调试）
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode == http.StatusOK {
			logger.Logger.Info("模拟回调成功",
				zap.String("order_no", req.OrderNo),
				zap.String("out_order_no", req.OutOrderNo),
				zap.String("callback_url", callbackURL),
				zap.Int("status_code", resp.StatusCode),
				zap.String("response", string(bodyBytes)),
				zap.Int("attempt", attempt))
			return // 成功，退出重试循环
		}

		lastErr = fmt.Errorf("回调返回非200状态码: %d, response: %s", resp.StatusCode, string(bodyBytes))
		logger.Logger.Warn("模拟回调返回非200状态码",
			zap.String("order_no", req.OrderNo),
			zap.String("out_order_no", req.OutOrderNo),
			zap.String("callback_url", callbackURL),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(bodyBytes)),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries))
	}

	// 所有重试都失败
	logger.Logger.Error("模拟回调失败（已重试所有次数）",
		zap.String("order_no", req.OrderNo),
		zap.String("out_order_no", req.OutOrderNo),
		zap.String("callback_url", callbackURL),
		zap.Int("max_retries", maxRetries),
		zap.Error(lastErr))
	// 注意：回调失败不影响主流程，只记录错误日志
}

// buildCallbackURL 构建回调URL
// 模拟支付宝回调到我们自己的系统回调接口（不是回调给商户）
// 格式：{scheme}://{host}:{port}/api/pay/order/notify/alipay_mock/{product_id}/
func (p *MockPlugin) buildCallbackURL(ctx context.Context, req *plugin.CreateOrderRequest) string {
	// 获取通知域名（从系统配置获取，如果没有则使用 DomainURL）
	// 参考其他支付宝插件的实现：alipay.notify_domain 或 notify.host
	notifyDomain := getSystemConfigByPath(ctx, "alipay.notify_domain")
	if notifyDomain == "" {
		// 如果系统配置中没有，尝试使用 notify.host
		notifyDomain = getSystemConfigByPath(ctx, "notify.host")
	}
	if notifyDomain == "" {
		// 如果系统配置中都没有，使用 DomainURL
		notifyDomain = req.DomainURL
	}

	if notifyDomain == "" {
		logger.Logger.Warn("无法获取通知域名，跳过模拟回调",
			zap.String("order_no", req.OrderNo),
			zap.String("domain_url", req.DomainURL))
		return ""
	}

	// 解析通知域名URL（确保格式正确）
	baseURL, err := url.Parse(notifyDomain)
	if err != nil {
		logger.Logger.Warn("解析通知域名失败",
			zap.String("notify_domain", notifyDomain),
			zap.String("order_no", req.OrderNo),
			zap.Error(err))
		return ""
	}

	// 构建回调路径：/api/pay/order/notify/alipay_mock/{product_id}/
	callbackPath := fmt.Sprintf("/api/pay/order/notify/alipay_mock/%s/", req.ProductID)

	// 组合完整URL（包含端口号，如果有）
	var callbackURL string
	if baseURL.Port() != "" {
		callbackURL = fmt.Sprintf("%s://%s:%s%s", baseURL.Scheme, baseURL.Hostname(), baseURL.Port(), callbackPath)
	} else {
		callbackURL = fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Hostname(), callbackPath)
	}

	return callbackURL
}

// buildCallbackParams 构建回调参数（模拟支付宝回调格式）
// 完全模拟真实支付宝回调参数，确保格式一致
func (p *MockPlugin) buildCallbackParams(req *plugin.CreateOrderRequest) map[string]string {
	now := time.Now()
	params := make(map[string]string)

	// 核心参数（必需）
	params["out_trade_no"] = req.OutOrderNo                                  // 商户订单号（必需）
	params["trade_no"] = fmt.Sprintf("MOCK_%d", now.UnixNano())              // 模拟支付宝交易号
	params["trade_status"] = "TRADE_SUCCESS"                                 // 交易状态：成功
	params["total_amount"] = fmt.Sprintf("%.2f", float64(req.Money)/100.0)   // 金额（元，必需）
	params["receipt_amount"] = fmt.Sprintf("%.2f", float64(req.Money)/100.0) // 实收金额

	// 买家信息
	params["buyer_id"] = fmt.Sprintf("MOCK_BUYER_%d", rand.Int63n(10000))                   // 模拟买家ID
	params["buyer_logon_id"] = fmt.Sprintf("mock_buyer_%d@example.com", rand.Int63n(10000)) // 模拟买家账号

	// 卖家信息
	params["seller_id"] = fmt.Sprintf("MOCK_SELLER_%d", rand.Int63n(10000))                // 模拟卖家ID
	params["seller_email"] = fmt.Sprintf("mock_seller_%d@example.com", rand.Int63n(10000)) // 模拟卖家邮箱

	// 时间信息（使用支付宝标准格式：yyyy-MM-dd HH:mm:ss）
	timeStr := now.Format("2006-01-02 15:04:05")
	params["gmt_payment"] = timeStr // 支付时间
	params["gmt_create"] = timeStr  // 交易创建时间
	params["gmt_refund"] = ""       // 退款时间（成功时为空）
	params["notify_time"] = timeStr // 通知时间

	// 通知相关
	params["notify_type"] = "trade_status_sync"                         // 通知类型
	params["notify_id"] = fmt.Sprintf("MOCK_NOTIFY_%d", now.UnixNano()) // 通知ID（唯一）

	// 应用信息
	params["app_id"] = "MOCK_APP_ID" // 模拟APP ID

	// 其他标准参数
	params["charset"] = "utf-8"       // 字符集
	params["version"] = "1.0"         // 版本
	params["sign_type"] = "RSA2"      // 签名类型
	params["sign"] = "MOCK_SIGNATURE" // 模拟签名（mock插件不需要真实签名）

	// 可选参数（模拟真实回调可能包含的字段）
	params["invoice_amount"] = fmt.Sprintf("%.2f", float64(req.Money)/100.0) // 开票金额
	params["point_amount"] = "0.00"                                          // 积分金额
	params["refund_fee"] = "0.00"                                            // 退款金额
	params["subject"] = fmt.Sprintf("订单-%s", req.OutOrderNo)                 // 订单标题
	params["body"] = fmt.Sprintf("模拟订单-%s", req.OutOrderNo)                  // 订单描述

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
// 从数据库获取真实的超时时间配置，如果没有配置则使用默认值
func (p *MockPlugin) GetTimeout(ctx context.Context, pluginID int64) int {
	// 尝试从插件配置中获取超时时间（使用缓存）
	pluginService := service.NewPluginService()
	pluginConfig, err := pluginService.GetPluginConfigByKey(ctx, pluginID, "timeout")
	if err == nil && pluginConfig != nil {
		// 解析配置值（可能是 JSON 字符串或直接的值）
		var timeout int
		if err := json.Unmarshal([]byte(pluginConfig.Value), &timeout); err != nil {
			// 如果不是 JSON，尝试直接解析为整数
			if parsed, err := strconv.Atoi(pluginConfig.Value); err == nil {
				timeout = parsed
			}
		}
		// 如果解析成功且值有效，返回配置的超时时间
		if timeout > 0 {
			return timeout
		}
	}

	// 如果缓存中没有，尝试从数据库直接查询（作为备用方案）
	var dbConfig models.PayPluginConfig
	if err := database.DB.Where("parent_id = ? AND key = ? AND status = ?", pluginID, "timeout", true).
		First(&dbConfig).Error; err == nil {
		var timeout int
		if err := json.Unmarshal([]byte(dbConfig.Value), &timeout); err == nil {
			if timeout > 0 {
				return timeout
			}
		} else if parsed, err := strconv.Atoi(dbConfig.Value); err == nil && parsed > 0 {
			return parsed
		}
	}

	// 如果都没有配置，使用 BasePlugin 的默认值（5分钟）
	return p.BasePlugin.GetTimeout(ctx, pluginID)
}
