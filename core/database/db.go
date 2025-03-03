package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/kruily/gofastcrud/config"
)

// Database 数据库管理器
type Database struct {
	db     *gorm.DB
	config *config.DatabaseConfig
}

// New 创建数据库管理器实例
func New() *Database {
	return &Database{}
}

// ConfigurePool 配置连接池
func (d *Database) ConfigurePool(cfg *config.DatabaseConfig) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	// 设置最大空闲连接数
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	// 设置最大打开连接数
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	// 设置连接最大生命周期
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	// 设置连接最大空闲时间
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	// 验证连接池配置
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("连接池配置验证失败: %v", err)
	}

	return nil
}

// GetStats 获取连接池统计信息
func (d *Database) GetStats() sql.DBStats {
	sqlDB, err := d.db.DB()
	if err != nil {
		return sql.DBStats{}
	}
	return sqlDB.Stats()
}

// Init 初始化数据库连接
func (d *Database) Init(cfg *config.DatabaseConfig) error {
	var err error
	d.config = cfg

	switch cfg.Driver {
	case "mysql":
		if cfg.Charset == "" {
			cfg.Charset = "utf8mb4"
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)
		d.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port)
		d.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		d.db, err = gorm.Open(sqlite.Open(cfg.Database), &gorm.Config{})
	default:
		return fmt.Errorf("不支持的数据库类型: %s", cfg.Driver)
	}

	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池
	if err := d.ConfigurePool(cfg); err != nil {
		return err
	}

	log.Printf("Database connected successfully with pool configuration: (MaxIdleConns: %d, MaxOpenConns: %d, ConnMaxLifetime: %ds)",
		cfg.MaxIdleConns, cfg.MaxOpenConns, cfg.ConnMaxLifetime)
	return nil
}

// DB 获取 gorm.DB 实例
func (d *Database) DB() *gorm.DB {
	return d.db
}
