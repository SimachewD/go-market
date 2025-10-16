
package jobs

import (
	"fmt"
	"time"

	"go-market/internal/models"
)

func (q *JobQueue) handleDeadLetters() {
	for orderID := range q.DeadLetters {
		fmt.Printf("[DLQ] Received failed order %d. Logging for review...\n", orderID)

		var order models.Order
		if err := q.DB.First(&order, orderID).Error; err == nil {
			order.Status = "dead_letter"
			q.DB.Save(&order)
		}

		time.Sleep(1 * time.Second) // simulate slow processing
	}
}
