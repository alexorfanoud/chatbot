package routes

import (
	"chat/internal/api/handlers"
	"chat/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {

	router := gin.Default()
	gin.Default()

	// Apply all middlewares
	for _, middlewareFn := range middleware.GetAllMiddlewares() {
		middlewareFn(router)
	}

	apiGroup := router.Group("/api/v1")
	{
		apiGroup.POST("/review", handlers.HandleReview)
		apiGroup.PUT("/prompts", handlers.HandlePromptUpdate)
	}

	return router
}
