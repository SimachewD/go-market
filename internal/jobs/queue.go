package jobs

import (

	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type JobQueue struct {
    Jobs chan uint
    DB *gorm.DB
    Workers int
    MaxRetries int
    RetryBackoff time.Duration
    DeadLetters chan uint

	// Prometheus metrics
    JobsProcessed prometheus.Counter
    JobsFailed    prometheus.Counter
    JobsRetry     prometheus.Counter
    JobsDead      prometheus.Counter
    QueueDepth    prometheus.Gauge
}

func NewJobQueue(db *gorm.DB, workers int) *JobQueue {
	q := &JobQueue{
		Jobs:        make(chan uint, 100),
		DeadLetters: make(chan uint, 50),
		DB:          db,
		Workers:     workers,
		MaxRetries:  3,
		RetryBackoff: 2 * time.Second,
		JobsProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "jobs_processed_total",
			Help: "Total number of successfully processed jobs",
		}),
		JobsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "jobs_failed_total",
			Help: "Total number of failed jobs",
		}),
		JobsRetry: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "jobs_retry_total",
			Help: "Total number of retries attempted",
		}),
		JobsDead: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "jobs_deadletter_total",
			Help: "Total number of jobs sent to dead letter queue",
		}),
		QueueDepth: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "jobs_queue_depth",
			Help: "Current length of the job queue",
		}),
	}

	// Register metrics AFTER struct is created
	prometheus.MustRegister(q.JobsProcessed, q.JobsFailed, q.JobsRetry, q.JobsDead, q.QueueDepth)

	return q
}

func (q *JobQueue) Enqueue(orderID uint) {
    q.Jobs <- orderID
	q.QueueDepth.Set(float64(len(q.Jobs))) // update gauge after enqueue
}

func (q *JobQueue) Start() {
    for i := 0; i < q.Workers; i++ {
        go q.worker(i)
    }
    // Load and run scheduled retries
	go q.retryOrders()
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

