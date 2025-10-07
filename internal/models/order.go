package models

import (
    "gorm.io/gorm"
)

type Order struct {
    gorm.Model
    UserID    uint   `gorm:"not null"`
    ProductID uint   `gorm:"not null"`
    Quantity  int    `gorm:"not null"`
    Status    string `gorm:"default:'pending'"` // pending, processing, completed, failed
}