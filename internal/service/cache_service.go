package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
	"gorm.io/gorm"
)

// CacheService 缓存服务，用于优化数据获取性能
type CacheService struct {
	redis *redis.Client
}

// NewCacheService 创建缓存服务
func NewCacheService() *CacheService {
	return &CacheService{
		redis: database.RDB,
	}
}

// GetMerchantWithUser 获取商户及其用户信息（带缓存）
func (s *CacheService) GetMerchantWithUser(ctx context.Context, merchantID int64) (*models.Merchant, *SystemUser, error) {
	cacheKey := fmt.Sprintf("merchant:%d", merchantID)

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var merchant models.Merchant
		if err := json.Unmarshal([]byte(val), &merchant); err == nil {
			// 获取用户信息
			if merchant.SystemUserID != nil {
				user, err := s.GetUser(ctx, *merchant.SystemUserID)
				if err == nil {
					return &merchant, user, nil
				}
			}
		}
	}

	// 从数据库获取
	var merchant models.Merchant
	if err := database.DB.First(&merchant, merchantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("商户不存在")
		}
		return nil, nil, err
	}

	// 缓存商户信息
	if data, err := json.Marshal(merchant); err == nil {
		s.redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	// 获取用户信息
	var user *SystemUser
	if merchant.SystemUserID != nil {
		user, _ = s.GetUser(ctx, *merchant.SystemUserID)
	}

	return &merchant, user, nil
}

// GetUser 获取系统用户信息（带缓存）
func (s *CacheService) GetUser(ctx context.Context, userID int64) (*SystemUser, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var user SystemUser
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			return &user, nil
		}
	}

	// 从数据库获取
	var user SystemUser
	if err := database.DB.Table("dvadmin_system_users").
		First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}

	// 缓存用户信息
	if data, err := json.Marshal(user); err == nil {
		s.redis.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return &user, nil
}

// GetTenantWithUser 获取租户及其用户信息（带缓存）
func (s *CacheService) GetTenantWithUser(ctx context.Context, tenantID int64) (*models.Tenant, *SystemUser, error) {
	cacheKey := fmt.Sprintf("tenant:%d", tenantID)

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var tenant models.Tenant
		if err := json.Unmarshal([]byte(val), &tenant); err == nil {
			// 获取用户信息
			var systemUserID *int64
			database.DB.Table("dvadmin_tenant").
				Where("id = ?", tenantID).
				Select("system_user_id").
				Scan(&systemUserID)

			if systemUserID != nil {
				user, err := s.GetUser(ctx, *systemUserID)
				if err == nil {
					return &tenant, user, nil
				}
			}
		}
	}

	// 从数据库获取
	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("租户不存在")
		}
		return nil, nil, err
	}

	// 缓存租户信息
	if data, err := json.Marshal(tenant); err == nil {
		s.redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	// 获取用户信息
	var systemUserID *int64
	database.DB.Table("dvadmin_tenant").
		Where("id = ?", tenantID).
		Select("system_user_id").
		Scan(&systemUserID)

	var user *SystemUser
	if systemUserID != nil {
		user, _ = s.GetUser(ctx, *systemUserID)
	}

	return &tenant, user, nil
}

// GetPayChannel 获取支付渠道信息（带缓存）
func (s *CacheService) GetPayChannel(ctx context.Context, channelID int64) (*models.PayChannel, error) {
	cacheKey := fmt.Sprintf("channel:%d", channelID)

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var channel models.PayChannel
		if err := json.Unmarshal([]byte(val), &channel); err == nil {
			return &channel, nil
		}
	}

	// 从数据库获取
	var channel models.PayChannel
	if err := database.DB.First(&channel, channelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("渠道不存在")
		}
		return nil, err
	}

	// 缓存渠道信息
	if data, err := json.Marshal(channel); err == nil {
		s.redis.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return &channel, nil
}

// GetAvailableDomains 获取可用域名列表（带缓存，按上游类型分类）
// upstream: 5=支付宝, 6=微信, 0=全部
func (s *CacheService) GetAvailableDomains(ctx context.Context, upstream int) ([]models.PayDomain, error) {
	var cacheKey string
	switch upstream {
	case 5:
		cacheKey = "domains:alipay"
	case 6:
		cacheKey = "domains:wechat"
	default:
		cacheKey = "domains:all"
	}

	// 尝试从缓存获取
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var domains []models.PayDomain
		if err := json.Unmarshal([]byte(val), &domains); err == nil {
			return domains, nil
		}
	}

	// 从数据库获取
	var domains []models.PayDomain
	query := database.DB.Model(&models.PayDomain{}).Where("status = ?", true)

	// 根据上游类型过滤
	if upstream == 6 {
		// 微信
		query = query.Where("wechat_status = ?", true)
	} else if upstream == 5 {
		// 支付宝
		query = query.Where("pay_status = ?", true)
	}
	// else: 其他类型，只检查 status

	if err := query.Find(&domains).Error; err != nil {
		return nil, err
	}

	// 缓存域名列表（5分钟过期，因为域名状态可能变化）
	if data, err := json.Marshal(domains); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return domains, nil
}

// GetTenantBalance 获取租户余额（带短期缓存，与缓存刷新服务同步）
// 缓存刷新服务每秒更新一次余额，这里使用1秒缓存确保一致性
func (s *CacheService) GetTenantBalance(ctx context.Context, tenantID int64) (balance int64, preTax int64, trust bool, err error) {
	cacheKey := fmt.Sprintf("tenant_balance:%d", tenantID)

	// 尝试从缓存获取（1秒过期，与缓存刷新服务同步）
	if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var balanceData struct {
			Balance int64 `json:"balance"`
			PreTax  int64 `json:"pre_tax"`
			Trust   bool  `json:"trust"`
		}
		if err := json.Unmarshal([]byte(val), &balanceData); err == nil {
			return balanceData.Balance, balanceData.PreTax, balanceData.Trust, nil
		}
	}

	// 从数据库获取（降级方案：如果缓存未命中）
	var tenant models.Tenant
	if err := database.DB.Select("id, balance, pre_tax, trust").
		Where("id = ?", tenantID).
		First(&tenant).Error; err != nil {
		return 0, 0, false, err
	}

	// 缓存余额信息（1秒过期，与缓存刷新服务同步）
	balanceData := struct {
		Balance int64 `json:"balance"`
		PreTax  int64 `json:"pre_tax"`
		Trust   bool  `json:"trust"`
	}{
		Balance: tenant.Balance,
		PreTax:  int64(tenant.PreTax),
		Trust:   tenant.Trust,
	}
	if data, err := json.Marshal(balanceData); err == nil {
		s.redis.Set(ctx, cacheKey, data, 1*time.Second)
	}

	return tenant.Balance, int64(tenant.PreTax), tenant.Trust, nil
}

// SystemUser 系统用户模型（用于查询）
type SystemUser struct {
	ID       int64  `json:"id"`
	Key      string `json:"key"`
	Status   bool   `json:"status"`
	IsActive bool   `json:"is_active"`
}
