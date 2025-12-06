package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-pay-core/config"
	"github.com/golang-pay-core/internal/response"
)

// MetricsAuth Prometheus 指标端点认证中间件
// 支持两种认证方式：
// 1. Token 认证（通过 Authorization 头或查询参数）
// 2. IP 白名单（通过配置）
func MetricsAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从配置获取认证 Token
		metricsToken := config.Cfg.Monitoring.MetricsToken
		
		// 如果未配置 Token，允许访问（开发环境）
		if metricsToken == "" {
			c.Next()
			return
		}

		// 方式一：从 Authorization 头获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				if parts[1] == metricsToken {
					c.Next()
					return
				}
			}
		}

		// 方式二：从查询参数获取 Token
		token := c.Query("token")
		if token == metricsToken {
			c.Next()
			return
		}

		// 方式三：检查 IP 白名单
		if len(config.Cfg.Monitoring.MetricsIPWhitelist) > 0 {
			clientIP := c.ClientIP()
			for _, allowedIP := range config.Cfg.Monitoring.MetricsIPWhitelist {
				if allowedIP == "*" {
					c.Next()
					return
				}
				
				// 检查精确匹配
				if clientIP == allowedIP {
					c.Next()
					return
				}
				
				// 检查 CIDR 格式
				if strings.Contains(allowedIP, "/") {
					if isIPInCIDR(clientIP, allowedIP) {
						c.Next()
						return
					}
				}
			}
		}

		// 认证失败
		response.Fail(c, http.StatusUnauthorized, "未授权访问")
		c.Abort()
	}
}

// isIPInCIDR 检查 IP 是否在 CIDR 范围内
func isIPInCIDR(ip, cidr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	return ipNet.Contains(parsedIP)
}

