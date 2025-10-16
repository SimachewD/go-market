package jobs

import (
	"fmt"
	"go-market/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (q *JobQueue) processOrder(orderID uint) error {
    var order models.Order
    if err := q.DB.First(&order, orderID).Error; err != nil {
        return fmt.Errorf("order not found: %v", err)
    }

    fmt.Printf("[Worker] Processing order %d (ProductID=%d, Quantity=%d)\n", order.ID, order.ProductID, order.Quantity)
    time.Sleep(2 * time.Second)

    err := q.DB.Transaction(func(tx *gorm.DB) error {
        var product models.Product

        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, order.ProductID).Error; err != nil {
            return fmt.Errorf("product %d not found", order.ProductID)
        }

        if product.Stock < order.Quantity {
            return fmt.Errorf("insufficient stock")
        }

        product.Stock -= order.Quantity
        if err := tx.Save(&product).Error; err != nil {
            return err
        }

        order.Status = "completed"
        if err := tx.Save(&order).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        order.Status = "failed"
        q.DB.Save(&order)
        return err
    }

    return nil
}
