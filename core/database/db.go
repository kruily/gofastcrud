package database

import (
	"context"
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
	"github.com/qiniu/qmgo"
)

// Database 数据库管理器
type Database struct {
	db      *gorm.DB
	mdb     *qmgo.Database
	mClient *qmgo.Client
	config  []config.DatabaseConfig
}

// New 创建数据库管理器实例
func New(cfg []config.DatabaseConfig) *Database {
	obj := &Database{}
	var err error
	obj.config = cfg

	for _, c := range cfg {
		switch c.Driver {
		case "mysql":
			if c.Charset == "" {
				c.Charset = "utf8mb4"
			}
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
				c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
			obj.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		case "postgres":
			dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
				c.Host, c.Username, c.Password, c.Database, c.Port)
			obj.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		case "sqlite":
			obj.db, err = gorm.Open(sqlite.Open(c.Database), &gorm.Config{})
		case "mongo":
			dsn := fmt.Sprintf("mongodb://%s:%d", c.Host, c.Port)
			// 连接 MongoDB

			client, err1 := qmgo.NewClient(context.Background(), &qmgo.Config{
				Uri: dsn,
				Auth: &qmgo.Credential{
					Username: c.Username,
					Password: c.Password,
				},
			})
			if err1 != nil {
				panic(fmt.Errorf("连接 MongoDB 失败: %v", err1))
			}
			// 选择数据库
			obj.mClient = client
			obj.mdb = client.Database(c.Database)
		default:
			panic(fmt.Errorf("不支持的数据库类型: %s", c.Driver))
		}

		if err != nil {
			panic(fmt.Errorf("连接数据库失败: %v", err))
		}

		if c.Driver == "mongo" {
			continue
		}
		// gorm 配置连接池
		if err := obj.ConfigurePool(&c); err != nil {
			panic(err)
		}
		log.Printf("Database connected successfully with pool configuration: (MaxIdleConns: %d, MaxOpenConns: %d, ConnMaxLifetime: %ds)",
			c.MaxIdleConns, c.MaxOpenConns, c.ConnMaxLifetime)
	}

	return obj
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

// GetStats 获取gorm 连接池统计信息
func (d *Database) GetStats() sql.DBStats {
	sqlDB, err := d.db.DB()
	if err != nil {
		return sql.DBStats{}
	}
	return sqlDB.Stats()
}

// DB 获取 gorm.DB 实例
func (d *Database) DB() *gorm.DB {
	return d.db
}

func (d *Database) MDB() *qmgo.Database {
	return d.mdb
}

// Close 关闭数据库连接
func (d *Database) Close() (err error) {
	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return fmt.Errorf("获取数据库实例失败: %v", err)
		}
		err = sqlDB.Close()
	}
	if d.mdb != nil {
		err = d.mClient.Close(context.Background())
	}
	return
}
