package database

import (
	"log"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"

	"ttpos-server-go/pkg/otlp"
)

func enableGormTracing(db *gorm.DB, dbName string) {
	cfg := otlp.Config()
	if cfg == nil || !cfg.Enabled || db == nil {
		return
	}

	if err := db.Use(otelgorm.NewPlugin(
		otelgorm.WithTracerProvider(otel.GetTracerProvider()),
		otelgorm.WithDBName(dbName),
	)); err != nil {
		log.Printf("[OTLP] register gorm otel plugin failed: %v", err)
	}
}
