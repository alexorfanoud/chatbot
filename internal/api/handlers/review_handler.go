package handlers

import (
	"chat/internal/data/dao"
	"chat/internal/model"
	"chat/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReviewRequest struct {
	UserId    int
	ProductId int
}

func HandleReview(c *gin.Context) {
	ctx := c.Request.Context()

	var request ReviewRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		utils.Log(ctx, fmt.Sprintf("Error decoding message: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message format"})
		return
	}
	utils.Log(ctx, fmt.Sprintf("Received new review request: %q", request))

	user, err := dao.GetUserById(ctx, request.UserId)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Error fetching user: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to fetch user")})
		return
	}
	product, err := dao.GetProductByID(ctx, request.ProductId)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Error fetching product: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to fetch product")})
		return
	}

	resp, err := conversationManager.HandleRequest(ctx,
		"", user.Id, true,
		&model.WorkflowExecutionContext{
			Workflow:      model.REVIEW,
			ContextWindow: 0,
			PromptVariables: map[string]string{
				"product":    fmt.Sprintf("{name: %s, description: %s}", product.Name, product.Description),
				"product_id": product.ID,
				"user":       user.Name,
				"user_id":    fmt.Sprintf("%d", user.Id)},
		})

	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Error triggering review process: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to start review process")})
		return
	}

	// Send response back to the user
	err = notifier.Notify(ctx, user, resp)
	if err != nil {
		utils.Log(ctx, fmt.Sprintf("Error sending response notification: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to send response notification")})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
