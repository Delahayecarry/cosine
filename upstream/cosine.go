package upstream

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"cosine/config"
	"cosine/models"
)

type CosineClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCosineClient() *CosineClient {
	return &CosineClient{
		baseURL:    config.GlobalConfig.Upstream.BaseURL,
		httpClient: &http.Client{},
	}
}

// SendChatRequest 发送聊天请求到 Cosine API，返回响应体供流式处理
func (c *CosineClient) SendChatRequest(req *models.CosineChatRequest, auth string) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Cookie", fmt.Sprintf("auth=%s", auth))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// ParseCosineStream 解析 Cosine 的自定义流式格式
// 返回一个 channel 用于接收解析后的内容
func ParseCosineStream(reader io.Reader) (<-chan StreamEvent, <-chan error) {
	eventCh := make(chan StreamEvent, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 {
				continue
			}

			event := parseLine(line)
			if event != nil {
				eventCh <- *event
			}
		}

		if err := scanner.Err(); err != nil {
			errCh <- err
		}
	}()

	return eventCh, errCh
}

type StreamEvent struct {
	Type    string // "content", "finish", "other"
	Content string
	Finish  *models.CosineFinishEvent
}

func parseLine(line string) *StreamEvent {
	// 找到第一个冒号的位置
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 || colonIdx == 0 {
		return nil
	}

	prefix := strings.TrimSpace(line[:colonIdx])
	data := strings.TrimSpace(line[colonIdx+1:])

	switch prefix {
	case "0":
		// 文本内容，格式为 JSON 字符串
		var content string
		if err := json.Unmarshal([]byte(data), &content); err != nil {
			// 如果解析失败，直接使用原始数据
			content = data
		}
		return &StreamEvent{Type: "content", Content: content}

	case "e":
		// 结束事件
		var finish models.CosineFinishEvent
		if err := json.Unmarshal([]byte(data), &finish); err != nil {
			return &StreamEvent{Type: "finish", Finish: &models.CosineFinishEvent{FinishReason: "stop"}}
		}
		return &StreamEvent{Type: "finish", Finish: &finish}

	case "d":
		// 最终结束标记，也作为 finish 处理
		var finish models.CosineFinishEvent
		if err := json.Unmarshal([]byte(data), &finish); err != nil {
			return nil
		}
		return &StreamEvent{Type: "finish", Finish: &finish}

	default:
		// 其他类型（2, f 等）忽略
		return nil
	}
}

// CollectFullResponse 收集完整的非流式响应
func CollectFullResponse(reader io.Reader) (string, *models.CosineFinishEvent, error) {
	eventCh, errCh := ParseCosineStream(reader)

	var content strings.Builder
	var finishEvent *models.CosineFinishEvent

	for event := range eventCh {
		switch event.Type {
		case "content":
			content.WriteString(event.Content)
		case "finish":
			finishEvent = event.Finish
		}
	}

	if err := <-errCh; err != nil {
		return "", nil, err
	}

	return content.String(), finishEvent, nil
}
