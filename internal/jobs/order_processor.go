package jobs

import (
	"fmt"
	"go-market/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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