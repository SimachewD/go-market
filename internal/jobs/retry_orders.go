package jobs

import (
	"go-market/internal/models"
	"time"
)

func (q *JobQueue) retryOrders() {
	for {
		var orders []models.Order
		// Find orders that are due for retry
		if err := q.DB.Where("status = ? AND next_retry_at <= ?", "pending", time.Now()).
			Find(&orders).Error; err == nil {
			for _, o := range orders {
				q.Enqueue(o.ID)
			}
		}
		time.Sleep(30 * time.Second) // check every 30 seconds
	}
}