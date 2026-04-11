package main

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	openrouter "github.com/revrost/go-openrouter"
)

type ChatContent struct {
	Text   string
	Images []string
}

func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{client: openrouter.NewClient(apiKey)}
}

func (c *OpenRouterClient) CreateChatCompletion(ctx context.Context, model string, content ChatContent) (*AIResponse, error) {
	temperature := float32(0.7)

	var msg openrouter.ChatCompletionMessage
	if len(content.Images) > 0 {
		msg = openrouter.UserMessageWithImage(content.Text, content.Images[0])
	} else {
		msg = openrouter.UserMessage(content.Text)
	}

	resp, err := c.client.CreateChatCompletion(ctx, openrouter.ChatCompletionRequest{
		Model:       model,
		Temperature: temperature,
		Messages:    []openrouter.ChatCompletionMessage{msg},
	})
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
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	return &OpenAIClient{client: &client}
}

func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, model string, content ChatContent) (*AIResponse, error) {
	temperature := float64(0.7)

	var msg openai.ChatCompletionMessageParamUnion
	if len(content.Images) > 0 {
		msg = openai.UserMessage(content.Text)
	} else {
		msg = openai.UserMessage(content.Text)
	}

	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       model,
		Messages:    []openai.ChatCompletionMessageParamUnion{msg},
		Temperature: openai.Float(temperature),
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
