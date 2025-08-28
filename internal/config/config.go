package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Cache        CacheConfig        `mapstructure:"cache"`
	Queue        QueueConfig        `mapstructure:"queue"`
	Registration RegistrationConfig `mapstructure:"registration"`
	Log          LogConfig          `mapstructure:"log"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Description string `mapstructure:"description"`
}

type ServerConfig struct {
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	ReadTimeout    int    `mapstructure:"read_timeout"`
	WriteTimeout   int    `mapstructure:"write_timeout"`
	MaxHeaderBytes int    `mapstructure:"max_header_bytes"`
}

type DatabaseConfig struct {
	Driver                 string `mapstructure:"driver"`
	Host                   string `mapstructure:"host"`
	Port                   int    `mapstructure:"port"`
	Username               string `mapstructure:"username"`
	Password               string `mapstructure:"password"`
	Name                   string `mapstructure:"name"`
	SSLMode                string `mapstructure:"ssl_mode"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

type CacheConfig struct {
	Type        string         `mapstructure:"type"`
	Host        string         `mapstructure:"host"`
	Port        int            `mapstructure:"port"`
	Password    string         `mapstructure:"password"`
	DB          int            `mapstructure:"db"`
	MaxRetries  int            `mapstructure:"max_retries"`
	PoolSize    int            `mapstructure:"pool_size"`
	PoolTimeout int            `mapstructure:"pool_timeout"`
	IdleTimeout int            `mapstructure:"idle_timeout"`
	TTLMinutes  int            `mapstructure:"ttl_minutes"`
	Sentinel    SentinelConfig `mapstructure:"sentinel"`
}

type SentinelConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	MasterName       string   `mapstructure:"master_name"`
	SentinelAddrs    []string `mapstructure:"sentinel_addrs"`
	SentinelPassword string   `mapstructure:"sentinel_password"`
}

type QueueConfig struct {
	Type          string `mapstructure:"type"`
	BufferSize    int    `mapstructure:"buffer_size"`
	WorkerCount   int    `mapstructure:"worker_count"`
	RetryAttempts int    `mapstructure:"retry_attempts"`
}

type RegistrationConfig struct {
	MaxCoursesPerStudent         int    `mapstructure:"max_courses_per_student"`
	WaitlistMaxSize              int    `mapstructure:"waitlist_max_size"`
	RegistrationTimeoutMinutes   int    `mapstructure:"registration_timeout_minutes"`
	ConcurrentRegistrationsLimit int    `mapstructure:"concurrent_registrations_limit"`
	WaitlistRepository           string `mapstructure:"waitlist_repository"`
	WaitlistFallbackEnabled      bool   `mapstructure:"waitlist_fallback_enabled"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

var config *Config

func Init() {
	config = &Config{}

	setDefaults()
	if err := viper.Unmarshal(config); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}
}

func Get() *Config {
	if config == nil {
		Init()
	}
	return config
}

func setDefaults() {

	viper.SetDefault("app.name", "cobra-template")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", 15)
	viper.SetDefault("server.write_timeout", 15)
	viper.SetDefault("server.max_header_bytes", 1048576)
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "pgbouncer")
	viper.SetDefault("database.port", 6432)
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "course_registration")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime_minutes", 30)
	viper.SetDefault("cache.type", "redis")
	viper.SetDefault("cache.host", "redis-master")
	viper.SetDefault("cache.port", 6379)
	viper.SetDefault("cache.password", "")
	viper.SetDefault("cache.db", 0)
	viper.SetDefault("cache.max_retries", 3)
	viper.SetDefault("cache.pool_size", 10)
	viper.SetDefault("cache.pool_timeout", 30)
	viper.SetDefault("cache.idle_timeout", 300)
	viper.SetDefault("cache.ttl_minutes", 60)
	viper.SetDefault("cache.sentinel.enabled", true)
	viper.SetDefault("cache.sentinel.master_name", "mymaster")
	viper.SetDefault("cache.sentinel.sentinel_addrs", []string{"redis-sentinel-1:26379", "redis-sentinel-2:26379", "redis-sentinel-3:26379"})
	viper.SetDefault("cache.sentinel.sentinel_password", "")
	viper.SetDefault("queue.type", "redis")
	viper.SetDefault("queue.buffer_size", 1000)
	viper.SetDefault("queue.worker_count", 10)
	viper.SetDefault("queue.retry_attempts", 3)
	viper.SetDefault("registration.max_courses_per_student", 6)
	viper.SetDefault("registration.waitlist_max_size", 50)
	viper.SetDefault("registration.registration_timeout_minutes", 5)
	viper.SetDefault("registration.concurrent_registrations_limit", 100)
	viper.SetDefault("registration.waitlist_repository", "redis")
	viper.SetDefault("registration.waitlist_fallback_enabled", true)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.file_path", "")
}
