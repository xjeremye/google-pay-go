package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-pay-core/config"
)

var RDB *redis.Client
var ctx = context.Background()

// InitRedis 初始化 Redis 连接
func InitRedis() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.Cfg.Redis.GetAddr(),
		Password:     config.Cfg.Redis.Password,
		DB:           config.Cfg.Redis.DB,
		PoolSize:     config.Cfg.Redis.PoolSize,
		MinIdleConns: config.Cfg.Redis.MinIdleConns,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接 Redis 失败: %w", err)
	}

	RDB = rdb
	return nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}

// GetContext 获取 Redis 上下文
func GetContext() context.Context {
	return ctx
}

