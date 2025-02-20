package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	general_i "github.com/beka-birhanu/vinom-common/interfaces/general"
	logger "github.com/beka-birhanu/vinom-common/log"
	"github.com/beka-birhanu/vinom-matchmaking/api"
	"github.com/beka-birhanu/vinom-matchmaking/config"
	"github.com/beka-birhanu/vinom-matchmaking/infrastructure/sessiongrpc"
	"github.com/beka-birhanu/vinom-matchmaking/infrastructure/sortedqueue"
	"github.com/beka-birhanu/vinom-matchmaking/service"
	"github.com/beka-birhanu/vinom-matchmaking/service/i"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Global variables for dependencies
var (
	redisClient            *redis.Client
	matchmaker             i.Matchmaker
	appLogger              general_i.Logger
	grpcConnListener       net.Listener
	grpcServer             *grpc.Server
	sessionManagerGrpcConn *grpc.ClientConn
	sessionManager         *service.SessionService
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

func initGrpcConns() {
	var err error
	sessionmanagerAddr := fmt.Sprintf("%s:%d", config.Envs.SessionManagerHost, config.Envs.SessionManagerPort)
	sessionManagerGrpcConn, err = grpc.NewClient(sessionmanagerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating session manager gRPC connection : %v", err))
		os.Exit(1)
	}

	appLogger.Info("Created session manager gRPC connection")
}

func initSessionManager() {
	sessionLogger, err := logger.New("SESSION-MANAGER", config.ColorCyan, os.Stdout)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating session manager logger: %v", err))
		os.Exit(1)
	}

	grpcRquester, err := sessiongrpc.NewClient(sessionManagerGrpcConn, time.Duration(config.Envs.RPCTimeout)*time.Millisecond)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating session grpc requester: %v", err))
		os.Exit(1)
	}

	sessionManager, err = service.NewSessionService(grpcRquester, sessionLogger)
	if err != nil {
		appLogger.Error(fmt.Sprintf("Creating grpc session manager client: %v", err))
		os.Exit(1)
	}

	appLogger.Info("Session manager initialized")
}

func initMatchmaker(redisClient *redis.Client) {
	sortedQueue, err := sortedqueue.NewRedisUniqueSortedQueue(redisClient, 300)
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
		Handler:          sessionManager.NewGame,
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
	initGrpcConns()
	defer sessionManagerGrpcConn.Close()

	initSessionManager()
	initRedis(ctx)
	initMatchmaker(redisClient)
	initMatchmakingController()

	var err error
	grpcConnListener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", config.Envs.ProxyIP, config.Envs.GrpcPort))
	if err != nil {
		appLogger.Error(fmt.Sprintf("Listening tcp: %v", err))
		os.Exit(1)
	}

	appLogger.Info(fmt.Sprintf("Serving gRPC at %s:%d", config.Envs.ProxyIP, config.Envs.GrpcPort))
	if err := grpcServer.Serve(grpcConnListener); err != nil {
		appLogger.Error(fmt.Sprintf("Serving gRPC: %v", err))
		os.Exit(1)
	}
}
