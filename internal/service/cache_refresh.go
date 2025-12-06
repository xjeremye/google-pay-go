package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/internal/database"
	"github.com/golang-pay-core/internal/models"
)

// CacheRefreshService 缓存刷新服务（增量更新版本）
type CacheRefreshService struct {
	redis       *redis.Client
	cacheExpiry time.Duration
	stopChan    chan struct{}
	// 最后刷新时间戳（用于增量更新）
	lastRefreshTime time.Time
	// 刷新窗口（查询最近多少秒内更新的数据）
	refreshWindow time.Duration
}

// NewCacheRefreshService 创建缓存刷新服务
func NewCacheRefreshService() *CacheRefreshService {
	now := time.Now()
	return &CacheRefreshService{
		redis:           database.RDB,
		cacheExpiry:     24 * time.Hour, // 缓存过期时间
		stopChan:        make(chan struct{}),
		lastRefreshTime: now.Add(-2 * time.Second), // 初始化为2秒前，确保第一次查询所有数据
		refreshWindow:   2 * time.Second,           // 查询最近2秒内更新的数据
	}
}

// Start 启动缓存刷新服务
func (s *CacheRefreshService) Start(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second) // 每秒刷新一次
	defer ticker.Stop()

	// 立即执行一次全量刷新（初始化）
	s.refreshAllIncremental(ctx, true)

	for {
		select {
		case <-ticker.C:
			// 增量刷新（只刷新最近更新的数据）
			s.refreshAllIncremental(ctx, false)
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop 停止缓存刷新服务
func (s *CacheRefreshService) Stop() {
	close(s.stopChan)
}

// refreshAllIncremental 增量刷新所有热点数据缓存
func (s *CacheRefreshService) refreshAllIncremental(ctx context.Context, fullRefresh bool) {
	now := time.Now()
	refreshSince := s.lastRefreshTime

	if fullRefresh {
		// 全量刷新：查询所有数据
		refreshSince = time.Time{} // 空时间表示查询所有
	}

	// 刷新系统用户缓存
	s.refreshUsersIncremental(ctx, refreshSince)

	// 刷新商户缓存
	s.refreshMerchantsIncremental(ctx, refreshSince)

	// 刷新租户缓存
	s.refreshTenantsIncremental(ctx, refreshSince)

	// 刷新支付渠道缓存（关键数据，必须刷新）
	s.refreshPayChannelsIncremental(ctx, refreshSince)

	// 刷新插件缓存（关键数据，必须刷新）
	s.refreshPluginsIncremental(ctx, refreshSince)

	// 刷新插件配置缓存
	s.refreshPluginConfigsIncremental(ctx, refreshSince)

	// 刷新插件支付类型缓存
	s.refreshPluginPayTypesIncremental(ctx, refreshSince)

	// 更新最后刷新时间
	s.lastRefreshTime = now.Add(-500 * time.Millisecond) // 留500ms缓冲，避免遗漏
}

// refreshUsersIncremental 增量刷新系统用户缓存
func (s *CacheRefreshService) refreshUsersIncremental(ctx context.Context, since time.Time) {
	tableKey := "table:dvadmin_system_users"

	// 全量刷新时，直接查询所有用户
	if since.IsZero() {
		// 查询所有用户
		var users []struct {
			ID             int64      `gorm:"column:id"`
			Key            string     `gorm:"column:key"`
			Status         bool       `gorm:"column:status"`
			UpdateDatetime *time.Time `gorm:"column:update_datetime"`
		}

		if err := database.DB.Table("dvadmin_system_users").
			Select("id, key, status, update_datetime").
			Find(&users).Error; err != nil {
			return
		}

		// 更新所有用户的缓存
		var maxUpdateTime time.Time
		for _, user := range users {
			cacheKey := fmt.Sprintf("user:%d", user.ID)

			// 更新缓存（只缓存必要的字段）
			userData := struct {
				ID     int64  `json:"id"`
				Key    string `json:"key"`
				Status bool   `json:"status"`
			}{
				ID:     user.ID,
				Key:    user.Key,
				Status: user.Status,
			}

			if data, err := json.Marshal(userData); err == nil {
				_ = s.redis.Set(ctx, cacheKey, data, 1*time.Hour).Err()
			}

			// 记录最大的更新时间
			if user.UpdateDatetime != nil && user.UpdateDatetime.After(maxUpdateTime) {
				maxUpdateTime = *user.UpdateDatetime
			}
		}

		// 更新表的最后更新时间
		if !maxUpdateTime.IsZero() {
			s.setTableUpdateTime(ctx, tableKey, maxUpdateTime)
		} else {
			// 如果没有更新时间，使用当前时间
			s.setTableUpdateTime(ctx, tableKey, time.Now())
		}
		return
	}

	// 增量刷新：检查表的最后更新时间
	tableLastUpdate, _ := s.getTableUpdateTime(ctx, tableKey)

	// 如果表没有更新，跳过
	if !tableLastUpdate.IsZero() && !tableLastUpdate.After(since) {
		return
	}

	// 查询表的实际最后更新时间（从数据库获取）
	var maxUpdateTime time.Time
	database.DB.Table("dvadmin_system_users").
		Select("MAX(update_datetime) as max_time").
		Scan(&maxUpdateTime)

	// 如果表没有更新，跳过
	if !maxUpdateTime.IsZero() && !maxUpdateTime.After(since) {
		return
	}

	// 查询需要刷新的用户
	var users []struct {
		ID             int64      `gorm:"column:id"`
		Key            string     `gorm:"column:key"`
		Status         bool       `gorm:"column:status"`
		UpdateDatetime *time.Time `gorm:"column:update_datetime"`
	}

	query := database.DB.Table("dvadmin_system_users").
		Select("id, key, status, update_datetime").
		Where("update_datetime > ? OR update_datetime IS NULL", since)

	if err := query.Find(&users).Error; err != nil {
		return
	}

	// 更新所有用户的缓存
	for _, user := range users {
		cacheKey := fmt.Sprintf("user:%d", user.ID)

		// 更新缓存（只缓存必要的字段）
		userData := struct {
			ID     int64  `json:"id"`
			Key    string `json:"key"`
			Status bool   `json:"status"`
		}{
			ID:     user.ID,
			Key:    user.Key,
			Status: user.Status,
		}

		if data, err := json.Marshal(userData); err == nil {
			_ = s.redis.Set(ctx, cacheKey, data, 1*time.Hour).Err()
		}
	}

	// 更新表的最后更新时间
	if !maxUpdateTime.IsZero() {
		s.setTableUpdateTime(ctx, tableKey, maxUpdateTime)
	}
}

// refreshMerchantsIncremental 增量刷新商户缓存
func (s *CacheRefreshService) refreshMerchantsIncremental(ctx context.Context, since time.Time) {
	tableKey := "table:dvadmin_merchant"

	// 获取表的最后更新时间
	tableLastUpdate, _ := s.getTableUpdateTime(ctx, tableKey)

	// 如果表没有更新，跳过
	if !since.IsZero() && !tableLastUpdate.IsZero() && !tableLastUpdate.After(since) {
		return
	}

	// 查询表的实际最后更新时间（从数据库获取）
	var maxUpdateTime time.Time
	database.DB.Model(&models.Merchant{}).
		Select("MAX(update_datetime) as max_time").
		Scan(&maxUpdateTime)

	// 如果表没有更新，跳过
	if !since.IsZero() && !maxUpdateTime.IsZero() && !maxUpdateTime.After(since) {
		return
	}

	var merchants []models.Merchant
	query := database.DB.Model(&models.Merchant{})

	// 如果有时间限制，只查询最近更新的
	if !since.IsZero() {
		query = query.Where("update_datetime > ? OR update_datetime IS NULL", since)
	}

	if err := query.Find(&merchants).Error; err != nil {
		return
	}

	// 更新所有商户的缓存
	for _, merchant := range merchants {
		cacheKey := fmt.Sprintf("merchant:%d", merchant.ID)

		// 更新缓存
		if data, err := json.Marshal(merchant); err == nil {
			_ = s.redis.Set(ctx, cacheKey, data, s.cacheExpiry).Err()
		}
	}

	// 更新表的最后更新时间
	if !maxUpdateTime.IsZero() {
		s.setTableUpdateTime(ctx, tableKey, maxUpdateTime)
	}
}

// refreshTenantsIncremental 增量刷新租户缓存
func (s *CacheRefreshService) refreshTenantsIncremental(ctx context.Context, since time.Time) {
	tableKey := "table:dvadmin_tenant"

	// 获取表的最后更新时间
	tableLastUpdate, _ := s.getTableUpdateTime(ctx, tableKey)

	// 如果表没有更新，跳过
	if !since.IsZero() && !tableLastUpdate.IsZero() && !tableLastUpdate.After(since) {
		return
	}

	// 查询表的实际最后更新时间（从数据库获取）
	var maxUpdateTime time.Time
	database.DB.Model(&models.Tenant{}).
		Select("MAX(update_datetime) as max_time").
		Scan(&maxUpdateTime)

	// 如果表没有更新，跳过
	if !since.IsZero() && !maxUpdateTime.IsZero() && !maxUpdateTime.After(since) {
		return
	}

	var tenants []models.Tenant
	query := database.DB.Model(&models.Tenant{})

	if !since.IsZero() {
		query = query.Where("update_datetime > ? OR update_datetime IS NULL", since)
	}

	if err := query.Find(&tenants).Error; err != nil {
		return
	}

	// 更新所有租户的缓存
	for _, tenant := range tenants {
		cacheKey := fmt.Sprintf("tenant:%d", tenant.ID)

		if data, err := json.Marshal(tenant); err == nil {
			_ = s.redis.Set(ctx, cacheKey, data, s.cacheExpiry).Err()
		}
	}

	// 更新表的最后更新时间
	if !maxUpdateTime.IsZero() {
		s.setTableUpdateTime(ctx, tableKey, maxUpdateTime)
	}
}

// refreshPayChannelsIncremental 增量刷新支付渠道缓存（关键数据）
func (s *CacheRefreshService) refreshPayChannelsIncremental(ctx context.Context, since time.Time) {
	tableKey := "table:dvadmin_pay_channel"

	// 获取表的最后更新时间
	tableLastUpdate, _ := s.getTableUpdateTime(ctx, tableKey)

	// 如果表没有更新，跳过
	if !since.IsZero() && !tableLastUpdate.IsZero() && !tableLastUpdate.After(since) {
		return
	}

	// 查询表的实际最后更新时间（从数据库获取）
	var maxUpdateTime time.Time
	database.DB.Model(&models.PayChannel{}).
		Select("MAX(update_datetime) as max_time").
		Scan(&maxUpdateTime)

	// 如果表没有更新，跳过
	if !since.IsZero() && !maxUpdateTime.IsZero() && !maxUpdateTime.After(since) {
		return
	}

	var channels []models.PayChannel
	query := database.DB.Model(&models.PayChannel{})

	if !since.IsZero() {
		query = query.Where("update_datetime > ? OR update_datetime IS NULL", since)
	}

	if err := query.Find(&channels).Error; err != nil {
		return
	}

	// 更新所有渠道的缓存
	for _, channel := range channels {
		cacheKey := fmt.Sprintf("channel:%d", channel.ID)

		if data, err := json.Marshal(channel); err == nil {
			_ = s.redis.Set(ctx, cacheKey, data, s.cacheExpiry).Err()
		}
	}

	// 更新表的最后更新时间
	if !maxUpdateTime.IsZero() {
		s.setTableUpdateTime(ctx, tableKey, maxUpdateTime)
	}
}

// refreshPluginsIncremental 增量刷新插件缓存（关键数据）
func (s *CacheRefreshService) refreshPluginsIncremental(ctx context.Context, since time.Time) {
	tableKey := "table:dvadmin_pay_plugin"

	// 获取表的最后更新时间
	tableLastUpdate, _ := s.getTableUpdateTime(ctx, tableKey)

	// 如果表没有更新，跳过
	if !since.IsZero() && !tableLastUpdate.IsZero() && !tableLastUpdate.After(since) {
		return
	}

	// 查询表的实际最后更新时间（从数据库获取）
	var maxUpdateTime time.Time
	database.DB.Model(&models.PayPlugin{}).
		Select("MAX(update_datetime) as max_time").
		Scan(&maxUpdateTime)

	// 如果表没有更新，跳过
	if !since.IsZero() && !maxUpdateTime.IsZero() && !maxUpdateTime.After(since) {
		return
	}

	var plugins []models.PayPlugin
	query := database.DB.Model(&models.PayPlugin{})

	if !since.IsZero() {
		query = query.Where("update_datetime > ? OR update_datetime IS NULL", since)
	}

	if err := query.Find(&plugins).Error; err != nil {
		return
	}

	// 更新所有插件的缓存
	for _, plugin := range plugins {
		cacheKey := fmt.Sprintf("plugin:%d", plugin.ID)

		if data, err := json.Marshal(plugin); err == nil {
			_ = s.redis.Set(ctx, cacheKey, data, s.cacheExpiry).Err()
		}
	}

	// 更新表的最后更新时间
	if !maxUpdateTime.IsZero() {
		s.setTableUpdateTime(ctx, tableKey, maxUpdateTime)
	}
}

// refreshPluginConfigsIncremental 增量刷新插件配置缓存
func (s *CacheRefreshService) refreshPluginConfigsIncremental(ctx context.Context, since time.Time) {
	tableKey := "table:dvadmin_pay_plugin_config"

	// 获取表的最后更新时间
	tableLastUpdate, _ := s.getTableUpdateTime(ctx, tableKey)

	// 如果表没有更新，跳过
	if !since.IsZero() && !tableLastUpdate.IsZero() && !tableLastUpdate.After(since) {
		return
	}

	// 查询表的实际最后更新时间（从数据库获取）
	var maxUpdateTime time.Time
	database.DB.Model(&models.PayPluginConfig{}).
		Select("MAX(update_datetime) as max_time").
		Scan(&maxUpdateTime)

	// 如果表没有更新，跳过
	if !since.IsZero() && !maxUpdateTime.IsZero() && !maxUpdateTime.After(since) {
		return
	}

	var plugins []models.PayPlugin
	if err := database.DB.Find(&plugins).Error; err != nil {
		return
	}

	for _, plugin := range plugins {
		var configs []models.PayPluginConfig
		query := database.DB.Where("parent_id = ?", plugin.ID)

		// 如果有时间限制，只查询最近更新的配置
		if !since.IsZero() {
			query = query.Where("update_datetime > ? OR update_datetime IS NULL", since)
		}

		if err := query.Find(&configs).Error; err != nil {
			continue
		}

		// 如果有更新的配置，刷新缓存
		if len(configs) > 0 {
			cacheKey := fmt.Sprintf("plugin_config:%d", plugin.ID)
			if data, err := json.Marshal(configs); err == nil {
				_ = s.redis.Set(ctx, cacheKey, data, s.cacheExpiry).Err()
			}
		}
	}

	// 更新表的最后更新时间
	if !maxUpdateTime.IsZero() {
		s.setTableUpdateTime(ctx, tableKey, maxUpdateTime)
	}
}

// refreshPluginPayTypesIncremental 增量刷新插件支付类型缓存
// 由于关联表没有 update_datetime，采用以下策略：
// 1. 当插件（PayPlugin）更新时，刷新该插件的支付类型关联
// 2. 当支付类型（PayType）更新时，刷新所有相关插件的支付类型关联
// 3. 使用关联关系的哈希值检测关联表本身的变化
func (s *CacheRefreshService) refreshPluginPayTypesIncremental(ctx context.Context, since time.Time) {
	// 策略1：检查有更新的插件，刷新其支付类型关联
	var updatedPlugins []models.PayPlugin
	if !since.IsZero() {
		database.DB.Model(&models.PayPlugin{}).
			Where("update_datetime > ? OR update_datetime IS NULL", since).
			Find(&updatedPlugins)
	}

	// 策略2：检查有更新的支付类型，找出所有相关插件
	var updatedPayTypeIDs []int64
	if !since.IsZero() {
		var updatedPayTypes []models.PayType
		database.DB.Model(&models.PayType{}).
			Where("update_datetime > ? OR update_datetime IS NULL", since).
			Find(&updatedPayTypes)
		for _, pt := range updatedPayTypes {
			updatedPayTypeIDs = append(updatedPayTypeIDs, pt.ID)
		}
	}

	// 收集需要刷新的插件ID
	pluginIDsToRefresh := make(map[int64]bool)

	// 添加有更新的插件
	for _, plugin := range updatedPlugins {
		pluginIDsToRefresh[plugin.ID] = true
	}

	// 添加与更新支付类型相关的插件
	if len(updatedPayTypeIDs) > 0 {
		var relatedPluginIDs []int64
		database.DB.Table("dvadmin_pay_plugin_pay_types").
			Where("paytype_id IN ?", updatedPayTypeIDs).
			Pluck("payplugin_id", &relatedPluginIDs)
		for _, pluginID := range relatedPluginIDs {
			pluginIDsToRefresh[pluginID] = true
		}
	}

	// 策略3：全量刷新时，刷新所有插件的支付类型关联
	if since.IsZero() {
		var allPlugins []models.PayPlugin
		database.DB.Find(&allPlugins)
		for _, plugin := range allPlugins {
			pluginIDsToRefresh[plugin.ID] = true
		}
	}
	// 注意：关联表本身的变化（直接增删关联关系）通过策略1和2已经覆盖
	// 如果直接修改关联表，通常也会更新父表（插件或支付类型）的 update_datetime

	// 刷新需要更新的插件支付类型
	for pluginID := range pluginIDsToRefresh {
		s.refreshPluginPayTypesForPlugin(ctx, pluginID)
	}
}

// refreshPluginPayTypesForPlugin 刷新指定插件的支付类型关联
func (s *CacheRefreshService) refreshPluginPayTypesForPlugin(ctx context.Context, pluginID int64) {
	var payTypes []models.PayType
	if err := database.DB.Table("dvadmin_pay_type").
		Joins("INNER JOIN dvadmin_pay_plugin_pay_types ON dvadmin_pay_type.id = dvadmin_pay_plugin_pay_types.paytype_id").
		Where("dvadmin_pay_plugin_pay_types.payplugin_id = ?", pluginID).
		Find(&payTypes).Error; err != nil {
		return
	}

	cacheKey := fmt.Sprintf("plugin_pay_types:%d", pluginID)
	if data, err := json.Marshal(payTypes); err == nil {
		_ = s.redis.Set(ctx, cacheKey, data, s.cacheExpiry).Err()
		// 更新关联关系的哈希值
		s.setCachedRelationHash(ctx, cacheKey, payTypes)
	}
}

// calculateRelationHash 计算关联关系的哈希值（使用 MD5）
func (s *CacheRefreshService) calculateRelationHash(payTypeIDs []int64) string {
	if len(payTypeIDs) == 0 {
		return ""
	}

	// 对 payTypeIDs 排序，确保哈希值稳定
	sorted := make([]int64, len(payTypeIDs))
	copy(sorted, payTypeIDs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// 生成哈希字符串
	hashStr := ""
	for _, id := range sorted {
		hashStr += fmt.Sprintf("%d,", id)
	}

	// 使用 MD5 生成固定长度的哈希值
	hash := md5.Sum([]byte(hashStr))
	return hex.EncodeToString(hash[:])
}

// getCachedRelationHash 获取缓存中存储的关联关系哈希值
func (s *CacheRefreshService) getCachedRelationHash(ctx context.Context, cacheKey string) (string, error) {
	hashKey := cacheKey + ":hash"
	return s.redis.Get(ctx, hashKey).Result()
}

// setCachedRelationHash 设置缓存中存储的关联关系哈希值
func (s *CacheRefreshService) setCachedRelationHash(ctx context.Context, cacheKey string, payTypes []models.PayType) {
	hashKey := cacheKey + ":hash"
	payTypeIDs := make([]int64, len(payTypes))
	for i, pt := range payTypes {
		payTypeIDs[i] = pt.ID
	}
	hash := s.calculateRelationHash(payTypeIDs)
	_ = s.redis.Set(ctx, hashKey, hash, s.cacheExpiry).Err()
}

// mapKeysToSlice 将 map 的 key 转换为 slice
func mapKeysToSlice(m map[int64]bool) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getTableUpdateTime 获取表的最后更新时间
func (s *CacheRefreshService) getTableUpdateTime(ctx context.Context, tableKey string) (time.Time, error) {
	val, err := s.redis.Get(ctx, tableKey).Result()
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse(time.RFC3339Nano, val)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

// setTableUpdateTime 设置表的最后更新时间
func (s *CacheRefreshService) setTableUpdateTime(ctx context.Context, tableKey string, updateTime time.Time) {
	_ = s.redis.Set(ctx, tableKey, updateTime.Format(time.RFC3339Nano), s.cacheExpiry).Err()
}
