package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"server/internal/global"
)

// OTelMiddleware wraps the otelgin middleware with route filtering
func OTelMiddleware() gin.HandlerFunc {
	// Only return middleware if tracing is enabled
	if !global.TD27_CONFIG.Observability.Otel.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	otelMiddleware := otelgin.Middleware(global.TD27_CONFIG.Observability.Otel.ServiceName,
		otelgin.WithFilter(func(r *http.Request) bool {
			path := r.URL.Path
			// Exclude health, metrics, and swagger routes from tracing
			return path != "/health" && path != global.TD27_CONFIG.Observability.Prometheus.MetricsPath && !strings.HasPrefix(path, "/swagger/")
		}),
	)

	return otelMiddleware
}
