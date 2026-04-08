package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"bsky-schwartz/db"
	"bsky-schwartz/pkg/schwartz"
)

func AnalyzePosts() {
	if err := db.Init("../data.db"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	posts, err := db.GetUnanalyzedPosts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not get posts: %v\n", err)
		os.Exit(1)
	}

	if len(posts) == 0 {
		fmt.Println("No posts to analyze!")
		return
	}

	fmt.Printf("Found %d posts to analyze\n", len(posts))

	if *limitAnalyze > 0 && len(posts) > *limitAnalyze {
		posts = posts[:*limitAnalyze]
		fmt.Printf("Limited to %d posts\n", len(posts))
	}

	for i, cfg := range modelConfigs {
		fmt.Printf("\n[%d/%d] Model: %s (%s)\n", i+1, len(modelConfigs), cfg.Name, cfg.Provider)
		fmt.Println("----------------------------------------")

		var client AIClient
		var err error

		switch cfg.Provider {
		case "openrouter":
			client, err = GetOpenRouterClient()
		case "siliconflow":
			client, err = GetSiliconFlowClient()
		default:
			fmt.Printf("ERROR: Unknown provider: %s\n", cfg.Provider)
			continue
		}

		if err != nil {
			fmt.Printf("ERROR: Failed to init client: %v\n", err)
			continue
		}

		if err := runAnalysisForModel(client, posts, cfg.ModelID, cfg.Provider); err != nil {
			fmt.Printf("ERROR: Analysis failed: %v\n", err)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("All models processed.")
	fmt.Println("========================================")
}

func runAnalysisForModel(client AIClient, posts []schwartz.Post, model string, provider string) error {
	startTime := time.Now()

	for i := range posts {
		postStart := time.Now()
		fmt.Printf("  [Post %d/%d] Analyzing: %s\n", i+1, len(posts), truncate(posts[i].Text, 50))

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

		analysis, err := CalculateRating(ctx, client, model, &posts[i])
		if err != nil {
			posts[i].ValueAnalysis.Error = err.Error()
			fmt.Printf("  [Post %d/%d] ERROR: %v\n", i+1, len(posts), err)
		} else {
			posts[i].ValueAnalysis = *analysis
			fmt.Printf("  [Post %d/%d] OK - Tokens: %d - Time: %v\n",
				i+1, len(posts),
				analysis.Stats.TotalTokens,
				time.Since(postStart).Round(time.Millisecond))

			if err := db.SaveAnalysis(posts[i].AtURI, model, provider, *analysis); err != nil {
				fmt.Printf("  ERROR saving analysis: %v\n", err)
			}
		}

		cancel()
		time.Sleep(3 * time.Second)
	}

	fmt.Printf("Completed in %v\n", time.Since(startTime).Round(time.Millisecond))
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func filterModels(configs []ModelConfig, name string) []ModelConfig {
	name = strings.ToLower(name)
	var filtered []ModelConfig
	for _, cfg := range configs {
		if strings.Contains(strings.ToLower(cfg.Name), name) {
			filtered = append(filtered, cfg)
		}
	}
	return filtered
}

func GetOpenRouterClient() (AIClient, error) {
	key := os.Getenv("OPEN_ROUTER_KEY")
	if key == "" {
		return nil, fmt.Errorf("OPEN_ROUTER_KEY not set")
	}
	return NewOpenRouterClient(key), nil
}

func GetSiliconFlowClient() (AIClient, error) {
	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SILICONFLOW_API_KEY not set")
	}
	return NewOpenAIClient(apiKey, "https://api.siliconflow.com/v1"), nil
}
