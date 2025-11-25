package analytics

import (
	"runtime"
	"time"

	"gorm.io/gorm"
)

// StartSystemMetricsCollector starts a goroutine that collects system metrics periodically
func StartSystemMetricsCollector(metrics *PrometheusMetrics, db *gorm.DB) {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Collect Go runtime metrics
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			metrics.MemoryUsage.Set(float64(m.Alloc))
			metrics.GoroutinesCount.Set(float64(runtime.NumGoroutine()))

			// Collect database connection pool metrics
			if db != nil {
				sqlDB, err := db.DB()
				if err == nil {
					stats := sqlDB.Stats()
					metrics.UpdateDBConnections(stats.Idle, stats.InUse, stats.MaxOpenConnections)
				}
			}
		}
	}()
}
