package alipay

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/golang-pay-core/internal/models"
)

// Client 支付宝客户端
type Client struct {
	AppID            string
	AppPrivateKey    *rsa.PrivateKey
	AlipayPublicKey  *rsa.PublicKey
	AppAuthToken     string
	AppAuthCode      string
	Gateway          string
	NotifyURL        string
	SignType         string
	Proxies          map[string]string
	HTTPClient       *http.Client
	IsDC             bool // 是否使用数字证书
	AppPublicCert    string
	AlipayPublicCert string
	AlipayRootCert   string
}

// NewClient 创建支付宝客户端
func NewClient(product *models.AlipayProduct, notifyURL string, isOrder bool) (*Client, error) {
	// 根据产品类型选择配置
	var appID, privateKey, publicKey string
	var signType string
	var appAuthToken string

	// 参考 Python: get_alipay_sdk 的逻辑
	// 注意：这里简化处理，直接使用产品的配置
	// TODO: 实现完整的父商户配置逻辑
	appID = product.AppID
	privateKey = product.PrivateKey
	publicKey = product.PublicKey
	signType = product.SignType
	appAuthToken = product.AppAuthToken

	// 解析私钥
	appPrivateKey, err := parsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("解析应用私钥失败: %w", err)
	}

	// 解析公钥
	alipayPublicKey, err := parsePublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("解析支付宝公钥失败: %w", err)
	}

	// 设置代理
	proxies := make(map[string]string)
	if product.ProxyIP != "" {
		proxyURL := fmt.Sprintf("http://%s:%d", product.ProxyIP, product.ProxyPort)
		if product.ProxyUser != "" && product.ProxyPwd != "" {
			proxyURL = fmt.Sprintf("http://%s:%s@%s:%d", product.ProxyUser, product.ProxyPwd, product.ProxyIP, product.ProxyPort)
		}
		proxies["http"] = proxyURL
		proxies["https"] = proxyURL
	}

	// 创建 HTTP 客户端
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	// 判断是否使用数字证书（signType == "0" 表示普通公钥，其他表示数字证书）
	isDC := signType != "0" && signType != ""

	client := &Client{
		AppID:           appID,
		AppPrivateKey:   appPrivateKey,
		AlipayPublicKey: alipayPublicKey,
		AppAuthToken:    appAuthToken,
		Gateway:         "https://openapi.alipay.com/gateway.do",
		NotifyURL:       notifyURL,
		SignType:        "RSA2",
		Proxies:         proxies,
		HTTPClient:      httpClient,
		IsDC:            isDC,
	}

	// 如果是数字证书模式，设置证书
	if isDC {
		// TODO: 从产品获取证书信息
		// client.AppPublicCert = product.AppPublicCrt
		// client.AlipayPublicCert = product.AlipayPublicCrt
		// client.AlipayRootCert = product.AlipayRootCrt
	}

	return client, nil
}

// parsePrivateKey 解析 RSA 私钥
func parsePrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	// 移除可能的头部和尾部标记
	keyStr = strings.TrimSpace(keyStr)
	keyStr = strings.ReplaceAll(keyStr, "-----BEGIN RSA PRIVATE KEY-----", "")
	keyStr = strings.ReplaceAll(keyStr, "-----END RSA PRIVATE KEY-----", "")
	keyStr = strings.ReplaceAll(keyStr, "-----BEGIN PRIVATE KEY-----", "")
	keyStr = strings.ReplaceAll(keyStr, "-----END PRIVATE KEY-----", "")
	keyStr = strings.ReplaceAll(keyStr, "\n", "")
	keyStr = strings.ReplaceAll(keyStr, " ", "")

	// 每 64 个字符换行（PEM 格式要求）
	formattedKey := formatKey(keyStr, 64)

	block, _ := pem.Decode([]byte("-----BEGIN RSA PRIVATE KEY-----\n" + formattedKey + "\n-----END RSA PRIVATE KEY-----"))
	if block == nil {
		return nil, fmt.Errorf("无法解析私钥")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试 PKCS8 格式
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("解析私钥失败: %v, %v", err, err2)
		}
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("私钥格式不正确")
	}

	return privateKey, nil
}

