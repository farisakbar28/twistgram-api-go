package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Config struct {
	AppEnv            string
	Port              string
	CORSAllowOrigins  string
	DatabaseURL       string
	SupabaseURL       string
	SupabaseSecretKey string
	SupabaseJWTSecret string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		Port:              getEnv("PORT", "8080"),
		CORSAllowOrigins:  getEnv("CORS_ALLOW_ORIGINS", "*"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseSecretKey: getEnv("SUPABASE_SECRET_KEY", ""),
		SupabaseJWTSecret: getEnv("SUPABASE_JWT_SECRET", ""),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required in .env file")
	}

	return cfg
}

func InitDatabase(cfg *Config) *gorm.DB {
	logMode := logger.Warn
	if cfg.AppEnv == "development" {
		logMode = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connected successfully")
	DB = db
	return db
}

func GetDB() *gorm.DB {
	return DB
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
