package alipay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/logger"
	"github.com/golang-pay-core/internal/models"
	"go.uber.org/zap"
)

// VerifyNotify 验证支付宝回调签名
// 参考 Python: 验证支付宝回调的签名
func VerifyNotify(params map[string]string, publicKey *rsa.PublicKey) bool {
	// 获取签名
	sign, ok := params["sign"]
	if !ok || sign == "" {
		return false
	}

	// 移除 sign 和 sign_type 字段，构建待验证字符串
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建待验证字符串
	var signContent strings.Builder
	for i, k := range keys {
		if i > 0 {
			signContent.WriteString("&")
		}
		signContent.WriteString(k)
		signContent.WriteString("=")
		signContent.WriteString(params[k])
	}

	// Base64 解码签名
	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		logger.Logger.Warn("解码签名失败", zap.Error(err))
		return false
	}

	// 验证签名
	hashed := sha256.Sum256([]byte(signContent.String()))
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		logger.Logger.Warn("验证签名失败", zap.Error(err))
		return false
	}

	return true
}

// ParseNotifyParams 解析支付宝回调参数
// 参考 Python: 解析支付宝回调的订单信息
func ParseNotifyParams(params map[string]string) (*NotifyData, error) {
	// 解析 JSON 格式的响应（支付宝回调可能是 JSON 格式）
	var notifyData NotifyData

	// 尝试从 params 中获取订单号
	outTradeNo, ok := params["out_trade_no"]
	if !ok {
		// 可能是 JSON 格式，尝试解析
		if tradeNoJSON, ok := params["biz_content"]; ok {
			var bizContent map[string]interface{}
			if err := json.Unmarshal([]byte(tradeNoJSON), &bizContent); err == nil {
				if no, ok := bizContent["out_trade_no"].(string); ok {
					outTradeNo = no
				}
			}
		}
	}

	if outTradeNo == "" {
		return nil, fmt.Errorf("缺少订单号")
	}

	notifyData.OutTradeNo = outTradeNo
	notifyData.TradeNo = params["trade_no"]
	notifyData.TradeStatus = params["trade_status"]

	// 解析金额
	if totalAmount, ok := params["total_amount"]; ok {
		// 金额是字符串格式（元），需要转换为分
		var amount float64
		if _, err := fmt.Sscanf(totalAmount, "%f", &amount); err == nil {
			notifyData.TotalAmount = int(amount * 100) // 转换为分
		}
	}

	return &notifyData, nil
}

// NotifyData 支付宝回调数据
type NotifyData struct {
	OutTradeNo  string // 商户订单号
	TradeNo     string // 支付宝交易号
	TradeStatus string // 交易状态
	TotalAmount int    // 金额（分）
}

// GetAlipayProductByID 根据产品ID获取支付宝产品信息
func GetAlipayProductByID(productID string) (*models.AlipayProduct, error) {
	productIDInt, err := strconv.ParseInt(productID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("产品ID格式错误: %w", err)
	}

	var product models.AlipayProduct
	if err := database.DB.Where("id = ?", productIDInt).First(&product).Error; err != nil {
		return nil, fmt.Errorf("产品不存在: %w", err)
	}

	return &product, nil
}
