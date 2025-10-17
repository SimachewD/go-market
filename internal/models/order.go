package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
    gorm.Model
    UserID    uint   `gorm:"not null"`
    ProductID uint   `gorm:"not null"`
    Quantity  int    `gorm:"not null"`
    Status    string `gorm:"default:'pending'"` // pending, processing, completed, failed
    RetryCount  int       // number of retry attempts
    NextRetryAt time.Time // optional: schedule next retry
}