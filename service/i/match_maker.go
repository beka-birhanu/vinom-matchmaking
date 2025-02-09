package i

import (
	"context"
)

type Matchmaker interface {
	PushToQueue(ctx context.Context, id string, rating int32, latency int32) error
}
