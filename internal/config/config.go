package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var Env = initConfig()

type Config struct {
	PostgresConnStr          string
	ServerAddr               string
	AccessTokenSecret        string
	RefreshTokenSecret       string
	AccessTokenExpiryInSecs  int64
	RefreshTokenExpiryInSecs int64
}

func initConfig() *Config {
	log.Println("Loading envs...")

	if err := godotenv.Load(); err != nil {
		log.Fatalln("Error loading .env file: ", err)
	}

	return &Config{
		PostgresConnStr: getEnvAsStr(
			"POSTGRES_CONN_STR",
			"user=postgres password=secret host=localhost port=5432 dbname=postgres sslmode=disable",
		),
		ServerAddr: getEnvAsStr("SERVER_ADDR",
			"localhost:8080"),
		AccessTokenSecret: getEnvAsStr(
			"ACCESS_TOKEN_SECRET",
			"secret",
		),
		RefreshTokenSecret: getEnvAsStr(
			"REFRESH_TOKEN_SECRET",
			"secret",
		),
		AccessTokenExpiryInSecs: getEnvAsInt(
			"ACCESS_TOKEN_EXPIRY_IN_SECS",
			15*24*7,
		),
		RefreshTokenExpiryInSecs: getEnvAsInt(
			"REFRESH_TOKEN_EXPIRY_IN_SECS",
			720*24*7,
		),
	}
}

func getEnvAsStr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}

		return i
	}
	return fallback
}
