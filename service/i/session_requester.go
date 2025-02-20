package i

import (
	"context"

	"github.com/google/uuid"
)

type SessionRequester interface {
	NewGame(context.Context, []uuid.UUID) error
}
