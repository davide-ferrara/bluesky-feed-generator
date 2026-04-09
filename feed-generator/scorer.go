package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"bsky-schwartz/pkg/schwartz"
)

var logFile *os.File

func initLogging() error {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("could not create logs directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logDir, fmt.Sprintf("analysis_%s.log", timestamp))

	f, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("could not create log file: %w", err)
	}
	logFile = f

	fmt.Printf("Logging to: %s\n", logPath)
	return nil
}

func closeLogging() {
	if logFile != nil {
		logFile.Close()
	}
}

func logAnalysis(postAtURI, model string, rating schwartz.SchwartzValues, reasoning string, stats schwartz.AIStats) {
	if logFile == nil {
		return
	}

	logEntry := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"post_at_uri": postAtURI,
		"model":       model,
		"rating":      rating,
		"reasoning":   reasoning,
		"stats": map[string]interface{}{
			"model":             stats.Model,
			"response_time_ms":  stats.ResponseTimeMs,
			"prompt_tokens":     stats.PromptTokens,
			"completion_tokens": stats.CompletionTokens,
			"total_tokens":      stats.TotalTokens,
			"cost_usd":          stats.CostUsd,
			"provider":          stats.Provider,
		},
	}

	jsonBytes, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Printf("ERROR: could not marshal log entry: %v\n", err)
		return
	}

	logFile.Write(jsonBytes)
	logFile.WriteString("\n")
}

func cleanMarkdown(s string) string {
	re := regexp.MustCompile("(?s)^```(?:json)?\\s*\\n?")
	s = re.ReplaceAllString(strings.TrimSpace(s), "")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func BuildPromptContent(post *schwartz.Post, analyzeImages bool) string {
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

	if analyzeImages && len(post.Images) > 0 {
		var images []map[string]string
		for _, img := range post.Images {
			images = append(images, map[string]string{
				"url": img.Image,
				"alt": img.Alt,
			})
		}
		content["images"] = images
	}

	jsonBytes, _ := json.Marshal(content)
	return fmt.Sprintf("<post>\n%s\n</post>", jsonBytes)
}

func CalculateRating(ctx context.Context, client AIClient, model string, post *schwartz.Post, analyzeImages bool) (*schwartz.ValueAnalysis, error) {
	start := time.Now()

	taskPrompt, err := os.ReadFile("./prompts/PROMPT_V3.md")
	if err != nil {
		return nil, fmt.Errorf("reading prompt: %w", err)
	}

	promptContent := BuildPromptContent(post, analyzeImages)
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
