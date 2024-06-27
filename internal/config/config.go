package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Env = initConfig()

type Config struct {
	DBConnStr  string
	ServerAddr string
}

func initConfig() *Config {
	log.Println("Loading envs...")

	if err := godotenv.Load(); err != nil {
		log.Fatalln("Error loading .env file: ", err)
	}

	return &Config{
		DBConnStr:  getEnv("DB_CONN_STR", "user=postgres password=secret host=localhost port=5432 dbname=postgres sslmode=disable"),
		ServerAddr: getEnv("SERVER_ADDR", "localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
