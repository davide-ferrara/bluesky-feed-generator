package main

import (
	"context"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/openai/openai-go/v3"
	openrouter "github.com/revrost/go-openrouter"
)

var modelConfigs = []ModelConfig{
	{Name: "gpt-4o-mini", Provider: "openrouter", ModelID: "openai/gpt-4o-mini"},
	{Name: "ministral-3-14b", Provider: "openrouter", ModelID: "mistralai/ministral-14b-2512"},
	{Name: "qwen3-14b", Provider: "siliconflow", ModelID: "Qwen/Qwen3-14B"},
}

type ModelConfig struct {
	Name     string
	Provider string // "openrouter" or "siliconflow"
	ModelID  string // model ID for the provider
}

type Client struct {
	client *xrpc.Client
}

type AIClient interface {
	CreateChatCompletion(ctx context.Context, model string, content ChatContent) (*AIResponse, error)
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
