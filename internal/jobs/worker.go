package jobs

import (
	"fmt"
	"go-market/internal/models"
	"time"
)

func (q *JobQueue) worker(id int) {
	fmt.Printf("[Worker-%d] Started\n", id)
	
	for orderID := range q.Jobs {
		q.QueueDepth.Set(float64(len(q.Jobs))) // update gauge after dequeue
		err := q.processOrder(orderID)
		if err == nil {
			// Success
			q.JobsProcessed.Inc()
			fmt.Printf("[Worker-%d] Order %d processed successfully\n", id, orderID)
			continue
		}
		// Failure
		q.JobsFailed.Inc()

		var order models.Order
		if err := q.DB.First(&order, orderID).Error; err == nil {
			order.RetryCount++

			// Retry
			q.JobsRetry.Inc()
			fmt.Printf("[Worker-%d] Order %d failed (attempt %d/%d): %v\n", id, orderID, order.RetryCount, q.MaxRetries, err)

			if order.RetryCount >= q.MaxRetries {
				// Dead-letter
				q.JobsDead.Inc()
				order.Status = "dead_letter"
				q.DB.Save(&order)
				fmt.Printf("[Worker-%d] Order %d moved to Dead-Letter Queue\n", id, orderID)
			} else {
				order.Status = "pending"
				backoff := time.Duration(order.RetryCount) * q.RetryBackoff

				order.NextRetryAt = time.Now().Add(backoff)
				q.DB.Save(&order)
			}
		}
	}
}
