package config

import (
	"os"

	"gorm.io/gorm"
)

type dbConfig struct {
	DBUSER     string
	DBPASSWORD string
	DBHOST     string
	DBNAME     string
	DBPORT     string
}

type serverConfig struct {
	BACKEND_URL string
}

type AppConfig struct {
	Server   serverConfig
	DBConfig dbConfig
	DB       *gorm.DB
}

func NewConfig() *AppConfig {
	appConfig := AppConfig{

		Server: serverConfig{
			BACKEND_URL: getEnv("BACKEND_URL", ""),
		},

		DBConfig: dbConfig{
			DBUSER:     getEnv("DB_USERNAME", ""),
			DBPASSWORD: getEnv("DB_PASSWORD", ""),
			DBNAME:     getEnv("DB_NAME", ""),
			DBHOST:     getEnv("DB_HOST", ""),
			DBPORT:     getEnv("DB_PORT", ""),
		},
	}
	return &appConfig
}

func getEnv(key string, defaultval string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultval
}
