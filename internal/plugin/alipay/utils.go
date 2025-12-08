package alipay

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-pay-core/internal/alipay"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"github.com/golang-pay-core/internal/plugin"
	"github.com/golang-pay-core/internal/utils"
)

// buildAlipayPayParams 构建支付宝支付参数（公共逻辑）
// 根据产品账户类型和通道 extra_arg 构建 others 参数
func buildAlipayPayParams(product *models.AlipayProduct, req *plugin.CreateOrderRequest, totalAmount string) map[string]interface{} {
	others := make(map[string]interface{})

	// 根据产品账户类型添加参数
	if product.AccountType == 0 || product.AccountType == 7 {
		others["seller_id"] = product.UID
	} else if product.AccountType == 6 {
		// 分账模式
		others["settle_info"] = map[string]interface{}{
			"settle_detail_infos": []map[string]interface{}{
				{
					"amount":        totalAmount,
					"trans_in_type": "defaultSettle",
				},
			},
		}
		others["sub_merchant"] = map[string]interface{}{
			"merchant_id": product.AppID,
		}
	}

	// 获取通道的 extra_arg（B2B 模式）
	if req.Channel != nil {
		if extraArg, exists := req.Channel["extra_arg"]; exists {
			var extraArgInt int
			switch v := extraArg.(type) {
			case int:
				extraArgInt = v
			case int64:
				extraArgInt = int(v)
			case float64:
				extraArgInt = int(v)
			case *int:
				if v != nil {
					extraArgInt = *v
				}
			}

			if extraArgInt == 3 {
				// B2B 模式
				others["extend_params"] = map[string]interface{}{
					"paySolution":       "E_PAY",
					"paySolutionConfig": "{\"paySolutionScene\":\"ENTERPRISE_PAY\"}",
				}
			}
		}
	}

	return others
}

// generateSubject 生成订单主题（公共逻辑）
func generateSubject(ctx context.Context, req *plugin.CreateOrderRequest, productID string) string {
	if productID != "" {
		productIDInt, err := strconv.ParseInt(productID, 10, 64)
		if err == nil {
			var product models.AlipayProduct
			if err := database.DB.Select("subject").Where("id = ?", productIDInt).First(&product).Error; err == nil {
				if product.Subject != "" {
					// 格式化主题：替换占位符
					moneyStr := fmt.Sprintf("%.2f", float64(req.Money)/100)
					subject := product.Subject
					subject = strings.ReplaceAll(subject, "{money}", moneyStr)
					subject = strings.ReplaceAll(subject, "{order_no}", req.OrderNo)
					subject = strings.ReplaceAll(subject, "{out_order_no}", req.OutOrderNo)
					// 支持 Python 风格的格式化
					subject = strings.ReplaceAll(subject, "{{money}}", moneyStr)
					subject = strings.ReplaceAll(subject, "{{order_no}}", req.OrderNo)
					subject = strings.ReplaceAll(subject, "{{out_order_no}}", req.OutOrderNo)
					if subject != product.Subject {
						return subject
					}
				}
			}
		}
	}

	// 默认主题
	moneyStr := fmt.Sprintf("%.2f", float64(req.Money)/100)
	return fmt.Sprintf("订单支付-%s-%s元", req.OrderNo, moneyStr)
}

// getSystemConfigByPath 通过路径获取系统配置（避免循环依赖）
// path 格式：如 "alipay.inline_notify_domain"
func getSystemConfigByPath(ctx context.Context, path string) string {
	// 先尝试直接获取
	var config models.SystemConfig
	if err := database.DB.Where("key = ? AND status = ? AND parent_id IS NULL", path, true).
		First(&config).Error; err == nil {
		return parseSystemConfigValue(config.Value)
	}

	// 如果直接获取失败，尝试按点分割路径
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return ""
	}

	// 先找父配置
	var parentConfig models.SystemConfig
	if err := database.DB.Where("key = ? AND status = ? AND parent_id IS NULL", parts[0], true).
		First(&parentConfig).Error; err != nil {
		return ""
	}

	// 再找子配置
	if err := database.DB.Where("key = ? AND status = ? AND parent_id = ?", parts[1], true, parentConfig.ID).
		First(&config).Error; err != nil {
		return ""
	}

	return parseSystemConfigValue(config.Value)
}

// parseSystemConfigValue 解析系统配置的 JSON 值
func parseSystemConfigValue(valueStr string) string {
	if valueStr == "" {
		return ""
	}

	// 尝试解析 JSON
	var valueMap map[string]interface{}
	if err := json.Unmarshal([]byte(valueStr), &valueMap); err != nil {
		// 如果解析失败，尝试直接返回原始值
		return valueStr
	}

	// 尝试获取 value 字段
	if value, ok := valueMap["value"].(string); ok {
		return value
	}

	// 如果 value 不是字符串，返回空
	return ""
}

// createAlipayClient 创建支付宝客户端（公共逻辑）
// 返回 *alipay.Client，调用者可以使用其方法如 TradeWapPay、TradeAppPay 等
func createAlipayClient(product *models.AlipayProduct, notifyURL string) (*alipay.Client, error) {
	return alipay.NewClient(product, notifyURL, true)
}

// buildAuthURL 构建授权URL（公共逻辑）
// 参考 Python: get_auth_url(url, auth_key, auth_timeout)
// 生成带鉴权的URL，格式：{url}?auth_key={auth_key}&timestamp={timestamp}&sign={sign}
func buildAuthURL(baseURL, authKey string, authTimeout int) string {
	timestamp := time.Now().Unix()

	// 构建鉴权参数
	authParams := map[string]interface{}{
		"auth_key":  authKey,
		"timestamp": timestamp,
	}

	// 生成签名
	signData := make(map[string]interface{})
	for k, v := range authParams {
		if k != "sign" {
			signData[k] = v
		}
	}

	// 使用 utils.GetSign 生成签名
	_, sign := utils.GetSign(signData, authKey, nil, nil, 0)

	// 构建完整URL
	authURL := fmt.Sprintf("%s?auth_key=%s&timestamp=%d&sign=%s", baseURL, authKey, timestamp, sign)

	// 如果 auth_timeout > 0，添加过期时间参数
	if authTimeout > 0 {
		expireTime := timestamp + int64(authTimeout)
		authURL = fmt.Sprintf("%s&expire_time=%d", authURL, expireTime)
	}

	return authURL
}
