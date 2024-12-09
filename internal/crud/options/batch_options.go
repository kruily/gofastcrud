package options

// BatchOptions 批量操作选项
type BatchOptions struct {
	BatchSize int  // 每批次处理数量
	Async     bool // 是否异步处理
}
