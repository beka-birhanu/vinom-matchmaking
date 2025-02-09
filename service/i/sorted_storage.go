package i

import "context"

// SortedQueue defines an interface for managing sorted queues.
type SortedQueue interface {
	// Enqueue adds a member to the sorted queue with a given score.
	Enqueue(ctx context.Context, queueKey string, score float64, member string) error

	// DequeTops removes and retrieves up to `amount` members with the lowest scores from the queue.
	DequeTops(ctx context.Context, queueKey string, amount int64) ([]string, error)

	// Count returns the number of members in the sorted queue.
	Count(ctx context.Context, queueKey string) (int64, error)
}
