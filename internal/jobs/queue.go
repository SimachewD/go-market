package jobs

import (
    "fmt"
    "time"

    "go-market/internal/models"

    "gorm.io/gorm"
)

type JobQueue struct {
    Jobs chan uint
    DB *gorm.DB
    Workers int
    quit chan struct{}
}

func NewJobQueue(db *gorm.DB, workers int) *JobQueue {
    return &JobQueue{
        Jobs: make(chan uint, 100), // buffered channel
        DB: db,
        Workers: workers,
        quit: make(chan struct{}),
    }
}

func (q *JobQueue) Enqueue(orderID uint) {
    q.Jobs <- orderID
}

func (q *JobQueue) Start() {
    for i := 0; i < q.Workers; i++ {
        go q.worker(i)
    }
}

func (q *JobQueue) worker(id int) {
    for {
        select {
        case orderID := <-q.Jobs:
            fmt.Printf("Worker %d processing order %d\n", id, orderID)
            q.processOrder(orderID)
        case <-q.quit:
            fmt.Printf("Worker %d quitting\n", id)
            return
        }
    }
}

func (q *JobQueue) processOrder(orderID uint) {
    // Simulate payment processing
    time.Sleep(2 * time.Second)

    var order models.Order
    if err := q.DB.First(&order, orderID).Error; err != nil {
        fmt.Println("Failed to find order", orderID)
        return
    }

    order.Status = "completed"
    if err := q.DB.Save(&order).Error; err != nil {
        fmt.Println("Failed to update order", orderID)
    } else {
        fmt.Printf("Order %d completed\n", orderID)
    }
}

func (q *JobQueue) Stop() {
    close(q.quit)
}
