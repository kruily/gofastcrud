package logger

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// LogEntry 日志条目
type LogEntry struct {
	ID        uint      `gorm:"primarykey"`
	Level     string    `gorm:"size:10;not null"`
	Message   string    `gorm:"type:text;not null"`
	Fields    string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
}

// DBWriter 数据库日志写入器
type DBWriter struct {
	db        *gorm.DB
	tableName string
}

// NewDBWriter 创建数据库日志写入器
func NewDBWriter(db *gorm.DB, tableName string) *DBWriter {
	if tableName == "" {
		tableName = "system_logs"
	}

	// 自动迁移
	db.Table(tableName).AutoMigrate(&LogEntry{})

	return &DBWriter{
		db:        db,
		tableName: tableName,
	}
}

// Write 写入日志到数据库
func (w *DBWriter) Write(level LogLevel, message string, fields map[string]interface{}) error {
	levelStr := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[level]

	entry := LogEntry{
		Level:     levelStr,
		Message:   message,
		Fields:    fieldsToString(fields),
		CreatedAt: time.Now(),
	}

	return w.db.Table(w.tableName).Create(&entry).Error
}

// fieldsToString 将字段映射转换为JSON字符串
func fieldsToString(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}
	bytes, err := json.Marshal(fields)
	if err != nil {
		return ""
	}
	return string(bytes)
}
