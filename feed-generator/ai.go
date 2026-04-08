package main

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	openrouter "github.com/revrost/go-openrouter"
)

func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		client: openrouter.NewClient(apiKey),
	}
}

func (c *OpenRouterClient) CreateChatCompletion(ctx context.Context, model string, prompt string) (*AIResponse, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openrouter.ChatCompletionRequest{
			Model:       model,
			Temperature: 0.3,
			Messages: []openrouter.ChatCompletionMessage{
				openrouter.UserMessage(prompt),
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("openrouter error: %w", err)
	}

	return &AIResponse{
		Content:          resp.Choices[0].Message.Content.Text,
		Model:            resp.Model,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
		CostUsd:          resp.Usage.Cost,
		Provider:         resp.Provider,
	}, nil
}

func NewOpenAIClient(apiKey string, baseURL string) *OpenAIClient {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := openai.NewClient(opts...)
	return &OpenAIClient{client: &client}
}

func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, model string, prompt string) (*AIResponse, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("openai error: %w", err)
	}

	return &AIResponse{
		Content:          resp.Choices[0].Message.Content,
		Model:            string(resp.Model),
		PromptTokens:     int(resp.Usage.PromptTokens),
		CompletionTokens: int(resp.Usage.CompletionTokens),
		TotalTokens:      int(resp.Usage.TotalTokens),
		CostUsd:          0,
		Provider:         "openai",
	}, nil
}
