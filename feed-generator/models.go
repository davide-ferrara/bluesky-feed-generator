package main

import (
	"context"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/openai/openai-go/v3"
	openrouter "github.com/revrost/go-openrouter"
)

type ModelConfig struct {
	Name     string
	Provider string // "openrouter" or "siliconflow"
	ModelID  string // model ID for the provider
}

var modelConfigs = []ModelConfig{
	{Name: "gpt-4o-mini", Provider: "openrouter", ModelID: "openai/gpt-4o-mini"},
	{Name: "ministral-8b", Provider: "openrouter", ModelID: "mistralai/ministral-8b-2512"},
	{Name: "ministral-3-14b", Provider: "openrouter", ModelID: "mistralai/ministral-14b-2512"},
	{Name: "qwen3-vl-8b", Provider: "siliconflow", ModelID: "Qwen/Qwen3-VL-8B-Instruct"},
	{Name: "gpt-4.1", Provider: "openrouter", ModelID: "openai/gpt-4.1"},
	{Name: "claude-sonnet-4.6", Provider: "openrouter", ModelID: "anthropic/claude-sonnet-4.6"},
}

type Client struct {
	client *xrpc.Client
}

type AIClient interface {
	CreateChatCompletion(ctx context.Context, model string, prompt string) (*AIResponse, error)
}

type AIResponse struct {
	Content          string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	CostUsd          float64
	Provider         string
}

type OpenRouterClient struct {
	client *openrouter.Client
}

type OpenAIClient struct {
	client *openai.Client
}
