package repository

import "gorm.io/gorm"

type Where func(*gorm.DB) *gorm.DB
