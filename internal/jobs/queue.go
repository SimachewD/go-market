package jobs

import (
	"sync"
	"time"

	"gorm.io/gorm"
)

type JobQueue struct {
    Jobs chan uint
    DB *gorm.DB
    Workers int
    wg         sync.WaitGroup
}

func NewJobQueue(db *gorm.DB, workers int) *JobQueue {
    return &JobQueue{
        Jobs: make(chan uint, 100), // buffered channel
        DB: db,
        Workers: workers,
    }
}

func (q *JobQueue) Enqueue(orderID uint) {
    q.Jobs <- orderID
}

func (q *JobQueue) Start() {
    for i := 0; i < q.Workers; i++ {
        q.wg.Add(1)
        go q.worker(i)
    }
}

func (q *JobQueue) worker(id int) {
    defer q.wg.Done()
    
    orderID := <-q.Jobs
    time.Sleep(30 * time.Second)
    q.processOrder(orderID)
}


// func (q *JobQueue) Shutdown() {
//     fmt.Println("[JobQueue] Shutdown initiated...")
//     close(q.Jobs)
//     // q.cancel()

//     done := make(chan struct{})
//     go func() {
//         q.wg.Wait()
//         close(done)
//     }()

//     select {
//     case <-done:
//         fmt.Println("[JobQueue] All workers stopped gracefully")
//         q.cancel()
//     case <-time.After(10 * time.Second):
//         fmt.Println("[JobQueue] Timeout: Forcing shutdown")
//         q.cancel()
//     }
// }
