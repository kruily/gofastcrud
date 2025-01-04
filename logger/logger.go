package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
)

// LogLevel 日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Writer 日志写入器接口
type Writer interface {
	Write(level LogLevel, message string, fields map[string]interface{}) error
}

// Logger 日志管理器
type Logger struct {
	zapLogger  *zap.Logger
	writers    []Writer
	fileWriter *lumberjack.Logger // 保存文件写入器的引用
	config     *Config            // 保存配置
	ticker     *time.Ticker       // 定时器
	done       chan bool          // 关闭信号
	mu         sync.Mutex         // 添加互斥锁
}

// Config 日志配置
type Config struct {
	Level        LogLevel
	FileConfig   *FileConfig
	ConsoleLevel LogLevel
	DBConfig     *DBConfig
}

// FileConfig 文件日志配置
type FileConfig struct {
	Filename   string
	MaxSize    int    // 每个日志文件的最大大小（MB）
	MaxBackups int    // 保留的旧日志文件最大数量
	MaxAge     int    // 保留的旧日志文件最大天数
	Compress   bool   // 是否压缩旧日志文件
	TimeUnit   string // 时间分割单位: "daily", "hourly"
}

// DBConfig 数据库日志配置
type DBConfig struct {
	DB        *gorm.DB
	TableName string
}

// getRotateFilename 根据时间单位获取日志文件名
func getRotateFilename(baseFilename, timeUnit string) string {
	now := time.Now()
	dir := filepath.Dir(baseFilename)
	ext := filepath.Ext(baseFilename)
	basename := strings.TrimSuffix(filepath.Base(baseFilename), ext)

	var timeFormat string
	switch timeUnit {
	case "hourly":
		timeFormat = "2006010215" // YYYYMMDDHH
	case "daily":
		timeFormat = "20060102" // YYYYMMDD
	default:
		return baseFilename
	}

	filename := fmt.Sprintf("%s.%s%s", basename, now.Format(timeFormat), ext)
	return filepath.Join(dir, filename)
}

// NewLogger 创建新的日志管理器
func NewLogger(config Config) (*Logger, error) {
	// 创建基础配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建多个core
	var cores []zapcore.Core

	// 添加控制台输出
	if config.ConsoleLevel >= 0 {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			zapcore.Level(config.ConsoleLevel),
		)
		cores = append(cores, consoleCore)
	}

	// 添加文件输出
	var fileWriter *lumberjack.Logger
	if config.FileConfig != nil {
		if err := os.MkdirAll(filepath.Dir(config.FileConfig.Filename), 0744); err != nil {
			return nil, fmt.Errorf("can't create log directory: %v", err)
		}

		filename := config.FileConfig.Filename
		if config.FileConfig.TimeUnit != "" {
			filename = getRotateFilename(filename, config.FileConfig.TimeUnit)
		}

		fileWriter = &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    config.FileConfig.MaxSize,
			MaxBackups: config.FileConfig.MaxBackups,
			MaxAge:     config.FileConfig.MaxAge,
			Compress:   config.FileConfig.Compress,
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(fileWriter),
			zapcore.Level(config.Level),
		)
		cores = append(cores, fileCore)
	}

	// 创建zap logger
	logger := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	l := &Logger{
		zapLogger: logger,
		writers:   make([]Writer, 0),
		config:    &config,
		done:      make(chan bool),
	}

	// 添加数据库writer
	if config.DBConfig != nil {
		dbWriter := NewDBWriter(config.DBConfig.DB, config.DBConfig.TableName)
		l.AddWriter(dbWriter)
	}

	// 如果配置了按时间切割，启动定时器
	if config.FileConfig != nil && config.FileConfig.TimeUnit != "" && fileWriter != nil {
		l.fileWriter = fileWriter
		l.startRotationTimer()
	}

	return l, nil
}

// AddWriter 添加自定义writer
func (l *Logger) AddWriter(writer Writer) {
	l.writers = append(l.writers, writer)
}

// write 写入日志到所有writer
func (l *Logger) write(level LogLevel, message string, fields map[string]interface{}) {
	for _, writer := range l.writers {
		if err := writer.Write(level, message, fields); err != nil {
			l.Error("Failed to write log", map[string]interface{}{"error": err.Error()})
		}
	}
}

// Debug 输出Debug级别日志
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.zapLogger.Debug(message, fieldsToZapFields(fields)...)
	l.write(DebugLevel, message, fields)
}

// Info 输出Info级别日志
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.zapLogger.Info(message, fieldsToZapFields(fields)...)
	l.write(InfoLevel, message, fields)
}

// Warn 输出Warn级别日志
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.zapLogger.Warn(message, fieldsToZapFields(fields)...)
	l.write(WarnLevel, message, fields)
}

// Error 输出Error级别日志
func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.zapLogger.Error(message, fieldsToZapFields(fields)...)
	l.write(ErrorLevel, message, fields)
}

// Fatal 输出Fatal级别日志
func (l *Logger) Fatal(message string, fields map[string]interface{}) {
	l.zapLogger.Fatal(message, fieldsToZapFields(fields)...)
	l.write(FatalLevel, message, fields)
}

// fieldsToZapFields 转换字段格式
func fieldsToZapFields(fields map[string]interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

// startRotationTimer 启动定时器检查文件切换
func (l *Logger) startRotationTimer() {
	var interval time.Duration

	switch l.config.FileConfig.TimeUnit {
	case "hourly":
		now := time.Now()
		next := now.Add(time.Hour).Truncate(time.Hour)
		interval = next.Sub(now)
	case "daily":
		now := time.Now()
		next := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
		interval = next.Sub(now)
	default:
		return
	}

	l.ticker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-l.ticker.C:
				l.rotateFile()
				// 重新计算下一个间隔
				l.resetRotationTimer()
			case <-l.done:
				l.ticker.Stop()
				return
			}
		}
	}()
}

// resetRotationTimer 重置定时器间隔
func (l *Logger) resetRotationTimer() {
	var interval time.Duration

	switch l.config.FileConfig.TimeUnit {
	case "hourly":
		interval = time.Hour
	case "daily":
		interval = 24 * time.Hour
	default:
		return
	}

	l.ticker.Reset(interval)
}

// rotateFile 切换日志文件
func (l *Logger) rotateFile() {
	if l.fileWriter == nil || l.config.FileConfig == nil {
		return
	}

	l.mu.Lock() // 添加互斥锁
	defer l.mu.Unlock()

	// 获取新的文件名
	newFilename := getRotateFilename(l.config.FileConfig.Filename, l.config.FileConfig.TimeUnit)

	// 更新文件名
	l.fileWriter.Filename = newFilename

	// 触发切换
	l.fileWriter.Rotate()
}

// Close 关闭日志管理器
func (l *Logger) Close() error {
	if l.ticker != nil {
		l.ticker.Stop() // 先停止定时器
		close(l.done)   // 关闭通道
	}

	// 关闭所有writer
	for _, writer := range l.writers {
		if closer, ok := writer.(io.Closer); ok {
			closer.Close()
		}
	}

	// 同步缓存
	return l.zapLogger.Sync()
}
