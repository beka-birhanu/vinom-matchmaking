package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration values.
type Config struct {
	ProxyIP            string // Host IP for the server
	RedisHost          string // Hostname or IP address for the Redis server
	RedisPort          int32  // Port number for the Redis server
	MaxPlayer          int32  // Maximum number of players allowed in a game
	RankTolerance      int32  // Tolerance for player rank difference during matchmaking
	LatencyTolerance   int32  // Tolerance for latency (in milliseconds) during matchmaking
	GrpcPort           int    // Port for the GRPC server
	SessionManagerHost string // Hostname or IP address for the session manager server
	SessionManagerPort int    // Port number for the session manager server
	RPCTimeout         int    // Timeout duration for rpc calles
}

// Envs holds the application's configuration loaded from environment variables.
var Envs = initConfig()

// initConfig initializes and returns the application configuration.
// It loads environment variables from a .env file.
func initConfig() Config {
	// Load .env file if available
	if err := godotenv.Load(); err != nil {
		log.Printf("[APP] [INFO] .env file not found or could not be loaded: %v", err)
	}

	// Populate the Config struct with required environment variables
	return Config{
		ProxyIP:            mustGetEnv("PROXY_IP"),
		RedisHost:          mustGetEnv("REDIS_HOST"),
		RedisPort:          int32(mustGetEnvAsInt("REDIS_PORT")),
		MaxPlayer:          int32(mustGetEnvAsInt("MAX_PLAYER")),
		RankTolerance:      int32(mustGetEnvAsInt("RANK_TOLERANCE")),
		LatencyTolerance:   int32(mustGetEnvAsInt("LATENCY_TOLERANCE")),
		GrpcPort:           mustGetEnvAsInt("GRPC_PORT"),
		SessionManagerHost: mustGetEnv("SESSION_HOST"),
		SessionManagerPort: mustGetEnvAsInt("SESSION_PORT"),
		RPCTimeout:         mustGetEnvAsInt("RPC_TIMEOUT"),
	}
}

// mustGetEnv retrieves the value of an environment variable or logs a fatal error if not set.
func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("[APP] [FATAL] Environment variable %s is not set", key)
	}
	return value
}

// mustGetEnvAsInt retrieves the value of an environment variable as an integer or logs a fatal error if not set or cannot be parsed.
func mustGetEnvAsInt(key string) int {
	valueStr := mustGetEnv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatalf("[APP] [FATAL] Environment variable %s must be an integer: %v", key, err)
	}
	return value
}
