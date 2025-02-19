package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	general_i "github.com/beka-birhanu/vinom-common/interfaces/general"
	"github.com/beka-birhanu/vinom-matchmaking/service/i"
	"github.com/google/uuid"
)

const (
	defaultPrefix           = "matchmaker"
	defaultMaxPlayer        = 2
	defaultRankTolerance    = 0
	defaultLatencyTolerance = 0
	queueRankLatencyKeyFmt  = "%s:queue:rank_%d:latency_%d"
)

var (
	ErrPlayerNotFoundInQueue = errors.New("player not found in queue")
)

type handlerFunc func(IDs []uuid.UUID)

type player struct {
	ID      string
	Rank    int32
	Latency int32
}

type Options struct {
	Prefix           string
	Handler          handlerFunc
	MaxPlayer        int32
	RankTolerance    int32
	LatencyTolerance int32
}

type Matchmaker struct {
	sortedQueue i.SortedQueue
	logger      general_i.Logger
	opts        *Options
}

func NewMatchmaker(sortedQueue i.SortedQueue, logger general_i.Logger, opts *Options) (i.Matchmaker, error) {
	if opts == nil {
		opts = &Options{
			MaxPlayer: defaultMaxPlayer,
			Prefix:    defaultPrefix,
		}
	}

	if opts.MaxPlayer <= 0 {
		opts.MaxPlayer = defaultMaxPlayer
	}

	if opts.Prefix == "" {
		opts.Prefix = defaultPrefix
	}

	if opts.RankTolerance < 0 {
		opts.RankTolerance = defaultRankTolerance
	}

	if opts.LatencyTolerance < 0 {
		opts.LatencyTolerance = defaultLatencyTolerance
	}

	return &Matchmaker{
		opts:        opts,
		sortedQueue: sortedQueue,
		logger:      logger,
	}, nil
}

func (mm *Matchmaker) PushToQueue(ctx context.Context, id string, rank int32, latency int32) error {
	mm.logger.Info(fmt.Sprintf("Adding player to queue: ID=%s Rank=%d Latency=%d", id, rank, latency))
	return mm.pushPlayerToQueue(ctx, &player{
		ID:      id,
		Rank:    rank,
		Latency: latency,
	})
}

func (mm *Matchmaker) pushPlayerToQueue(ctx context.Context, player *player) error {
	score := float64(time.Now().UnixNano())
	err := mm.sortedQueue.Enqueue(ctx, mm.queueKey(player.Rank, player.Latency), score, player.ID)
	if err != nil {
		mm.logger.Error(fmt.Sprintf("Failed to enqueue player: %s", err))
		return err
	}

	mm.logger.Info(fmt.Sprintf("Player enqueued successfully: ID=%s", player.ID))
	go mm.match(context.Background(), player.Rank, player.Latency)
	return nil
}

func (mm *Matchmaker) match(ctx context.Context, rank int32, latency int32) {
	newCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	queueKey := mm.queueKey(rank, latency)
	qLen, err := mm.sortedQueue.Count(newCtx, queueKey)
	if err != nil {
		mm.logger.Error(fmt.Sprintf("Counting queue length: %s", err))
		return
	}

	if qLen >= int64(mm.opts.MaxPlayer) {
		rawPlayers, err := mm.sortedQueue.DequeTops(newCtx, queueKey, int64(mm.opts.MaxPlayer))
		if err != nil {
			mm.logger.Error(fmt.Sprintf("Dequeing tops: %s", err))
			return
		}

		var playersIDs []uuid.UUID
		for _, raw := range rawPlayers {
			if id, err := uuid.Parse(raw); err == nil {
				playersIDs = append(playersIDs, id)
			} else {
				mm.logger.Warning(fmt.Sprintf("Non-UUID value in queue: %s", raw))
			}
		}

		if mm.opts.Handler != nil {
			mm.logger.Info(fmt.Sprintf("Match found for players: %v", playersIDs))
			go mm.opts.Handler(playersIDs)
		}
	}
}

func (mm *Matchmaker) SetMatchHandler(f func([]uuid.UUID)) {
	mm.opts.Handler = f
}

func (mm *Matchmaker) queueKey(rank int32, latency int32) string {
	return fmt.Sprintf(queueRankLatencyKeyFmt, mm.opts.Prefix, scale(rank, mm.opts.RankTolerance), scale(latency, mm.opts.LatencyTolerance))
}

func scale(value, tolerance int32) int32 {
	return value / (tolerance + 1)
}
