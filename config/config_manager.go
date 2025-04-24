package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config 初始化
func init() {
	NewConfigManager()
	CONFIG_MANAGER.LoadConfig()
}

var CONFIG_MANAGER *ConfigManager

// ConfigManager 配置管理器
type ConfigManager struct {
	v        *viper.Viper
	config   *Config
	mutex    sync.RWMutex
	onReload []func()
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	if CONFIG_MANAGER != nil {
		return CONFIG_MANAGER
	}
	CONFIG_MANAGER = &ConfigManager{
		v:        viper.New(),
		onReload: make([]func(), 0),
	}
	return CONFIG_MANAGER
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig() error {

	// 加载配置
	configPath := os.Getenv("CONFIG_PATH")

	if configPath != "" {
		cm.v.SetConfigFile(configPath)
	} else {
		// 默认配置文件查找路径
		cm.v.AddConfigPath(".")
		cm.v.AddConfigPath("config")
		cm.v.AddConfigPath("configs")
		cm.v.SetConfigName("config")
	}

	// 设置环境变量前缀
	cm.v.SetEnvPrefix("APP")
	cm.v.AutomaticEnv()

	// 读取配置文件
	if err := cm.v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置到结构体
	cm.mutex.Lock()
	cm.config = &Config{}
	if err := cm.v.Unmarshal(cm.config); err != nil {
		cm.mutex.Unlock()
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	cm.mutex.Unlock()

	// 监听配置文件变化
	cm.v.WatchConfig()
	cm.v.OnConfigChange(func(e fsnotify.Event) {
		cm.mutex.Lock()
		if err := cm.v.Unmarshal(cm.config); err != nil {
			fmt.Printf("failed to reload config: %v\n", err)
		} else {
			// 触发重载回调
			for _, fn := range cm.onReload {
				fn()
			}
		}
		cm.mutex.Unlock()
	})

	return nil
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.config
}

// OnReload 注册配置重载回调
func (cm *ConfigManager) OnReload(fn func()) {
	cm.mutex.Lock()
	cm.onReload = append(cm.onReload, fn)
	cm.mutex.Unlock()
}

// GetString 获取字符串配置
func (cm *ConfigManager) GetString(key string) string {
	return cm.v.GetString(key)
}

// GetInt 获取整数配置
func (cm *ConfigManager) GetInt(key string) int {
	return cm.v.GetInt(key)
}

// GetBool 获取布尔配置
func (cm *ConfigManager) GetBool(key string) bool {
	return cm.v.GetBool(key)
}

// Set 设置配置值
func (cm *ConfigManager) Set(key string, value interface{}) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.v.Set(key, value)
}
