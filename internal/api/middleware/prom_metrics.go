package middleware

import (
	"chat/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

var customMetrics = []*ginprometheus.Metric{
	{
		Name:        "response_time",
		Description: "Response time histogram",
		Type:        "histogram_vec",
		Args:        []string{"status", "url", "handler"},
	},
}

var promMiddleware = ginprometheus.NewPrometheus("restapi", customMetrics)

func MetricsMiddleware(e *gin.Engine) {

	// Default metrics
	promMiddleware.Use(e)

	// Custom metrics
	e.Use(customMetricsMiddleware(promMiddleware))
}

func customMetricsMiddleware(p *ginprometheus.Prometheus) func(*gin.Context) {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)

		// Register response time
		responseMetrics := utils.FilterArray[*ginprometheus.Metric](
			p.MetricsList, func(m *ginprometheus.Metric) bool { return m.Name == "response_time" })

		if len(responseMetrics) != 1 {
			return
		}
		responseMetric := responseMetrics[0]
		responseMetric.MetricCollector.(*prometheus.HistogramVec).WithLabelValues(
			status, utils.ReplaceParamsValuesFromUrl(c), c.HandlerName(),
		).Observe(elapsed)
	}
}
