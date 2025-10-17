package jobs

import (
	"fmt"
	"go-market/internal/models"
	"time"
)

func (q *JobQueue) worker(id int) {
	fmt.Printf("[Worker-%d] Started\n", id)

	for orderID := range q.Jobs {
		err := q.processOrder(orderID)
		if err == nil {
			fmt.Printf("[Worker-%d] Order %d processed successfully\n", id, orderID)
			continue
		}

		var order models.Order
		if err := q.DB.First(&order, orderID).Error; err == nil {
			order.RetryCount++
			fmt.Printf("[Worker-%d] Order %d failed (attempt %d/%d): %v\n", id, orderID, order.RetryCount, q.MaxRetries, err)

			if order.RetryCount >= q.MaxRetries {
				order.Status = "dead_letter"
				q.DB.Save(&order)
				fmt.Printf("[Worker-%d] Order %d moved to Dead-Letter Queue\n", id, orderID)
			} else {
				order.Status = "pending"
				q.DB.Save(&order)

				// Non-blocking retry with backoff
				backoff := time.Duration(order.RetryCount) * q.RetryBackoff
				go func(id uint, d time.Duration) {
					time.Sleep(d)
					q.Jobs <- id
				}(orderID, backoff)
			}
		}
	}
}
