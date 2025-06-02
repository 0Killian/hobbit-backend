package common

import (
	go_context "context"
	"crypto/rsa"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// Configuration struct for Keycloak settings
type Config_T struct {
	DatabaseURL         string
	RedisURL            string
	StripePublicKey     string
	StripeWebhookSecret string
}

var (
	Config    *Config_T
	PublicKey *rsa.PublicKey
	Db        *sql.DB
	Rdb       *redis.Client
	Ctx       = go_context.Background()
)

func init() {
	godotenv.Load()
	Config = &Config_T{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		StripePublicKey:     os.Getenv("STRIPE_PUBLIC_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}

	publicKeyPath := os.Getenv("KEYCLOAK_PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		log.Fatal("KEYCLOAK_PUBLIC_KEY_PATH is not set")
	}

	if Config.StripePublicKey == "" {
		log.Fatal("STRIPE_PUBLIC_KEY is not set")
	}

	if Config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	if Config.StripeWebhookSecret == "" {
		log.Fatal("STRIPE_WEBHOOK_SECRET is not set")
	}

	if Config.RedisURL != "" {
		redisConfig, err := redis.ParseURL(Config.RedisURL)
		if err != nil {
			log.Fatal(err)
		}
		// Connections to render key value store usually takes 3-4 seconds to connect
		redisConfig.ReadTimeout = 6 * time.Second

		Rdb = redis.NewClient(redisConfig)

		log.Printf("Connected to Redis at %s", Config.RedisURL)

		_, err = Rdb.Ping(Ctx).Result()
		if err != nil {
			log.Fatal(err)
		}
	}

	var err error
	Db, err = sql.Open("postgres", Config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	PublicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		log.Fatal(err)
	}
}
