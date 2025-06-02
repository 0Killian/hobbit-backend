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
	"github.com/stripe/stripe-go/v82"
)

// Configuration struct for Keycloak settings
type Config_T struct {
	DatabaseURL         string
	RedisURL            string
	StripeSecretKey     string
	StripeWebhookSecret string
	StripePrice1KXP     string
	Hostname            string
	Port                string
	PublicBaseUrl       string
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
		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		StripePrice1KXP:     os.Getenv("STRIPE_PRICE_1KXP"),
		Hostname:            os.Getenv("HOSTNAME"),
		Port:                os.Getenv("PORT"),
		PublicBaseUrl:       os.Getenv("PUBLIC_BASE_URL"),
	}

	publicKeyPath := os.Getenv("KEYCLOAK_PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		log.Fatal("KEYCLOAK_PUBLIC_KEY_PATH is not set")
	}

	if Config.StripeSecretKey == "" {
		log.Fatal("STRIPE_SECRET_KEY is not set")
	}

	if Config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	if Config.StripeWebhookSecret == "" {
		log.Fatal("STRIPE_WEBHOOK_SECRET is not set")
	}

	if Config.StripePrice1KXP == "" {
		log.Fatal("STRIPE_PRICE_1KXP is not set")
	}

	if Config.Hostname == "" {
		Config.Hostname = "localhost"
	}

	if Config.Port == "" {
		Config.Port = "8080"
	}

	if Config.PublicBaseUrl == "" {
		Config.PublicBaseUrl = "http://" + Config.Hostname + ":" + Config.Port
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

	stripe.Key = Config.StripeSecretKey
}
