package config

import (
	"encoding/json"

	"github.com/spf13/viper"
)

type (
	AppConfig struct {
		Server        ServerConfig
		Logger        LoggerConfig
		Database      DatabaseConfig
		AuditDatabase DatabaseConfig
		Cache         CacheConfig
		NATS          NATSConfig
		Auth          AuthConfig
		Otel          OtelConfig
	}

	ServerConfig struct {
		Host      string
		Port      int
		RateLimit int
	}

	LoggerConfig struct {
		Level string
	}

	DatabaseConfig struct {
		Driver          string
		Host            string
		Port            int
		User            string
		Password        string
		Name            string
		SSLMode         string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
		ConnMaxIdleTime int
	}

	CacheConfig struct {
		Addr string
	}

	NATSConfig struct {
		URLs string
	}

	AuthConfig struct {
		JWT    JWTConfig
		Google GoogleConfig
		Cors   CorsConfig
	}

	JWTConfig struct {
		Secret      string
		ExpiryHours int
	}

	GoogleConfig struct {
		ClientID     string
		ClientSecret string
	}

	CorsConfig struct {
		AllowedOrigins   []string
		AllowedMethods   []string
		AllowedHeaders   []string
		ExposedHeaders   []string
		AllowCredentials bool
	}

	OtelConfig struct {
		ExporterEndpoint string
		ServiceName      string
		ServiceVersion   string
	}
)

func LoadConfig() (*AppConfig, error) {
	v := viper.New()

	v.BindEnv("server.host", "APP_API_HOST")
	v.BindEnv("server.port", "APP_API_PORT")
	v.BindEnv("server.ratelimit", "APP_API_RATE_LIMIT")
	v.BindEnv("logger.level", "APP_LOGGER_LEVEL")
	v.BindEnv("database.driver", "APP_DB_DRIVER")
	v.BindEnv("database.host", "APP_DB_HOST")
	v.BindEnv("database.port", "APP_DB_PORT")
	v.BindEnv("database.user", "APP_DB_USER")
	v.BindEnv("database.password", "APP_DB_PASSWORD")
	v.BindEnv("database.name", "APP_DB_NAME")
	v.BindEnv("database.sslmode", "APP_DB_SSL_MODE")
	v.BindEnv("database.maxopenconns", "APP_DB_MAXOPENCONNS")
	v.BindEnv("database.maxidleconns", "APP_DB_MAXIDLECONNS")
	v.BindEnv("database.connmaxlifetime", "APP_DB_CONNMAXLIFETIME")
	v.BindEnv("database.connmaxidletime", "APP_DB_CONNMAXIDLETIME")

	v.BindEnv("auditdatabase.driver", "APP_AUDIT_DB_DRIVER")
	v.BindEnv("auditdatabase.host", "APP_AUDIT_DB_HOST")
	v.BindEnv("auditdatabase.port", "APP_AUDIT_DB_PORT")
	v.BindEnv("auditdatabase.user", "APP_AUDIT_DB_USER")
	v.BindEnv("auditdatabase.password", "APP_AUDIT_DB_PASSWORD")
	v.BindEnv("auditdatabase.name", "APP_AUDIT_DB_NAME")
	v.BindEnv("auditdatabase.sslmode", "APP_AUDIT_DB_SSL_MODE")
	v.BindEnv("auditdatabase.maxopenconns", "APP_AUDIT_DB_MAXOPENCONNS")
	v.BindEnv("auditdatabase.maxidleconns", "APP_AUDIT_DB_MAXIDLECONNS")
	v.BindEnv("auditdatabase.connmaxlifetime", "APP_AUDIT_DB_CONNMAXLIFETIME")
	v.BindEnv("auditdatabase.connmaxidletime", "APP_AUDIT_DB_CONNMAXIDLETIME")

	v.BindEnv("cache.addr", "APP_CACHE_ADDR")
	v.BindEnv("nats.urls", "APP_NATS_URLS")
	v.BindEnv("auth.jwt.secret", "APP_AUTH_JWT_SECRET")
	v.BindEnv("auth.jwt.expiryhours", "APP_AUTH_JWT_EXPIRYHOURS")
	v.BindEnv("auth.cors.allowedorigins", "APP_AUTH_CORS_ALLOWEDORIGINS")
	v.BindEnv("auth.cors.allowedmethods", "APP_AUTH_CORS_ALLOWEDMETHODS")
	v.BindEnv("auth.cors.allowedheaders", "APP_AUTH_CORS_ALLOWEDHEADERS")
	v.BindEnv("auth.cors.exposedheaders", "APP_AUTH_CORS_EXPOSEDHEADERS")
	v.BindEnv("auth.cors.allowcredentials", "APP_AUTH_CORS_ALLOWCREDENTIALS")

	v.BindEnv("otel.exporterendpoint", "APP_OTEL_EXPORTER_OTLP_ENDPOINT")
	v.BindEnv("otel.servicename", "APP_OTEL_SERVICE_NAME")
	v.BindEnv("otel.serviceversion", "APP_VERSION")

	// defaults...
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("logger.level", "info")
	v.SetDefault("server.ratelimit", 100)
	v.SetDefault("database.maxOpenConns", 10)
	v.SetDefault("database.maxIdleConns", 10)
	v.SetDefault("database.connMaxLifetime", 5)
	v.SetDefault("database.connMaxIdleTime", 5)
	v.SetDefault("nats.urls", "nats://localhost:4222")
	v.SetDefault("otel.servicename", "redtogreen-api")

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *AppConfig) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
