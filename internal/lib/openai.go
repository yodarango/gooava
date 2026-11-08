package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ChatMessage is one message in an OpenAI chat completions conversation.
type ChatMessage struct {
	Role    string `json:"role"` // "system" | "user" | "assistant"
	Content string `json:"content"`
}

const openAIChatURL = "https://api.openai.com/v1/chat/completions"
const openAIModel = "gpt-5-mini"

// 60s cap — a hung OpenAI call must not pin a goroutine forever.
var openAIClient = &http.Client{Timeout: 60 * time.Second}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	// gpt-5-mini rejects custom sampling params — never add temperature/top_p here.
}

type chatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

/**************************************************************************************
* ChatCompletion sends the conversation to OpenAI's chat completions API and returns
* the assistant's reply text. OPENAI_API_KEY must be set (docker-compose passes it
* through) — a missing key is a clear, actionable error.
**************************************************************************************/
func ChatCompletion(messages []ChatMessage) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	body, err := json.Marshal(chatCompletionRequest{Model: openAIModel, Messages: messages})
	if err != nil {
		return "", fmt.Errorf("could not encode chat request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, openAIChatURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("could not build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	res, err := openAIClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach OpenAI: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("could not read OpenAI response: %w", err)
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(resBody, &parsed); err != nil {
		return "", fmt.Errorf("could not parse OpenAI response (status %d): %w", res.StatusCode, err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("OpenAI error: %s", parsed.Error.Message)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI returned status %d", res.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("OpenAI returned no choices")
	}

	return parsed.Choices[0].Message.Content, nil
}
