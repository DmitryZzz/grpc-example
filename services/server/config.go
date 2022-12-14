package main

import (
	"os"
	"strconv"
)

const (
	redisUsersKey = "users"
)

type GRPCConfig struct {
	port int
}

type PGConfig struct {
	host     string
	port     int
	username string
	password string
	db       string
}

type RedisConfig struct {
	addr     string
	password string
	db       int
}

type KafkaConfig struct {
	server string
	topic  string
}

type Config struct {
	psql  PGConfig
	grpc  GRPCConfig
	redis RedisConfig
	kafka KafkaConfig
}

// New returns a new Config struct
func NewConfig() *Config {
	return &Config{
		grpc: GRPCConfig{
			port: getEnvInt("APP_GRPC_PORT", 8001),
		},
		psql: PGConfig{
			host:     getEnv("APP_POSTGRES_HOST", "127.0.0.1"),
			port:     getEnvInt("APP_POSTGRES_PORT", 5432),
			username: getEnv("APP_POSTGRES_USERNAME", "postgres"),
			password: getEnv("APP_POSTGRES_PASSWORD", "postgres"),
			db:       getEnv("APP_POSTGRES_DB", "user_service"),
		},
		redis: RedisConfig{
			addr:     getEnv("APP_REDIS_ADDRESS", "localhost:6379"),
			password: getEnv("APP_REDIS_PASSWORD", ""),
			db:       getEnvInt("APP_REDIS_DB", 0),
		},
		kafka: KafkaConfig{
			server: getEnv("APP_KAFKA_SERVER", "localhost:9092"),
			topic:  getEnv("APP_KAFKA_TOPIC", "user_creation"),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return defaultVal
}
