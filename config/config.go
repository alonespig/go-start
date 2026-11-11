package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Server ServerConfig `mapstructure:"server"`
	CORS   CORSConfig   `mapstructure:"cors"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Log    LogConfig    `mapstructure:"log"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

// ServerConfig HTTP 服务配置，时长用带单位的字符串（如 "10s"、"1m30s"），
// 显式写成 0 表示不限制。SSE / WebSocket 的注意事项见 config.example.yaml。
type ServerConfig struct {
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
	MaxHeaderBytes    int           `mapstructure:"max_header_bytes"`
	// TrustedProxies 为可信代理 CIDR；留空表示不信任任何 X-Forwarded-For，
	// 避免客户端伪造 IP。部署在 nginx / 负载均衡后面时填写其网段。
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

// CORSConfig 跨域配置。allowed_origins 支持 "*" 或来源列表；
// allow_credentials 为 true 时会回显请求 Origin（不能与通配符同时生效）。
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	PrepareStmt     bool          `mapstructure:"prepare_stmt"`
	SlowThreshold   time.Duration `mapstructure:"slow_threshold"`
}

// DSN 返回 GORM 的 MySQL 连接串
func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.DBName)
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LogConfig 日志配置。File 为空时只输出到 stdout；
// 写入文件时按 MaxSize(MB) / MaxBackups / MaxAge(天) 自动轮转。
type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// Load 加载配置文件（默认 config.yaml），支持 GO_STARTER_ 前缀的环境变量覆盖，
// 例如 GO_STARTER_MYSQL_PASSWORD 对应 mysql.password。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	// 环境变量优先级高于配置文件，便于容器/CI 注入密码等敏感信息
	v.SetEnvPrefix("GO_STARTER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 默认值：配置文件里缺省时兜底
	v.SetDefault("app.name", "go-starter")
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.port", 8080)
	v.SetDefault("server.read_header_timeout", "10s")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("server.max_header_bytes", 1<<20)
	v.SetDefault("server.trusted_proxies", []string{})
	v.SetDefault("cors.allowed_origins", []string{"*"})
	v.SetDefault("cors.allow_credentials", false)
	v.SetDefault("mysql.host", "127.0.0.1")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.max_open_conns", 100)
	v.SetDefault("mysql.conn_max_lifetime", "1h")
	v.SetDefault("mysql.conn_max_idle_time", "10m")
	v.SetDefault("mysql.prepare_stmt", false)
	v.SetDefault("mysql.slow_threshold", "200ms")
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 7)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置校验失败: %w", err)
	}
	return &cfg, nil
}

// Validate 做启动期校验，把配置错误挡在进程启动阶段，而不是运行到一半才暴露。
func (c *Config) Validate() error {
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return fmt.Errorf("app.port 必须在 1-65535 之间，当前为 %d", c.App.Port)
	}
	if c.MySQL.Host == "" || c.MySQL.User == "" || c.MySQL.DBName == "" {
		return fmt.Errorf("mysql.host、mysql.user、mysql.dbname 不能为空")
	}
	if c.MySQL.Port <= 0 || c.MySQL.Port > 65535 {
		return fmt.Errorf("mysql.port 非法: %d", c.MySQL.Port)
	}
	if c.MySQL.MaxIdleConns < 0 || c.MySQL.MaxOpenConns < 0 {
		return fmt.Errorf("mysql.max_idle_conns、mysql.max_open_conns 不能为负数")
	}
	if c.MySQL.ConnMaxLifetime < 0 || c.MySQL.ConnMaxIdleTime < 0 {
		return fmt.Errorf("mysql.conn_max_lifetime、mysql.conn_max_idle_time 不能为负数")
	}
	if c.MySQL.SlowThreshold <= 0 {
		return fmt.Errorf("mysql.slow_threshold 必须大于 0")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr 不能为空")
	}
	if c.Server.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("server.read_header_timeout 必须大于 0")
	}
	if c.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("server.shutdown_timeout 必须大于 0")
	}
	if c.Log.MaxSize <= 0 {
		return fmt.Errorf("log.max_size 必须大于 0")
	}
	if c.Log.MaxBackups < 0 || c.Log.MaxAge < 0 {
		return fmt.Errorf("log.max_backups、log.max_age 不能为负数")
	}
	return nil
}
