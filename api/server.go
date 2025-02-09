package api

import (
	"context"

	"github.com/beka-birhanu/vinom-matchmaking/service/i"
	"google.golang.org/grpc"
)

// MatchmakingServer implements the Matchmaking gRPC service
type Server struct {
	matchMaker i.Matchmaker
	UnimplementedMatchmakingServer
}

func RegisterNewMatchmaker(grpcServer grpc.ServiceRegistrar, mm i.Matchmaker) error {
	server := &Server{
		matchMaker: mm,
	}
	RegisterMatchmakingServer(grpcServer, server)

	return nil
}

// Match handles incoming Match requests
func (s *Server) Match(ctx context.Context, req *MatchRequest) (*MatchResponse, error) {
	err := s.matchMaker.PushToQueue(ctx, req.ID, req.Rating, req.Latency)
	return &MatchResponse{}, err
}
