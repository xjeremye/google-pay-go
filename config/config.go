package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

var Cfg *Config

// Config 应用配置结构
type Config struct {
	App     AppConfig     `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis   RedisConfig   `mapstructure:"redis"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	Log     LogConfig     `mapstructure:"log"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name         string        `mapstructure:"name"`
	Version      string        `mapstructure:"version"`
	Mode         string        `mapstructure:"mode"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	Charset         string        `mapstructure:"charset"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	LogMode         bool          `mapstructure:"log_mode"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// Load 加载配置文件
// 如果 configPath 为空，则根据环境变量 APP_ENV 自动选择配置文件
// APP_ENV 可选值: dev(默认), test, prod
func Load(configPath string) error {
	// 如果未指定配置文件路径，根据环境变量自动选择
	if configPath == "" {
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "dev" // 默认使用开发环境
		}
		
		switch env {
		case "prod", "production":
			configPath = "config/config.prod.yaml"
		case "test", "testing":
			configPath = "config/config.test.yaml"
		case "dev", "development", "":
			configPath = "config/config.yaml"
		default:
			configPath = fmt.Sprintf("config/config.%s.yaml", env)
		}
	}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)

	// 设置默认值
	setDefaults()

	// 支持环境变量覆盖配置
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败 [%s]: %w", configPath, err)
	}

	// 解析配置到结构体
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// setDefaults 设置默认值
func setDefaults() {
	viper.SetDefault("app.name", "golang-pay-core")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.mode", "release")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("log.level", "info")
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.Charset)
}

// GetAddr 获取 Redis 地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

