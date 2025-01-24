package repository

import "gorm.io/gorm"

type Where func(*gorm.DB) *gorm.DB
type With func(*gorm.DB) *gorm.DB
