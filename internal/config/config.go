package config

import (
	"time"
)

type Config struct {
	DB     DBConfig
	App    AppConfig
	Server ServerConfig
	JWT    JWTConfig
	Redis  RedisConfig
}

type ServerConfig struct {
	Port            string        `env:"PORT, default=2025"`
	EnableSwagger   bool          `env:"ENABLE_SWAGGER, default=true"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT, default=5s"`
}

type AppConfig struct {
	AdminEmail    string `env:"ADMIN_EMAIL, default=admin@payterbank.app"`
	AdminPassword string `env:"ADMIN_PASSWORD, default=admin"`
	Environment   string `env:"ENVIRONMENT, default=dev"`
}

type JWTConfig struct {
	Secret   string        `env:"JWT_SECRET, required"`
	Expiry   time.Duration `env:"JWT_EXPIRY, default=24h"`
	Issuer   string        `env:"JWT_ISSUER, default=payter-bank"`
	Audience string        `env:"JWT_AUDIENCE, default=payter-bank"`
}

type DBConfig struct {
	Host     string `env:"DATABASE_HOST"`
	Port     string `env:"DATABASE_PORT"`
	Username string `env:"DATABASE_USERNAME"`
	Password string `env:"DATABASE_PASSWORD"`
	Database string `env:"DATABASE_NAME"`
	DSN      string `env:"DB_DSN, required"`
}

type RedisConfig struct {
	Addr string `env:"REDIS_ADDR"`
}
