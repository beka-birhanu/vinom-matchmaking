package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	logger "github.com/beka-birhanu/vinom-api/infrastruture/log"
	general_i "github.com/beka-birhanu/vinom-interfaces/general"
	"github.com/beka-birhanu/vinom-matchmaking/api"
	"github.com/beka-birhanu/vinom-matchmaking/config"
	"github.com/beka-birhanu/vinom-matchmaking/infrastructure"
	"github.com/beka-birhanu/vinom-matchmaking/service"
	"github.com/beka-birhanu/vinom-matchmaking/service/i"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

// Global variables for dependencies
var (
	redisClient      *redis.Client
	matchmaker       i.Matchmaker
	appLogger        general_i.Logger
	grpcConnListener net.Listener
	grpcServer       *grpc.Server
)

// Initialization functions
func initRedis(ctx context.Context) {
	addr := fmt.Sprintf("%s:%v", config.Envs.RedisHost, config.Envs.RedisPort)

	redisClient = redis.NewClient(&redis.Options{Addr: addr, DB: 0})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		appLogger.Error(fmt.Sprintf("Failed to connect to Redis: %v", err))
		os.Exit(1)
	}
	appLogger.Info("Connected to Redis")
}

func initMatchmaker(redisClient *redis.Client) {
	sortedQueue, err := infrastructure.NewRedisUniqueSortedQueue(redisClient, 300)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating Redis sorted queue: %v", err))
		os.Exit(1)
	}

	matchLogger, err := logger.New("MATCH-MAKER", config.ColorPurple, os.Stdout)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating matchmaker logger: %v", err))
		os.Exit(1)
	}
	options := &service.Options{
		MaxPlayer:        config.Envs.MaxPlayer,
		RankTolerance:    config.Envs.RankTolerance,
		LatencyTolerance: config.Envs.LatencyTolerance,
	}

	matchmaker, err = service.NewMatchmaker(sortedQueue, matchLogger, options)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating matchmaker: %v", err))
		os.Exit(1)
	}
	appLogger.Info("Matchmaker initialized")
}

func initMatchmakingController() {
	grpcServer = grpc.NewServer()
	err := api.RegisterNewMatchmaker(grpcServer, matchmaker)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating and Registering matchmaking controller: %v", err))
		os.Exit(1)
	}
	appLogger.Info("Matchmaking controller initialized")
}

// TODO: add socket monitoring.
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer func() {
		cancel()
		redisClient.Close()
	}()

	// Initialize dependencies
	appLogger, _ = logger.New("APP", config.ColorGreen, os.Stdout)
	initRedis(ctx)
	initMatchmaker(redisClient)
	initMatchmakingController()

	var err error
	grpcConnListener, err = net.Listen("tcp", fmt.Sprintf("%s:%v", config.Envs.HostIP, config.Envs.GrpcPort))
	if err != nil {
		appLogger.Error(fmt.Sprintf("Listening tcp: %v", err))
		os.Exit(1)
	}

	if err := grpcServer.Serve(grpcConnListener); err != nil {
		appLogger.Error(fmt.Sprintf("Serving gRPC: %v", err))
		os.Exit(1)
	}
}
