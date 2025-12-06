package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-pay-core/internal/database"
)

// Cache 缓存工具
type Cache struct{}

// Set 设置缓存
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	return database.RDB.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := database.RDB.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Delete 删除缓存
func (c *Cache) Delete(ctx context.Context, key string) error {
	return database.RDB.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := database.RDB.Exists(ctx, key).Result()
	return count > 0, err
}

// GetCacheKey 生成缓存键
func GetCacheKey(prefix string, keys ...string) string {
	key := prefix
	for _, k := range keys {
		key += ":" + k
	}
	return key
}

