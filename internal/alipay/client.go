package alipay

import (
	"crypto/md5"
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

	"github.com/golang-pay-core/internal/database"
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
	AppCertSN        string // 应用证书 SN（证书模式）
	AlipayRootCertSN string // 支付宝根证书 SN（证书模式）
}

// NewClient 创建支付宝客户端
// 参考 Python: get_alipay_sdk 的逻辑
func NewClient(product *models.AlipayProduct, notifyURL string, isOrder bool) (*Client, error) {
	// 根据产品类型选择配置
	var appID, privateKey, publicKey string
	var signType string
	var appAuthToken string
	var appPublicCrt, alipayPublicCrt, alipayRootCrt string

	// 参考 Python: get_alipay_sdk 的逻辑
	// 拉单用父商户的appid,转账用自己的appid
	// if (product.account_type == 0 and is_order) or (product.account_type in [4, 6, 7]):
	if (product.AccountType == 0 && isOrder) || (product.AccountType == 4 || product.AccountType == 6 || product.AccountType == 7) {
		// 子商户/服务商授权商户，需要使用父商户的配置
		if product.AccountType != 7 {
			// 使用直接父商户
			if product.ParentID == nil {
				return nil, fmt.Errorf("子商户产品缺少父商户配置")
			}

			// 查询父产品
			var parent models.AlipayProduct
			if err := database.DB.Where("id = ?", *product.ParentID).First(&parent).Error; err != nil {
				return nil, fmt.Errorf("查询父商户失败: %w", err)
			}

			appID = parent.AppID
			privateKey = parent.PrivateKey
			publicKey = parent.PublicKey
			signType = parent.SignType
			appPublicCrt = parent.AppPublicCrt
			alipayPublicCrt = parent.AlipayPublicCrt
			alipayRootCrt = parent.AlipayRootCrt
			// app_auth_token 仍然使用子商户的
			appAuthToken = product.AppAuthToken
		} else {
			// account_type == 7，使用父商户的父商户（祖父商户）
			if product.ParentID == nil {
				return nil, fmt.Errorf("服务商授权商户缺少父商户配置")
			}

			// 查询父产品
			var parent models.AlipayProduct
			if err := database.DB.Where("id = ?", *product.ParentID).First(&parent).Error; err != nil {
				return nil, fmt.Errorf("查询父商户失败: %w", err)
			}

			// 查询祖父商户
			if parent.ParentID == nil {
				return nil, fmt.Errorf("服务商授权商户缺少祖父商户配置")
			}

			var grandParent models.AlipayProduct
			if err := database.DB.Where("id = ?", *parent.ParentID).First(&grandParent).Error; err != nil {
				return nil, fmt.Errorf("查询祖父商户失败: %w", err)
			}

			appID = grandParent.AppID
			privateKey = grandParent.PrivateKey
			publicKey = grandParent.PublicKey
			signType = grandParent.SignType
			appPublicCrt = grandParent.AppPublicCrt
			alipayPublicCrt = grandParent.AlipayPublicCrt
			alipayRootCrt = grandParent.AlipayRootCrt
			// app_auth_token 仍然使用子商户的
			appAuthToken = product.AppAuthToken
		}
	} else {
		// 使用产品自己的配置
		appID = product.AppID
		privateKey = product.PrivateKey
		publicKey = product.PublicKey
		signType = product.SignType
		appPublicCrt = product.AppPublicCrt
		alipayPublicCrt = product.AlipayPublicCrt
		alipayRootCrt = product.AlipayRootCrt
		appAuthToken = product.AppAuthToken
	}

	// 解析私钥
	appPrivateKey, err := parsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("解析应用私钥失败: %w", err)
	}

	// 判断是否使用数字证书（signType == "0" 表示普通公钥，其他表示数字证书）
	isDC := signType != "0" && signType != ""

	// 解析公钥（证书模式下可能不需要公钥字符串，使用证书代替）
	var alipayPublicKey *rsa.PublicKey
	if !isDC {
		// 普通公钥模式，必须解析公钥
		var err error
		alipayPublicKey, err = parsePublicKey(publicKey)
		if err != nil {
			return nil, fmt.Errorf("解析支付宝公钥失败: %w", err)
		}
	} else {
		// 证书模式，公钥可以为空（使用证书代替）
		// 但如果提供了公钥，也可以解析（用于验证回调等场景）
		if publicKey != "" {
			var err error
			alipayPublicKey, err = parsePublicKey(publicKey)
			if err != nil {
				// 证书模式下公钥解析失败不影响，只记录警告
				// 因为主要使用证书进行签名和验证
			}
		}
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
		client.AppPublicCert = appPublicCrt
		client.AlipayPublicCert = alipayPublicCrt
		client.AlipayRootCert = alipayRootCrt

		// 验证证书是否已配置（证书模式下必须提供证书）
		if appPublicCrt == "" || alipayPublicCrt == "" || alipayRootCrt == "" {
			return nil, fmt.Errorf("证书模式下必须提供完整的证书配置（应用公钥证书、支付宝公钥证书、支付宝根证书）")
		}

		// 提取证书 SN
		appCertSN, err := getCertSN(appPublicCrt)
		if err != nil {
			return nil, fmt.Errorf("提取应用证书SN失败: %w", err)
		}
		client.AppCertSN = appCertSN

		alipayRootCertSN, err := getRootCertSN(alipayRootCrt)
		if err != nil {
			return nil, fmt.Errorf("提取支付宝根证书SN失败: %w", err)
		}
		client.AlipayRootCertSN = alipayRootCertSN
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
// 参考 Python: cut_key 函数和 get_alipay_sdk 中的公钥处理
func parsePublicKey(keyStr string) (*rsa.PublicKey, error) {
	if keyStr == "" {
		return nil, fmt.Errorf("公钥不能为空")
	}

	// 移除可能的头部和尾部标记
	keyStr = strings.TrimSpace(keyStr)

	// 如果已经是完整的 PEM 格式，直接解析
	if strings.Contains(keyStr, "-----BEGIN PUBLIC KEY-----") {
		block, _ := pem.Decode([]byte(keyStr))
		if block != nil {
			publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				if rsaKey, ok := publicKey.(*rsa.PublicKey); ok {
					return rsaKey, nil
				}
			}
		}
	}

	// 移除头部和尾部标记
	keyStr = strings.ReplaceAll(keyStr, "-----BEGIN PUBLIC KEY-----", "")
	keyStr = strings.ReplaceAll(keyStr, "-----END PUBLIC KEY-----", "")
	keyStr = strings.ReplaceAll(keyStr, "\n", "")
	keyStr = strings.ReplaceAll(keyStr, "\r", "")
	keyStr = strings.ReplaceAll(keyStr, " ", "")
	keyStr = strings.ReplaceAll(keyStr, "\t", "")

	if keyStr == "" {
		return nil, fmt.Errorf("公钥内容为空")
	}

	// 每 64 个字符换行（PEM 格式要求）
	formattedKey := formatKey(keyStr, 64)

	// 构建完整的 PEM 格式
	pemKey := "-----BEGIN PUBLIC KEY-----\n" + formattedKey + "\n-----END PUBLIC KEY-----"

	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("无法解析公钥: PEM 解码失败")
	}

	// 尝试解析 PKIX 格式
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// 如果失败，尝试解析 RSA 公钥（PKCS#1 格式）
		rsaKey, err2 := x509.ParsePKCS1PublicKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("解析公钥失败: PKIX=%v, PKCS1=%v", err, err2)
		}
		return rsaKey, nil
	}

	if rsaKey, ok := publicKey.(*rsa.PublicKey); ok {
		return rsaKey, nil
	}

	return nil, fmt.Errorf("公钥格式不正确: 不是 RSA 公钥")
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

