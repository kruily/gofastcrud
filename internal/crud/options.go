package crud

// QueryOptions 查询选项
type QueryOptions struct {
	// 分页
	Page     int
	PageSize int
	// 排序
	OrderBy []string
	// 查询条件
	Where map[string]interface{}
	// 预加载关系
	Preload []string
	// 选择特定字段
	Select []string
}

// DeleteOptions 删除选项
type DeleteOptions struct {
	// 是否物理删除
	Force bool
}
