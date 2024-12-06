package crud

import "github.com/kruily/GoFastCrud/pkg/utils"

// ICrudResponse 定义响应处理器接口
type ICrudResponse interface {
	Success(data interface{}) interface{}
	Error(err error) interface{}
	List(items interface{}, total int64) interface{}
}

// CrudConfig 通用CRUD配置
type CrudConfig struct {
	// 响应处理器
	Responser ICrudResponse
	// 是否启用软删除
	SoftDelete bool
	// 分页配置
	DefaultPageSize int
	MaxPageSize     int
}

// 默认配置
var defaultConfig = CrudConfig{
	Responser:       &utils.DefaultResponseHandler{},
	SoftDelete:      true,
	DefaultPageSize: 10,
	MaxPageSize:     100,
}

// 全局配置实例
var globalConfig = defaultConfig

// SetConfig 设置全局配置
func SetConfig(config CrudConfig) {
	if config.Responser != nil {
		globalConfig.Responser = config.Responser
	}
	if config.DefaultPageSize > 0 {
		globalConfig.DefaultPageSize = config.DefaultPageSize
	}
	if config.MaxPageSize > 0 {
		globalConfig.MaxPageSize = config.MaxPageSize
	}
	globalConfig.SoftDelete = config.SoftDelete
}

// GetConfig 获取当前配置
func GetConfig() *CrudConfig {
	return &globalConfig
}
