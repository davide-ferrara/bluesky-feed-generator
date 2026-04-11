package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	bskySchwartz "bsky-schwartz/pkg/schwartz"
)

var promptCache sync.Map

// CalculateRating calculates the Schwartz values (0-6) for a post using the AI client.
func CalculateRating(ctx context.Context, client AIClient, model string, post *bskySchwartz.Post, promptVersion string) (*bskySchwartz.ValueAnalysis, error) {
	start := time.Now()

	taskPrompt, err := getPrompt(promptVersion)
	if err != nil {
		return nil, fmt.Errorf("reading prompt: %w", err)
	}

	promptContent := BuildPromptContent(post)
	prompt := fmt.Sprintf("%s\n\n%s", taskPrompt, promptContent)

	resp, err := client.CreateChatCompletion(ctx, model, ChatContent{Text: prompt, Images: extractImageURLs(post)})
	if err != nil {
		return nil, fmt.Errorf("ai client error: %w", err)
	}

	elapsed := time.Since(start)

	jsonStr := cleanMarkdown(resp.Content)

	var result struct {
		Rating    bskySchwartz.SchwartzValues `json:"Rating"`
		Reasoning string                      `json:"Reasoning"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w, json: %s", err, jsonStr)
	}

	stats := bskySchwartz.AIStats{
		Model:            resp.Model,
		ResponseTimeMs:   elapsed.Milliseconds(),
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
		TotalTokens:      resp.TotalTokens,
		CostUsd:          resp.CostUsd,
		Provider:         resp.Provider,
	}

	return &bskySchwartz.ValueAnalysis{
		Rating:    result.Rating,
		Reasoning: result.Reasoning,
		Stats:     stats,
	}, nil
}

// BuildPromptContent returns a string containing the full prompt.
// It includes: Text of the Post, Author, External Links.
func BuildPromptContent(post *bskySchwartz.Post) string {
	content := map[string]any{
		"text":   post.Text,
		"author": post.AuthorName,
	}

	if len(post.Links) > 0 {
		var links []map[string]string
		for _, l := range post.Links {
			links = append(links, map[string]string{
				"title":       l.Title,
				"description": l.Description,
			})
		}
		content["external_links"] = links
	}

	jsonBytes, _ := json.Marshal(content)
	return fmt.Sprintf("<post>\n%s\n</post>", jsonBytes)
}

func extractImageURLs(post *bskySchwartz.Post) []string {
	var urls []string
	for _, img := range post.Images {
		if img.Image != "" {
			urls = append(urls, img.Image)
		}
	}
	return urls
}

// getPrompt gets the prompt from the cache or
// it gets loaded from the according file and
// saved to cache.
func getPrompt(version string) (string, error) {
	var promptFile string

	switch version {
	case "v1":
		promptFile = "prompts/PROMPT_V1.md"
	case "v2":
		promptFile = "prompts/PROMPT_V2.md"
	case "v3":
		promptFile = "prompts/PROMPT_V3.md"
	default:
		promptFile = "prompts/PROMPT_V4.md"
	}

	if val, ok := promptCache.Load(promptFile); ok {
		return val.(string), nil
	}

	data, err := os.ReadFile(promptFile)
	if err != nil {
		return "", err
	}

	prompt := string(data)
	promptCache.Store(promptFile, prompt)
	return prompt, nil
}

// cleanMarkdown cleans the response from MarkDown since
// smaller models could make mistakes like ministral-8b,
// not necessary with models upper to 8b in my experience
func cleanMarkdown(s string) string {
	re := regexp.MustCompile("(?s)^```(?:json)?\\s*\\n?")
	s = re.ReplaceAllString(strings.TrimSpace(s), "")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
