package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/kruily/GoFastCrud/internal/config"
)

// Database 数据库管理器
type Database struct {
	db     *gorm.DB
	models []interface{}
}

// New 创建数据库管理器实例
func New() *Database {
	return &Database{
		models: make([]interface{}, 0),
	}
}

// RegisterModels 注册需要迁移的模型
func (d *Database) RegisterModels(models ...interface{}) {
	d.models = append(d.models, models...)
}

// Init 初始化数据库连接
func (d *Database) Init(cfg *config.DatabaseConfig) error {
	var err error

	switch cfg.Driver {
	case "mysql":
		if cfg.Charset == "" {
			cfg.Charset = "utf8mb4"
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)
		d.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
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
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 自动迁移表结构
	if err := d.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %v", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// AutoMigrate 执行自动迁移
func (d *Database) AutoMigrate() error {
	if d.db == nil {
		return fmt.Errorf("database not initialized")
	}
	if len(d.models) == 0 {
		log.Println("No models to migrate")
		return nil
	}
	return d.db.AutoMigrate(d.models...)
}

// DB 获取 gorm.DB 实例
func (d *Database) DB() *gorm.DB {
	return d.db
}