// parsePublicKey 解析 RSA 公钥
func parsePublicKey(keyStr string) (*rsa.PublicKey, error) {
	// 移除可能的头部和尾部标记
	keyStr = strings.TrimSpace(keyStr)
	keyStr = strings.ReplaceAll(keyStr, "-----BEGIN PUBLIC KEY-----", "")
	keyStr = strings.ReplaceAll(keyStr, "-----END PUBLIC KEY-----", "")
	keyStr = strings.ReplaceAll(keyStr, "\n", "")
	keyStr = strings.ReplaceAll(keyStr, " ", "")

	// 每 64 个字符换行（PEM 格式要求）
	formattedKey := formatKey(keyStr, 64)

	block, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\n" + formattedKey + "\n-----END PUBLIC KEY-----"))
	if block == nil {
		return nil, fmt.Errorf("无法解析公钥")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	if rsaKey, ok := publicKey.(*rsa.PublicKey); ok {
		return rsaKey, nil
	}

	return nil, fmt.Errorf("公钥格式不正确")
}

// formatKey 格式化密钥字符串
func formatKey(keyStr string, lineLen int) string {
	var result strings.Builder
	for i := 0; i < len(keyStr); i += lineLen {
		end := i + lineLen
		if end > len(keyStr) {
			end = len(keyStr)
		}
		result.WriteString(keyStr[i:end])
		if end < len(keyStr) {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// TradeWapPay 手机网站支付
// 参考 Python: alipay.api_alipay_trade_wap_pay
func (c *Client) TradeWapPay(subject, outTradeNo, totalAmount, notifyURL string, others map[string]interface{}) (string, error) {
	// 构建 biz_content
	bizContent := map[string]interface{}{
		"subject":      subject,
		"out_trade_no": outTradeNo,
		"total_amount": totalAmount,
		"product_code": "QUICK_WAP_WAY",
	}

	// 添加其他参数到 biz_content
	if others != nil {
		for k, v := range others {
			bizContent[k] = v
		}
	}

	// 将 biz_content 转换为 JSON 字符串
	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return "", fmt.Errorf("序列化 biz_content 失败: %w", err)
	}

	// 构建请求参数
	params := map[string]interface{}{
		"app_id":      c.AppID,
		"method":      "alipay.trade.wap.pay",
		"charset":     "utf-8",
		"sign_type":   c.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"biz_content": string(bizContentJSON),
	}

	// 如果有 app_auth_token，添加到参数中
	if c.AppAuthToken != "" {
		params["app_auth_token"] = c.AppAuthToken
	}

	// 生成签名
	sign, err := signParams(params, c.AppPrivateKey)
	if err != nil {
		return "", fmt.Errorf("生成签名失败: %w", err)
	}
	params["sign"] = sign

	// 构建 URL
	queryString := c.buildQueryString(params)
	payURL := c.Gateway + "?" + queryString

	return payURL, nil
}

// buildQueryString 构建查询字符串
func (c *Client) buildQueryString(params map[string]interface{}) string {
	values := url.Values{}
	for k, v := range params {
		if v != nil {
			values.Set(k, fmt.Sprintf("%v", v))
		}
	}
	return values.Encode()
}

// GetRedirectURL 获取重定向后的 URL
// 参考 Python: 如果 redirects=True，发送 GET 请求获取 Location
func (c *Client) GetRedirectURL(payURL string) (string, error) {
	req, err := http.NewRequest("GET", payURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 获取重定向 URL
	location := resp.Header.Get("Location")
	if location != "" {
		return location, nil
	}

	// 如果没有 Location 头，尝试从响应体中提取
	// 参考 Python: re.findall(r'<div class="Todo">(.+)</div>', response.text)
	if resp.StatusCode != 302 && resp.StatusCode != 301 {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])

		// 尝试提取错误信息
		re := regexp.MustCompile(`<div class="Todo">(.+?)</div>`)
		matches := re.FindStringSubmatch(bodyStr)
		if len(matches) > 1 {
			return "", fmt.Errorf("获取支付地址错误: %s", matches[1])
		}

		return "", fmt.Errorf("获取支付地址错误: %s", bodyStr)
	}

	return "", fmt.Errorf("无法获取重定向 URL")
}
