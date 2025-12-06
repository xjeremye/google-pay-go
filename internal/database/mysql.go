package database

import (
	"fmt"
	"time"

	"github.com/golang-pay-core/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitMySQL 初始化 MySQL 连接
func InitMySQL() error {
	dsn := config.Cfg.Database.GetDSN()

	var logLevel logger.LogLevel
	if config.Cfg.Database.LogMode {
		logLevel = logger.Info
	} else {
		logLevel = logger.Silent
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层 sql.DB 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.Cfg.Database.ConnMaxLifetime)

	DB = db
	return nil
}

// CloseMySQL 关闭数据库连接
func CloseMySQL() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

