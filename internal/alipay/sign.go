package alipay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

// signParams 生成 RSA2 签名
func signParams(params map[string]interface{}, privateKey *rsa.PrivateKey) (string, error) {
	// 1. 将参数按 key 排序
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != nil && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 2. 构建待签名字符串
	var signContent strings.Builder
	for i, k := range keys {
		if i > 0 {
			signContent.WriteString("&")
		}
		signContent.WriteString(k)
		signContent.WriteString("=")
		signContent.WriteString(fmt.Sprintf("%v", params[k]))
	}

	// 3. 使用私钥签名
	hashed := sha256.Sum256([]byte(signContent.String()))
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	// 4. Base64 编码
	sign := base64.StdEncoding.EncodeToString(signature)
	return sign, nil
}

