package alipay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TradePagePay PC网站支付
// 参考 Python: alipay.api_alipay_trade_page_pay
func (c *Client) TradePagePay(subject, outTradeNo, totalAmount, notifyURL string, others map[string]interface{}) (string, error) {
	// 构建 biz_content
	bizContent := map[string]interface{}{
		"subject":      subject,
		"out_trade_no": outTradeNo,
		"total_amount": totalAmount,
		"product_code": "FAST_INSTANT_TRADE_PAY",
	}

	// 添加其他参数到 biz_content
	for k, v := range others {
		bizContent[k] = v
	}

	// 将 biz_content 转换为 JSON 字符串
	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return "", fmt.Errorf("序列化 biz_content 失败: %w", err)
	}

	// 构建请求参数
	params := map[string]interface{}{
		"app_id":      c.AppID,
		"method":      "alipay.trade.page.pay",
		"charset":     "utf-8",
		"sign_type":   c.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyURL,
		"biz_content": string(bizContentJSON),
		"return_url":  "", // PC支付需要 return_url，但这里由调用方决定
	}

	// 如果有 app_auth_token，添加到参数中
	if c.AppAuthToken != "" {
		params["app_auth_token"] = c.AppAuthToken
	}

	// 证书模式下，添加证书 SN
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

// TradeAppPay APP支付
// 参考 Python: alipay.api_alipay_trade_app_pay
func (c *Client) TradeAppPay(subject, outTradeNo, totalAmount, notifyURL string, others map[string]interface{}) (string, error) {
	// 构建 biz_content
	bizContent := map[string]interface{}{
		"subject":      subject,
		"out_trade_no": outTradeNo,
		"total_amount": totalAmount,
		"product_code": "QUICK_MSECURITY_PAY",
	}

	// 添加其他参数到 biz_content
	for k, v := range others {
		bizContent[k] = v
	}

	// 将 biz_content 转换为 JSON 字符串
	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return "", fmt.Errorf("序列化 biz_content 失败: %w", err)
	}

	// 构建请求参数
	params := map[string]interface{}{
		"app_id":      c.AppID,
		"method":      "alipay.trade.app.pay",
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

	// 证书模式下，添加证书 SN
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

	// 构建查询字符串（APP支付返回的是查询字符串，不是完整URL）
	queryString := c.buildQueryString(params)

	// 记录 query_log
	requestBodyJSON, _ := json.Marshal(params)
	go c.createQueryLog(
		c.Gateway,
		"POST",
		string(requestBodyJSON),
		"",
		"",
		"支付宝APP支付API请求",
	)

	return queryString, nil
}

// TradePrecreate 扫码支付（预创建订单）
// 参考 Python: alipay.api_alipay_trade_precreate
func (c *Client) TradePrecreate(subject, outTradeNo, totalAmount, notifyURL string, others map[string]interface{}) (string, error) {
	// 构建 biz_content
	bizContent := map[string]interface{}{
		"subject":      subject,
		"out_trade_no": outTradeNo,
		"total_amount": totalAmount,
	}

	// 添加其他参数到 biz_content
	for k, v := range others {
		bizContent[k] = v
	}

	// 将 biz_content 转换为 JSON 字符串
	bizContentJSON, err := json.Marshal(bizContent)
	if err != nil {
		return "", fmt.Errorf("序列化 biz_content 失败: %w", err)
	}

	// 构建请求参数
	params := map[string]interface{}{
		"app_id":      c.AppID,
		"method":      "alipay.trade.precreate",
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

	// 证书模式下，添加证书 SN
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

	// 构建 URL（扫码支付需要发送 POST 请求）
	requestURL := c.Gateway
	requestBodyJSON, _ := json.Marshal(params)

	// 发送 POST 请求
	resp, err := c.sendPostRequest(requestURL, string(requestBodyJSON))
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}

	// 解析响应（支付宝返回的是 form 格式，需要解析）
	// 响应格式：alipay_trade_precreate_response={"code":"10000","msg":"Success",...}&sign=...
	// 需要提取 alipay_trade_precreate_response 的值
	respMap, err := parseAlipayFormResponse(resp)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应码
	responseNode, ok := respMap["alipay_trade_precreate_response"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误: 缺少 alipay_trade_precreate_response")
	}

	code, ok := responseNode["code"].(string)
	if !ok {
		return "", fmt.Errorf("响应格式错误: 缺少 code 字段")
	}

	if code != "10000" {
		msg := ""
		if m, ok := responseNode["msg"].(string); ok {
			msg = m
		}
		subCode := ""
		subMsg := ""
		if sc, ok := responseNode["sub_code"].(string); ok {
			subCode = sc
		}
		if sm, ok := responseNode["sub_msg"].(string); ok {
			subMsg = sm
		}
		return "", fmt.Errorf("%s,%s,%s", msg, subCode, subMsg)
	}

	// 返回二维码内容
	qrCode, ok := responseNode["qr_code"].(string)
	if !ok {
		return "", fmt.Errorf("响应中缺少 qr_code 字段")
	}

	return qrCode, nil
}

// sendPostRequest 发送 POST 请求（用于扫码支付等需要 POST 的接口）
func (c *Client) sendPostRequest(url, body string) (string, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 记录 query_log
	go c.createQueryLog(url, "POST", body, fmt.Sprintf("%d", resp.StatusCode), string(bodyBytes), "")

	return string(bodyBytes), nil
}

// parseAlipayFormResponse 解析支付宝 form 格式的响应
// 响应格式：alipay_trade_precreate_response={"code":"10000",...}&sign=xxx
func parseAlipayFormResponse(resp string) (map[string]interface{}, error) {
	// 解析 URL 编码的 form 数据
	values, err := url.ParseQuery(resp)
	if err != nil {
		return nil, fmt.Errorf("解析 form 数据失败: %w", err)
	}

	result := make(map[string]interface{})

	// 遍历所有键值对
	for key, val := range values {
		if len(val) == 0 {
			continue
		}

		// 尝试解析 JSON 值
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(val[0]), &jsonValue); err == nil {
			result[key] = jsonValue
		} else {
			// 如果不是 JSON，直接存储字符串
			result[key] = val[0]
		}
	}

	return result, nil
}
