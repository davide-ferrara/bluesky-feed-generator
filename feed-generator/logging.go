package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bsky-schwartz/pkg/schwartz"
)

var logFile *os.File

func initLogging() error {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
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
