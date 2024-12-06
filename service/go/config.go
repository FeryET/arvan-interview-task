package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	envconfig "github.com/sethvargo/go-envconfig"
)

// AppConfig is the api config, and is configurable via environment variables.
type AppConfig struct {
	// DB Config
	DBHost         string `env:"DB_HOST, default=localhost"`
	DBPort         int    `env:"DB_PORT, default=5432"`
	DBUser         string `env:"DB_USER, default=postgres"`
	DBPassword     string `env:"DB_PASSWORD, default=postgres"`
	DBName         string `env:"DB_NAME, default=db"`
	DBTableName    string `env:"DB_TABLE_NAME, default=ip_cache"`
	DBMaxOpenConns int    `env:"DB_MAX_OPEN_CONNS, default=1024"`
	DBMaxIdleConns int    `env:"DB_MAX_IDLE_CONNS, default=512"`
	DBMaxLifeTime  int    `env:"DB_MAX_LIFETIME_SECS, default=20"`
	DBMaxIdleTime  int    `env:"DB_MAX_IDLETIME_SECS, default=10"`
	// Server Port
	ServerPort int `env:"SERVER_PORT, default=3333"`
}

// CreateDBConnection establishes and returns a new database connection using the AppConfig struct.
func (config *AppConfig) CreateDBConnection() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(time.Second * time.Duration(config.DBMaxIdleTime))
	db.SetConnMaxLifetime(time.Second * time.Duration(config.DBMaxLifeTime))
	db.SetMaxIdleConns(config.DBMaxIdleConns)
	db.SetMaxOpenConns(config.DBMaxOpenConns)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// NewAppConfig is the constructor for AppConfig which reads config from environment variables.
func NewAppConfig() (*AppConfig, error) {
	//// Init database connection
	ctx := context.Background()
	var config AppConfig
	err := envconfig.Process(ctx, &config)
	return &config, err
}
