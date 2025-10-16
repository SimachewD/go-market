
package jobs

import (
	"fmt"
	"time"
)

func (q *JobQueue) worker(id int) {
	fmt.Printf("[Worker-%d] Started\n", id)

	for orderID := range q.Jobs {
		var attempt int
		for {
			err := q.processOrder(orderID)
			if err == nil {
				fmt.Printf("[Worker-%d] Order %d processed successfully\n", id, orderID)
				break
			}

			attempt++
			fmt.Printf("[Worker-%d] Order %d failed (attempt %d/%d): %v\n", id, orderID, attempt, q.MaxRetries, err)

			if attempt >= q.MaxRetries {
				fmt.Printf("[Worker-%d] Moving order %d to Dead-Letter Queue\n", id, orderID)
				q.DeadLetters <- orderID
				break
			}

			// Doubled backoff
			backoff := time.Duration(attempt) * q.RetryBackoff
			fmt.Printf("[Worker-%d] Retrying order %d in %v...\n", id, orderID, backoff)
			time.Sleep(backoff)
		}
	}
}
