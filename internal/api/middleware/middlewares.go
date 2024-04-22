package middleware

import "github.com/gin-gonic/gin"

type Middleware func(*gin.Engine)

func GetAllMiddlewares() []Middleware {
	return []Middleware{
		OtelMiddleware, MetricsMiddleware,
	}
}
