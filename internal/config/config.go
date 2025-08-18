package config

import (
	"log"

	"github.com/spf13/viper"
)


type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Log      LogConfig      `mapstructure:"log"`
}


type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}


type ServerConfig struct {
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	ReadTimeout    int    `mapstructure:"read_timeout"`
	WriteTimeout   int    `mapstructure:"write_timeout"`
	MaxHeaderBytes int    `mapstructure:"max_header_bytes"`
}


type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}


type CacheConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}


type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
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
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "cobra_template")
	viper.SetDefault("database.ssl_mode", "disable")

	
	viper.SetDefault("cache.type", "redis")
	viper.SetDefault("cache.host", "localhost")
	viper.SetDefault("cache.port", 6379)
	viper.SetDefault("cache.password", "")
	viper.SetDefault("cache.db", 0)

	
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
}
