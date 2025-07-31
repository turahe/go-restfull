package config

import (
	"sync"
)

var config *Config
var m sync.Mutex

type Config struct {
	Env        string           `yaml:"env"`
	App        App              `yaml:"app"`
	HttpServer HttpServer       `yaml:"httpServer"`
	Log        Log              `yaml:"log"`
	Scheduler  Scheduler        `yaml:"scheduler"`
	Schedules  []Schedule       `yaml:"schedules"`
	Postgres   Postgres         `yaml:"postgres"`  // Legacy single database config
	Databases  []DatabaseConfig `yaml:"databases"` // New multi-database config
	Minio      Minio            `yaml:"minio"`
	Redis      []Redis          `yaml:"redis"`
	RabbitMQ   RabbitMQ         `yaml:"rabbitmq"`
	Sentry     Sentry           `yaml:"sentry"`
	Email      Email            `yaml:"email"`
	Casbin     Casbin           `yaml:"casbin"`
	Backup     Backup           `yaml:"backup"`
}

// RabbitMQ configuration
type RabbitMQ struct {
	Enable     bool   `yaml:"enable"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	VHost      string `yaml:"vhost"`
	SSL        bool   `yaml:"ssl"`
	Connection struct {
		MaxRetries int `yaml:"maxRetries"`
		RetryDelay int `yaml:"retryDelay"` // in seconds
		Timeout    int `yaml:"timeout"`    // in seconds
	} `yaml:"connection"`
	Channel struct {
		PrefetchCount int `yaml:"prefetchCount"`
		Qos           int `yaml:"qos"`
	} `yaml:"channel"`
	Exchanges []ExchangeConfig `yaml:"exchanges"`
	Queues    []QueueConfig    `yaml:"queues"`
}

// ExchangeConfig represents a RabbitMQ exchange configuration
type ExchangeConfig struct {
	Name       string            `yaml:"name"`
	Type       string            `yaml:"type"` // direct, fanout, topic, headers
	Durable    bool              `yaml:"durable"`
	AutoDelete bool              `yaml:"autoDelete"`
	Internal   bool              `yaml:"internal"`
	Arguments  map[string]string `yaml:"arguments"`
}

// QueueConfig represents a RabbitMQ queue configuration
type QueueConfig struct {
	Name       string            `yaml:"name"`
	Durable    bool              `yaml:"durable"`
	AutoDelete bool              `yaml:"autoDelete"`
	Exclusive  bool              `yaml:"exclusive"`
	Arguments  map[string]string `yaml:"arguments"`
	Bindings   []BindingConfig   `yaml:"bindings"`
}

// BindingConfig represents a RabbitMQ binding configuration
type BindingConfig struct {
	Exchange   string `yaml:"exchange"`
	RoutingKey string `yaml:"routingKey"`
}

// DatabaseConfig represents a single database configuration
type DatabaseConfig struct {
	Name              string `yaml:"name"`
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Database          string `yaml:"database"`
	Schema            string `yaml:"schema"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	MaxConnections    int    `yaml:"maxConnections"`
	ConnectionTimeout int    `yaml:"connectionTimeout"`
	IdleTimeout       int    `yaml:"idleTimeout"`
	MaxIdleConns      int    `yaml:"maxIdleConns"`
	MaxOpenConns      int    `yaml:"maxOpenConns"`
	SSLMode           string `yaml:"sslMode"`
	IsDefault         bool   `yaml:"isDefault"`
}

type HttpServer struct {
	Port       int    `yaml:"port"`
	SwaggerURL string `yaml:"swaggerURL"`
}

type Log struct {
	Level           string `yaml:"level"`
	StacktraceLevel string `yaml:"stacktraceLevel"`
	FileEnabled     bool   `yaml:"fileEnabled"`
	FileSize        int    `yaml:"fileSize"`
	FilePath        string `yaml:"filePath"`
	FileCompress    bool   `yaml:"fileCompress"`
	MaxAge          int    `yaml:"maxAge"`
	MaxBackups      int    `yaml:"maxBackups"`
}

type Label struct {
	En string `json:"en"`
	Th string `json:"id"`
}

type App struct {
	Name                  string `yaml:"name"`
	NameSlug              string `yaml:"nameSlug"`
	JWTSecret             string `yaml:"jwtSecret"`
	AccessTokenExpiration int    `yaml:"accessTokenExpiration"`
}

