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

// ChatMessage is one message in an OpenAI chat completions conversation. Tool fields
// are only set on assistant messages that request tool calls, and on the "tool" role
// messages that answer them.
type ChatMessage struct {
	Role       string     `json:"role"` // "system" | "user" | "assistant" | "tool"
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // a "tool" message answers this call
}

// ToolCall is the model's request to run one function.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON-encoded string
	} `json:"function"`
}

// ToolDefinition advertises one callable function to the model.
type ToolDefinition struct {
	Type     string       `json:"type"` // "function"
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema
}

const openAIChatURL = "https://api.openai.com/v1/chat/completions"
const openAIModel = "gpt-5-mini"

// tool-call loops multiply round trips — 120s cap so a hung call can't pin a goroutine.
var openAIClient = &http.Client{Timeout: 120 * time.Second}

type chatCompletionRequest struct {
	Model    string           `json:"model"`
	Messages []ChatMessage    `json:"messages"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
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
* the assistant's reply message — either final text (Content) or one or more ToolCalls
* the caller must execute and feed back as "tool" messages. OPENAI_API_KEY must be set
* (docker-compose passes it through) — a missing key is a clear, actionable error.
**************************************************************************************/
func ChatCompletion(messages []ChatMessage, tools []ToolDefinition) (ChatMessage, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return ChatMessage{}, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	body, err := json.Marshal(chatCompletionRequest{Model: openAIModel, Messages: messages, Tools: tools})
	if err != nil {
		return ChatMessage{}, fmt.Errorf("could not encode chat request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, openAIChatURL, bytes.NewReader(body))
	if err != nil {
		return ChatMessage{}, fmt.Errorf("could not build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	res, err := openAIClient.Do(req)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("could not reach OpenAI: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("could not read OpenAI response: %w", err)
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(resBody, &parsed); err != nil {
		return ChatMessage{}, fmt.Errorf("could not parse OpenAI response (status %d): %w", res.StatusCode, err)
	}
	if parsed.Error != nil {
		return ChatMessage{}, fmt.Errorf("OpenAI error: %s", parsed.Error.Message)
	}
	if res.StatusCode != http.StatusOK {
		return ChatMessage{}, fmt.Errorf("OpenAI returned status %d", res.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return ChatMessage{}, fmt.Errorf("OpenAI returned no choices")
	}

	return parsed.Choices[0].Message, nil
}