// getCertSN 获取证书 SN
// 参考 Python: DCAliPay.get_cert_sn
// 算法：CN={},OU={},O={},C={} + serial_number 的 MD5
func getCertSN(certString string) (string, error) {
	// 解析证书
	block, _ := pem.Decode([]byte(certString))
	if block == nil {
		return "", fmt.Errorf("无法解析证书")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("解析证书失败: %w", err)
	}

	// 构建证书 SN 字符串
	// Python: name = 'CN={},OU={},O={},C={}'.format(certIssue.CN, certIssue.OU, certIssue.O, certIssue.C)
	issuer := cert.Issuer
	name := fmt.Sprintf("CN=%s,OU=%s,O=%s,C=%s",
		issuer.CommonName,
		strings.Join(issuer.OrganizationalUnit, ","),
		strings.Join(issuer.Organization, ","),
		strings.Join(issuer.Country, ","))

	// 添加序列号
	snString := name + cert.SerialNumber.String()

	// 计算 MD5
	hash := md5.Sum([]byte(snString))
	return fmt.Sprintf("%x", hash), nil
}

// getRootCertSN 获取根证书 SN
// 参考 Python: DCAliPay.get_root_cert_sn
// 根证书可能包含多个证书，需要找到第一个有效的证书 SN
func getRootCertSN(rootCertString string) (string, error) {
	// 根证书可能包含多个证书，用两个换行符分隔
	// Python: 根证书中，每个 cert 中间有两个回车间隔
	certs := strings.Split(rootCertString, "\n\n")

	for _, certStr := range certs {
		if strings.TrimSpace(certStr) == "" {
			continue
		}

		// 尝试解析每个证书
		sn, err := getCertSN(certStr)
		if err != nil {
			continue
		}

		// 返回第一个有效的证书 SN
		return sn, nil
	}

	return "", fmt.Errorf("无法从根证书中提取有效的证书 SN")
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

	// 证书模式下，添加证书 SN（必须在签名之前添加）
	if c.IsDC {
		params["app_cert_sn"] = c.AppCertSN
		params["alipay_root_cert_sn"] = c.AlipayRootCertSN
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
