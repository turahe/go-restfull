package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var config *Config
var m sync.Mutex

type Config struct {
	Env        string     `yaml:"env"`
	App        App        `yaml:"app"`
	HttpServer HttpServer `yaml:"httpServer"`
	Log        Log        `yaml:"log"`
	Scheduler  Scheduler  `yaml:"scheduler"`
	Schedules  []Schedule `yaml:"schedules"`
	Postgres   Postgres   `yaml:"postgres"`
	Minio      Minio      `yaml:"minio"`
	Redis      []Redis    `yaml:"redis"`
	Sentry     Sentry     `yaml:"sentry"`
	Email      Email      `yaml:"email"`
}

type HttpServer struct {
	Port int `yaml:"port"`
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
	Name      string `yaml:"name"`
	NameSlug  string `yaml:"nameSlug"`
	JWTSecret string `yaml:"jwtSecret"`
}

type Postgres struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Schema          string `yaml:"schema"`
	MaxConnections  int32  `yaml:"maxConnections"`
	MaxConnIdleTime int32  `yaml:"maxConnIdleTime"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
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

type Sentry struct {
	Dsn         string `yaml:"dsn"`
	Environment string `yaml:"environment"`
	Release     string `yaml:"release"`
	Debug       bool   `yaml:"debug"`
}

type Scheduler struct {
	Timezone string `yaml:"timezone"`
}

type Schedule struct {
	Job       string `yaml:"job"`
	Cron      string `yaml:"cron"`
	IsEnabled bool   `yaml:"isEnabled"`
}

type Authentication struct {
	Endpoint string `yaml:"endpoint"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Email struct {
	SMTPHost    string `yaml:"smtpHost"`
	SMTPPort    int    `yaml:"smtpPort"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	FromAddress string `yaml:"fromAddress"`
	FromName    string `yaml:"fromName"`
}

func GetConfig() *Config {
	return config
}

func SetConfig(configFile string) {
	m.Lock()
	defer m.Unlock()

	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error getting config file, %s", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Println("Unable to decode into struct, ", err)
	}
}
