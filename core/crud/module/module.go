package module

type IModule interface{}

const (
	ServerService   = "Server"
	ConfigService   = "Config"
	GormService     = "Gorm"
	MongoService    = "Mongo"
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
