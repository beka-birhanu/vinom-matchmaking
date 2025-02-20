package sessiongrpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
)

type clientAdapter struct {
	client     SessionClient
	rpcTimeout time.Duration
}

func NewClient(cc grpc.ClientConnInterface, rt time.Duration) (*clientAdapter, error) {
	client := NewSessionClient(cc)
	return &clientAdapter{
		client:     client,
		rpcTimeout: rt,
	}, nil
}

func (c *clientAdapter) NewGame(ctx context.Context, IDs []uuid.UUID) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.rpcTimeout)
	defer cancel()

	pIDs := make([]string, 0)
	for _, id := range IDs {
		pIDs = append(pIDs, id.String())
	}

	request := &NewGameRequest{
		PlayerIDs: pIDs,
	}

	_, err := c.client.NewGame(timeoutCtx, request)
	return err
}
