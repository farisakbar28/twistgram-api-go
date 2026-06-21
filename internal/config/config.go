package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Config struct {
	DatabaseURL     string
	SupabaseURL     string
	SupabaseJWTSecret string
	Port            string
}

func LoadConfig() *Config {
	// Load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseJWTSecret: getEnv("SUPABASE_JWT_SECRET", ""),
		Port:              getEnv("PORT", "8080"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required in .env file")
	}

	return cfg
}

func InitDatabase(cfg *Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test connection
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

// GetDSN returns a formatted DSN for logging purposes (without password)
func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgresql://%s@%s/%s", c.SupabaseURL, c.Port, "twistgram")
}
