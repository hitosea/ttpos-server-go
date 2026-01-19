package persistence

import "gorm.io/gorm"

type DBOption func(*gorm.DB) *gorm.DB
