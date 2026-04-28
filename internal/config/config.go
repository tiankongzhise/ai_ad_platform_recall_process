package config

import (
	"os"
	"strconv"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Token    TokenConfig    `yaml:"token"`
	Notify   NotifyConfig   `yaml:"notify"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	DBName          string `yaml:"dbname"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type TokenConfig struct {
	Secret            string `yaml:"secret"`
	ExpiryHours       int    `yaml:"expiry_hours"`
	JWTTTLHours       int    `yaml:"jwt_ttl_hours"`       // JWT有效期（小时），默认24
	RefreshTTLHours   int    `yaml:"refresh_ttl_hours"` // RefreshToken有效期（小时），默认720（30天）
}

type NotifyConfig struct {
	TimeoutSeconds       int `yaml:"timeout_seconds"`
	RetryTimes           int `yaml:"retry_times"`
	RetryIntervalSeconds int `yaml:"retry_interval_seconds"`
}

var (
	cfg  *Config
	once sync.Once
)

func Load(path string) (*Config, error) {
	var err error
	once.Do(func() {
		var data []byte
		data, err = os.ReadFile(path)
		if err != nil {
			return
		}

		cfg = &Config{}
		err = yaml.Unmarshal(data, cfg)
		if err != nil {
			return
		}

		// 使用环境变量覆盖机密字段
		applyEnvOverrides(cfg)
	})

	return cfg, err
}

// applyEnvOverrides 使用环境变量覆盖配置中的机密字段
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := os.Getenv("TOKEN_SECRET"); v != "" {
		cfg.Token.Secret = v
	}
	// 设置默认值
	if cfg.Token.JWTTTLHours <= 0 {
		cfg.Token.JWTTTLHours = 24
	}
	if cfg.Token.RefreshTTLHours <= 0 {
		cfg.Token.RefreshTTLHours = 720 // 30天
	}
}

func Get() *Config {
	return cfg
}
