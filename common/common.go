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
	ClientID     string
	ClientSecret string
	KeycloakURL  string
	Realm        string
	DatabaseURL  string
	RedisURL     string
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
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		KeycloakURL:  os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisURL:     os.Getenv("REDIS_URL"),
	}

	publicKeyPath := os.Getenv("KEYCLOAK_PUBLIC_KEY_PATH")
	/*if config.ClientID == "" {
		log.Fatal("KEYCLOAK_CLIENT_ID is not set")
	}

	if config.ClientSecret == "" {
		log.Fatal("KEYCLOAK_CLIENT_SECRET is not set")
	}

	if config.KeycloakURL == "" {
		log.Fatal("KEYCLOAK_URL is not set")
	}

	if config.Realm == "" {
		log.Fatal("KEYCLOAK_REALM is not set")
	}*/
	if publicKeyPath == "" {
		log.Fatal("KEYCLOAK_PUBLIC_KEY_PATH is not set")
	}

	if Config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
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
