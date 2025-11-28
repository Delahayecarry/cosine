package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"cosine/database"
	"cosine/models"
	"cosine/upstream"

	"github.com/gin-gonic/gin"
)

const maxRetries = 3

func ChatCompletionsHandler(c *gin.Context) {
	var req models.OpenAIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sendError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	// 转换请求格式
	cosineReq := convertToCosineRequest(&req)

	// 带重试的请求
	var resp *http.Response
	var account *models.Account
	var err error

	for i := 0; i < maxRetries; i++ {
		account, err = database.GetNextAccount()
		if err != nil {
			sendError(c, http.StatusServiceUnavailable, "service_unavailable", "no available accounts")
			return
		}

		cosineReq.TeamID = account.TeamID
		client := upstream.NewCosineClient()
		resp, err = client.SendChatRequest(cosineReq, account.Auth)

		if err != nil {
			log.Printf("Request failed for account %d: %v", account.ID, err)
			continue
		}

		// 检查响应状态码
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			log.Printf("Account %d returned %d, deactivating", account.ID, resp.StatusCode)
			database.DeactivateAccount(account.ID)
			resp.Body.Close()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Upstream returned status %d", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		// 请求成功
		break
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		sendError(c, http.StatusBadGateway, "upstream_error", "failed to get response from upstream after retries")
		return
	}
	defer resp.Body.Close()

	if req.Stream {
		handleStreamResponse(c, resp, req.Model)
	} else {
		handleNonStreamResponse(c, resp, req.Model)
	}
}

func convertToCosineRequest(req *models.OpenAIChatRequest) *models.CosineChatRequest {
	cosineMessages := make([]models.CosineMessage, len(req.Messages))
	for i, msg := range req.Messages {
		cosineMessages[i] = models.CosineMessage{
			Content:   msg.Content,
			Role:      msg.Role,
			ID:        generateID(12),
			CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		}
	}

	return &models.CosineChatRequest{
		ID:         "",
		Messages:   cosineMessages,
		Model:      req.Model,
		Visibility: "team",
	}
}

func handleStreamResponse(c *gin.Context, resp *http.Response, model string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	chatID := "chatcmpl-" + generateID(24)
	created := time.Now().Unix()

	eventCh, errCh := upstream.ParseCosineStream(resp.Body)

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel 关闭，发送 [DONE]
				fmt.Fprintf(w, "data: [DONE]\n\n")
				return false
			}

			switch event.Type {
			case "content":
				chunk := models.OpenAIChatResponse{
					ID:      chatID,
					Object:  "chat.completion.chunk",
					Created: created,
					Model:   model,
					Choices: []models.OpenAIChoice{
						{
							Index: 0,
							Delta: &models.OpenAIDelta{
								Content: event.Content,
							},
							FinishReason: nil,
						},
					},
				}
				data, _ := json.Marshal(chunk)
				fmt.Fprintf(w, "data: %s\n\n", data)

			case "finish":
				finishReason := "stop"
				if event.Finish != nil && event.Finish.FinishReason != "" {
					finishReason = event.Finish.FinishReason
				}
				chunk := models.OpenAIChatResponse{
					ID:      chatID,
					Object:  "chat.completion.chunk",
					Created: created,
					Model:   model,
					Choices: []models.OpenAIChoice{
						{
							Index:        0,
							Delta:        &models.OpenAIDelta{},
							FinishReason: &finishReason,
						},
					},
				}
				data, _ := json.Marshal(chunk)
				fmt.Fprintf(w, "data: %s\n\n", data)
			}
			return true

		case err := <-errCh:
			if err != nil {
				log.Printf("Stream error: %v", err)
			}
			fmt.Fprintf(w, "data: [DONE]\n\n")
			return false
		}
	})
}

func handleNonStreamResponse(c *gin.Context, resp *http.Response, model string) {
	content, finishEvent, err := upstream.CollectFullResponse(resp.Body)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	finishReason := "stop"
	if finishEvent != nil && finishEvent.FinishReason != "" {
		finishReason = finishEvent.FinishReason
	}

	response := models.OpenAIChatResponse{
		ID:      "chatcmpl-" + generateID(24),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []models.OpenAIChoice{
			{
				Index: 0,
				Message: &models.OpenAIMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: &finishReason,
			},
		},
		Usage: &models.OpenAIUsage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	c.JSON(http.StatusOK, response)
}

func sendError(c *gin.Context, status int, errType, message string) {
	c.JSON(status, models.ErrorResponse{
		Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		}{
			Message: message,
			Type:    errType,
			Code:    errType,
		},
	})
}

func generateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
