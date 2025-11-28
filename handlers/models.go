package handlers

import (
	"net/http"

	"cosine/models"

	"github.com/gin-gonic/gin"
)

var supportedModels = []models.OpenAIModel{
	{ID: "gpt-5", Object: "model", Created: 1700000000, OwnedBy: "cosine"},
	{ID: "gpt4.1", Object: "model", Created: 1700000000, OwnedBy: "cosine"},
	{ID: "claude-3-7-sonnet", Object: "model", Created: 1700000000, OwnedBy: "cosine"},
	{ID: "gemini-2.0-flash", Object: "model", Created: 1700000000, OwnedBy: "cosine"},
}

func ModelsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, models.OpenAIModelsResponse{
		Object: "list",
		Data:   supportedModels,
	})
}
