package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config contains all application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

// ServerConfig ...
type ServerConfig struct {
	Port string `mapstructure:"port"`
}

// JWTConfig ...
type JWTConfig struct {
	SecretKey      string `mapstructure:"secret_key"`
	ExpiresInHours int    `mapstructure:"expires_in_hours"`
}

// PostgresConfig ...
type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

// RedisConfig defines the configuration for Redis connection
type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LoggerConfig defines the configuration for logging
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
	FileOutput bool   `mapstructure:"file_output"` // false to disable file logging (stdout only)
}

// LoadConfig implements a loading strategy that prioritizes environment variables with config file fallback.
func LoadConfig(path string) (*Config, error) {
	// Load .env file into process environment (silently ignored if missing)
	_ = godotenv.Load()

	var config Config

	// 1. Attempt to load base configuration from config.yaml
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		// If the error is just "file not found", it can be safely ignored because environment variables may provide all configuration
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err // If it is another type of parsing error, it needs to be reported
		}
	}

	// Parse the file contents (if they exist) into the config struct first
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 2. Override file values with environment variables (manual override)
	if dsn := os.Getenv("POSTGRES_DSN"); dsn != "" {
		config.Postgres.DSN = dsn
	}
	if jwtSecret := os.Getenv("JWT_SECRET_KEY"); jwtSecret != "" {
		config.JWT.SecretKey = jwtSecret
	}
	if jwtExpires := os.Getenv("JWT_EXPIRES_IN_HOURS"); jwtExpires != "" {
		if val, err := strconv.Atoi(jwtExpires); err == nil {
			config.JWT.ExpiresInHours = val
		}
	}
	if logLevel := os.Getenv("LOGGER_LEVEL"); logLevel != "" {
		config.Logger.Level = logLevel
	}
	if fileOutput := os.Getenv("LOGGER_FILE_OUTPUT"); fileOutput == "false" {
		config.Logger.FileOutput = false
	}
	if redisAddr := os.Getenv("REDIS_ADDRESS"); redisAddr != "" {
		config.Redis.Address = redisAddr
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		if val, err := strconv.Atoi(redisDB); err == nil {
			config.Redis.DB = val
		}
	}

	// 3. Verify the final configuration is complete
	if config.Postgres.DSN == "" {
		return nil, errors.New("database connection string (POSTGRES_DSN) is not configured")
	}
	if config.JWT.SecretKey == "" {
		return nil, errors.New("JWT secret key (JWT_SECRET_KEY) is not configured")
	}
	if config.Redis.Address == "" {
		return nil, errors.New("Redis address (REDIS_ADDRESS) is not configured")
	}

	return &config, nil
}
