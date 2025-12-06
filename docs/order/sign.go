package order

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// combineValues 组合参数值，移除 sign 字段，按 key 排序，拼接成 key=value 格式，最后加上 key={key}
func combineValues(params map[string]interface{}, key string) string {
	// 创建新的 map，移除 sign 字段
	filtered := make(map[string]interface{})
	for k, v := range params {
		if k != "sign" && v != nil {
			filtered[k] = v
		}
	}

	// 按 key 排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接参数
	combinedValue := make([]string, 0, len(filtered)+1)
	for _, k := range keys {
		combinedValue = append(combinedValue, fmt.Sprintf("%s=%v", k, filtered[k]))
	}
	combinedValue = append(combinedValue, fmt.Sprintf("key=%s", key))

	return strings.Join(combinedValue, "&")
}

// md5Encryption MD5 加密并转大写
func md5Encryption(text string) string {
	hash := md5.Sum([]byte(text))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// toSign 标准签名方法
func toSign(data map[string]interface{}, key string) (string, string) {
	combinedValue := combineValues(data, key)
	encryptedValue := md5Encryption(combinedValue)
	return combinedValue, encryptedValue
}

// yiSign 兼容模式签名方法
func yiSign(params map[string]interface{}, key string) (string, string) {
	// 过滤掉 sign、sign_type 和空值
	type kv struct {
		key   string
		value interface{}
	}
	pairs := make([]kv, 0)
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != nil && v != "" {
			pairs = append(pairs, kv{key: k, value: v})
		}
	}

	// 按 key 排序
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	// 拼接参数
	urlEncodedParams := make([]string, 0, len(pairs))
	for _, p := range pairs {
		urlEncodedParams = append(urlEncodedParams, fmt.Sprintf("%s=%v", p.key, p.value))
	}
	signString := strings.Join(urlEncodedParams, "&") + key

	// MD5 加密
	md5Hash := md5Encryption(signString)
	return signString, md5Hash
}

// defaultUseList 默认使用的字段列表
var defaultUseList = []string{"mchId", "channelId", "mchOrderNo", "amount", "notifyUrl", "jumpUrl"}

// GetSign 根据 compatible 参数选择签名方法
func GetSign(data map[string]interface{}, key string, useList []string, optionalArgs []string, compatible int) (string, string) {
	if compatible == 1 {
		return yiSign(data, key)
	}

	if useList == nil {
		useList = defaultUseList
	}
	if optionalArgs == nil {
		optionalArgs = []string{}
	}

	// 构建签名数据
	da := make(map[string]interface{})
	for _, i := range useList {
		if val, ok := data[i]; ok {
			da[i] = val
		} else {
			// 如果字段不存在，使用空字符串（Python 中会 raise ValueError，这里返回空签名）
			return "", ""
		}
	}

	for _, j := range optionalArgs {
		if val, ok := data[j]; ok {
			da[j] = val
		}
	}

	return toSign(da, key)
}

// ToSign 生成响应签名
func (s *Helper) ToSign(data RawCreateOrderResponse, orderCtx *OrderCreateCtx) string {
	// 将 RawCreateOrderResponse 转换为 map[string]interface{}
	dataMap := make(map[string]interface{})
	if data.MchOrderNo != "" {
		dataMap["mchOrderNo"] = data.MchOrderNo
	}
	if data.PayOrderID != "" {
		dataMap["payOrderId"] = data.PayOrderID
	}
	if data.PayURL2 != "" {
		dataMap["payUrl"] = data.PayURL2
	}

	// 生成签名
	_, sign := GetSign(dataMap, orderCtx.SignKey, nil, nil, orderCtx.Compatible)
	return sign
}
