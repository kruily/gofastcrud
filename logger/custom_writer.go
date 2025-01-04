package logger

// CustomWriter 自定义写入器示例
type CustomWriter struct {
	output func(level LogLevel, message string, fields map[string]interface{})
}

// NewCustomWriter 创建自定义写入器
func NewCustomWriter(output func(level LogLevel, message string, fields map[string]interface{})) *CustomWriter {
	return &CustomWriter{output: output}
}

// Write 实现Writer接口
func (w *CustomWriter) Write(level LogLevel, message string, fields map[string]interface{}) error {
	w.output(level, message, fields)
	return nil
}
