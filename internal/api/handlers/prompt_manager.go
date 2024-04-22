package handlers

import (
	"chat/internal/data/dao"
	"chat/internal/model"
	"chat/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandlePromptUpdate(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := utils.GetTraceIDFromContext(ctx)

	var request model.Prompt
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		log.Printf("[traceID=%s] Error decoding message: %s\n", traceID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	err := dao.UpdatePrompt(ctx, request.ToPromptDTO())
	if err != nil {
		log.Printf("[traceID=%s] Error updating prompt: %s\n", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to update prompt")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
