package service

import (
	"context"
	"fmt"

	general_i "github.com/beka-birhanu/vinom-common/interfaces/general"
	"github.com/beka-birhanu/vinom-matchmaking/service/i"
	"github.com/google/uuid"
)

type SessionService struct {
	sessionRequester i.SessionRequester
	logger           general_i.Logger
}

func NewSessionService(sr i.SessionRequester, l general_i.Logger) (*SessionService, error) {
	return &SessionService{
		sessionRequester: sr,
		logger:           l,
	}, nil
}

func (ss *SessionService) NewGame(IDs []uuid.UUID) {
	ss.logger.Info(fmt.Sprintf("sending new game request for players %v", IDs))
	err := ss.sessionRequester.NewGame(context.Background(), IDs)
	if err != nil {
		ss.logger.Error(fmt.Sprintf("new game request failed for players %v: %s", IDs, err))
		return
	}

	ss.logger.Info(fmt.Sprintf("new game request success for players %v: %s", IDs, err))
}
