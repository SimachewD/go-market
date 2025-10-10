package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-market/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JobQueue struct {
    Jobs chan uint
    DB *gorm.DB
    Workers int
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
}

func NewJobQueue(db *gorm.DB, workers int) *JobQueue {
    ctx, cancel := context.WithCancel(context.Background())
    return &JobQueue{
        Jobs: make(chan uint, 100), // buffered channel
        DB: db,
        Workers: workers,
        ctx: ctx,
        cancel: cancel,
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

    for {
        select {
            case orderID, ok := <-q.Jobs:
                if !ok {
                    fmt.Printf("[Worker %d] job channel closed\n", id)
                    return
                }
                time.Sleep(30 * time.Second)
                q.processOrder(orderID)
            case <-q.ctx.Done():
                fmt.Printf("[Worker %d] stopping...\n", id)
                return
        }
    }
}

func (q *JobQueue) processOrder(orderID uint) {
    // Fetch the order first
    var order models.Order
    if err := q.DB.First(&order, orderID).Error; err != nil {
        fmt.Printf("[Worker] Order %d not found: %v\n", orderID, err)
        return
    }

    fmt.Printf("[Worker] Processing order %d (ProductID=%d, Quantity=%d)\n", order.ID, order.ProductID, order.Quantity)

    // Simulate payment delay
    time.Sleep(2 * time.Second)

    // Transaction with row-level locking
    err := q.DB.Transaction(func(tx *gorm.DB) error {
        var product models.Product

        // Lock the product row for update
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, order.ProductID).Error; err != nil {
            return fmt.Errorf("product %d not found", order.ProductID)
        }

        if product.Stock < order.Quantity {
            return fmt.Errorf("insufficient stock for product %d", product.ID)
        }

        // Deduct stock
        product.Stock -= order.Quantity
        if err := tx.Save(&product).Error; err != nil {
            return fmt.Errorf("failed to update product stock: %v", err)
        }

        // Mark order as completed
        order.Status = "completed"
        if err := tx.Save(&order).Error; err != nil {
            return fmt.Errorf("failed to update order status: %v", err)
        }

        return nil
    })

    if err != nil {
        fmt.Printf("[Worker] Failed to process order %d: %v\n", order.ID, err)
        // mark order as failed outside transaction
        order.Status = "failed"
        q.DB.Save(&order)
        return
    }

    fmt.Printf("[Worker] Order %d processed successfully\n", order.ID)
}

func (q *JobQueue) Shutdown() {
    fmt.Println("[JobQueue] Shutdown initiated...")
    close(q.Jobs)
    q.cancel()

    done := make(chan struct{})
    go func() {
        q.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        fmt.Println("[JobQueue] All workers stopped gracefully")
    case <-time.After(5 * time.Second):
        fmt.Println("[JobQueue] Timeout: Forcing shutdown")
    }
}



// using websocket example

// func (q *JobQueue) processOrder(orderID uint) {
//     time.Sleep(5 * time.Second)

//     var order models.Order
//     if err := q.DB.First(&order, orderID).Error; err != nil {
//         fmt.Println("Failed to find order", orderID)
//         return
//     }

//     var product models.Product

//     if err := q.DB.First(&product, order.ProductID).Error; err != nil {
//         return
//     }

//     if product.Stock < order.Quantity {
//         fmt.Println("insufficient stock of product")
//             return 
//         }

//         // Deduct stock
//         product.Stock -= order.Quantity
//         if err := q.DB.Save(&product).Error; err != nil {
//             fmt.Printf("Error updating stock %v", err)
//             return 
//         }

//     order.Status = "completed"
//     if err := q.DB.Save(&order).Error; err != nil {
//         fmt.Println("Failed to update order", orderID)
//         return
//     }

//     fmt.Printf("Order %d completed\n", orderID)

    // Send update via WebSocket
    // if q.WSManager != nil {
    //     q.WSManager.SendToUser(order.UserID, map[string]any{
    //         "order_id": order.ID,
    //         "status":   order.Status,
    //     })
    // }
// }
