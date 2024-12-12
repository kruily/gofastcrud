package module

import (
	"sync"

	"github.com/kruily/gofastcrud/core/database"
	"github.com/kruily/gofastcrud/pkg/config"
	"github.com/kruily/gofastcrud/pkg/utils"
)

type IModule interface{}

const (
	ServerService   = "Server"
	ConfigService   = "Config"
	DatabaseService = "Database"
	ResponseService = "Response"
	ScheduleService = "Schedule"
	CacheService    = "Cache"
	LoggerService   = "Logger"
	JwtService      = "Jwt"
	CasbinService   = "Casbin"
	EventBusService = "EventBus"
	FactoryService  = "Factory"
)

// CRUD_MODULE CRUD模组 全局变量
var CRUD_MODULE CrudModule = CrudModule{
	mu: sync.RWMutex{},
	services: map[string]IModule{
		ConfigService:   config.NewConfigManager(),
		DatabaseService: database.New(),
		ResponseService: &utils.DefaultResponseHandler{},
	},
}

// CrudModule CRUD模组
type CrudModule struct {
	mu sync.RWMutex

	services map[string]IModule
}

// GetConfig 获取当前配置
func GetCrudModule() *CrudModule {
	return &CRUD_MODULE
}

// GetService 获取服务
func (m *CrudModule) GetService(name string) IModule {
	return m.services[name]
}

// SetService 设置服务
func (m *CrudModule) SetService(name string, service IModule) {
	delete(m.services, name)
	m.services[name] = service
}

// RemoveService 移除服务
func (m *CrudModule) RemoveService(name string) {
	delete(m.services, name)
}

// WithResponse 设置Response服务
func (m *CrudModule) WithResponse(service ICrudResponse) {
	m.SetService(ResponseService, service)
}

// WithSchedule 设置Schedule服务
func (m *CrudModule) WithSchedule(service ISchedule) {
	m.SetService(ScheduleService, service)
}

// WithCache 设置Cache服务
func (m *CrudModule) WithCache(service ICache) {
	m.SetService(CacheService, service)
}

// WithLogger 设置Logger服务
func (m *CrudModule) WithLogger(service ILogger) {
	m.SetService(LoggerService, service)
}

// WithJwt 设置Jwt服务
func (m *CrudModule) WithJwt(service IJwt) {
	m.SetService(JwtService, service)
}

// WithCasbin 设置Casbin服务
func (m *CrudModule) WithCasbin(service ICasbin) {
	m.SetService(CasbinService, service)
}

// WithEventBus 设置EventBus服务
func (m *CrudModule) WithEventBus(service IEventBus) {
	m.SetService(EventBusService, service)
}

// WithServer 设置Server服务
func (m *CrudModule) WithServer(service IServer) {
	m.SetService(ServerService, service)
}