type Postgres struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Database          string `yaml:"database"`
	Schema            string `yaml:"schema"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	MaxConnections    int    `yaml:"maxConnections"`
	ConnectionTimeout int    `yaml:"connectionTimeout"`
	IdleTimeout       int    `yaml:"idleTimeout"`
	MaxIdleConns      int    `yaml:"maxIdleConns"`
	MaxOpenConns      int    `yaml:"maxOpenConns"`
}

type Scheduler struct {
	Timezone string `yaml:"timezone"`
}

type Schedule struct {
	Job       string `yaml:"job"`
	Cron      string `yaml:"cron"`
	IsEnabled bool   `yaml:"isEnabled"`
}

type Minio struct {
	Enable          bool   `yaml:"enable"`
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	UseSSL          bool   `yaml:"useSSL"`
	BucketName      string `yaml:"bucket"`
	Region          string `yaml:"region"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

type Sentry struct {
	Dsn         string `yaml:"dsn"`
	Environment string `yaml:"environment"`
	Release     string `yaml:"release"`
	Debug       bool   `yaml:"debug"`
}

type Email struct {
	SMTPHost    string `yaml:"smtpHost"`
	SMTPPort    int    `yaml:"smtpPort"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	FromAddress string `yaml:"fromAddress"`
	FromName    string `yaml:"fromName"`
}

type Casbin struct {
	Model  string      `yaml:"model"`
	Policy string      `yaml:"policy"`
	Redis  CasbinRedis `yaml:"redis"`
}

type CasbinRedis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"db"`
	Key      string `yaml:"key"`
}

type Backup struct {
	Enabled        bool   `yaml:"enabled"`
	Directory      string `yaml:"directory"`
	RetentionDays  int    `yaml:"retentionDays"`
	CleanupOld     bool   `yaml:"cleanupOld"`
	CompressBackup bool   `yaml:"compressBackup"`
}

func GetConfig() *Config {
	return config
}

func SetConfig(cfg *Config) {
	config = cfg
}

// GetDatabaseConfig returns a specific database configuration by name
func GetDatabaseConfig(name string) *DatabaseConfig {
	for _, db := range config.Databases {
		if db.Name == name {
			return &db
		}
	}
	return nil
}

// GetDefaultDatabaseConfig returns the default database configuration
func GetDefaultDatabaseConfig() *DatabaseConfig {
	// First try to find a database marked as default
	for _, db := range config.Databases {
		if db.IsDefault {
			return &db
		}
	}

	// If no default is marked, return the first database
	if len(config.Databases) > 0 {
		return &config.Databases[0]
	}

	// Fallback to legacy postgres config
	return &DatabaseConfig{
		Name:           "default",
		Host:           config.Postgres.Host,
		Port:           config.Postgres.Port,
		Database:       config.Postgres.Database,
		Schema:         config.Postgres.Schema,
		Username:       config.Postgres.Username,
		Password:       config.Postgres.Password,
		MaxConnections: config.Postgres.MaxConnections,
		IsDefault:      true,
	}
}

// GetAllDatabaseConfigs returns all database configurations
func GetAllDatabaseConfigs() []DatabaseConfig {
	return config.Databases
}

// HasDatabaseConfig checks if a database configuration exists
func HasDatabaseConfig(name string) bool {
	return GetDatabaseConfig(name) != nil
}

// GetDatabaseNames returns a list of all database names
func GetDatabaseNames() []string {
	names := make([]string, len(config.Databases))
	for i, db := range config.Databases {
		names[i] = db.Name
	}
	return names
}

// GetRabbitMQConfig returns the RabbitMQ configuration
func GetRabbitMQConfig() *RabbitMQ {
	return &config.RabbitMQ
}

// IsRabbitMQEnabled checks if RabbitMQ is enabled
func IsRabbitMQEnabled() bool {
	return config.RabbitMQ.Enable
}

// GetExchangeConfig returns a specific exchange configuration by name
func GetExchangeConfig(name string) *ExchangeConfig {
	for _, exchange := range config.RabbitMQ.Exchanges {
		if exchange.Name == name {
			return &exchange
		}
	}
	return nil
}

// GetQueueConfig returns a specific queue configuration by name
func GetQueueConfig(name string) *QueueConfig {
	for _, queue := range config.RabbitMQ.Queues {
		if queue.Name == name {
			return &queue
		}
	}
	return nil
}

// GetAllExchangeConfigs returns all exchange configurations
func GetAllExchangeConfigs() []ExchangeConfig {
	return config.RabbitMQ.Exchanges
}

// GetAllQueueConfigs returns all queue configurations
func GetAllQueueConfigs() []QueueConfig {
	return config.RabbitMQ.Queues
}
