package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"server/internal/core"
	"server/internal/global"
	"server/internal/initialize"
)

func main() {
	// Load configuration first (everything depends on it)
	global.TD27_VP = core.Viper()

	// Setup logger
	global.TD27_LOG = core.Logger()

	// Initialize OpenTelemetry tracer provider (if enabled)
	tp, err := initialize.InitTracerProvider()
	if err != nil {
		global.TD27_LOG.Error("Failed to initialize tracer provider", "error", err)
		// Continue execution even if tracing fails
	}
	if tp != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				global.TD27_LOG.Error("Failed to shutdown tracer provider", "error", err)
			}
		}()
		global.TD27_TP = tp
	}

	// Initialize Database
	global.TD27_DB = initialize.Gorm()
	if global.TD27_DB == nil {
		slog.Error("database connection failed")
		os.Exit(1)
	}
	db, _ := global.TD27_DB.DB()
	defer db.Close()

	// Auto migrate tables AFTER DB is initialized
	if !global.TD27_CONFIG.System.DisableAutoMigrate {
		if err := initialize.RegisterTables(global.TD27_DB); err != nil {
			global.TD27_LOG.Error("auto migrate failed", "error", err)
			os.Exit(1)
		}
	}

	// Initialize Cron AFTER DB ready
	initialize.InitCron()

	// Build router
	router := initialize.Routers()

	// Run HTTP server
	addr := fmt.Sprintf("%s:%d", global.TD27_CONFIG.System.Host, global.TD27_CONFIG.System.Port)
	initialize.RunServer(addr, router)
}
