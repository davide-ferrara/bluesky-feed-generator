package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"bsky-schwartz/pkg/schwartz"
)

func cleanMarkdown(s string) string {
	re := regexp.MustCompile("(?s)^```(?:json)?\\s*\\n?")
	s = re.ReplaceAllString(strings.TrimSpace(s), "")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func BuildPromptContent(post *schwartz.Post) string {
	content := map[string]interface{}{
		"text": post.Text,
	}

	if len(post.Langs) > 0 {
		content["language"] = strings.Join(post.Langs, ", ")
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

func CalculateRating(ctx context.Context, client AIClient, model string, post *schwartz.Post) (*schwartz.ValueAnalysis, error) {
	start := time.Now()

	taskPrompt, err := os.ReadFile("./prompts/PROMPT_V3.md")
	if err != nil {
		return nil, fmt.Errorf("reading prompt: %w", err)
	}

	promptContent := BuildPromptContent(post)
	prompt := fmt.Sprintf("%s\n\n%s", string(taskPrompt), promptContent)

	resp, err := client.CreateChatCompletion(ctx, model, prompt)
	if err != nil {
		return nil, fmt.Errorf("ai client error: %w", err)
	}

	elapsed := time.Since(start)

	jsonStr := cleanMarkdown(resp.Content)

	var result struct {
		Rating    schwartz.SchwartzValues `json:"Rating"`
		Reasoning string                  `json:"Reasoning"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w, json: %s", err, jsonStr)
	}

	stats := schwartz.AIStats{
		Model:            resp.Model,
		ResponseTimeMs:   elapsed.Milliseconds(),
		PromptTokens:     resp.PromptTokens,
		CompletionTokens: resp.CompletionTokens,
		TotalTokens:      resp.TotalTokens,
		CostUsd:          resp.CostUsd,
		Provider:         resp.Provider,
	}

	return &schwartz.ValueAnalysis{
		Rating:    result.Rating,
		Reasoning: result.Reasoning,
		Stats:     stats,
	}, nil
}
