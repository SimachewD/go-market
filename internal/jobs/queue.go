package jobs

import (
	// "sync"

	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type JobQueue struct {
    Jobs chan uint
    DB *gorm.DB
    Workers int
    MaxRetries int
    RetryBackoff time.Duration
    DeadLetters chan uint
}

func NewJobQueue(db *gorm.DB, workers int) *JobQueue {
    return &JobQueue{
        Jobs:         make(chan uint, 100), // Buffered channel for jobs
		DeadLetters:  make(chan uint, 50), // Buffered channel for dead letters
		DB:           db,
		Workers:      workers,
		MaxRetries:   3,
		RetryBackoff: 2 * time.Second,
    }
}

func (q *JobQueue) Enqueue(orderID uint) {
    q.Jobs <- orderID
}

func (q *JobQueue) Start() {
    for i := 0; i < q.Workers; i++ {
        go q.worker(i)
    }
    go q.handleDeadLetters()
}

func (q *JobQueue) Stop(ctx context.Context) {
	fmt.Println("[JobQueue] Stopping workers...")
	close(q.Jobs)
	close(q.DeadLetters)

	select {
	case <-ctx.Done():
		fmt.Println("[JobQueue] Graceful stop complete.")
	case <-time.After(3 * time.Second):
		fmt.Println("[JobQueue] Timeout while stopping.")
	}
}

// func (q *JobQueue) worker(id int) {
//     defer q.wg.Done()    
//     orderID := <-q.Jobs
//     q.processOrder(orderID)
// }

