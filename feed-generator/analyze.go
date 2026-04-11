package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"bsky-schwartz/db"
	"bsky-schwartz/pkg/schwartz"
)

// AnalyzePosts is the entry point for the CLI command.
func AnalyzePosts(modelFilter string) {
	if err := db.Init(DBPath); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := initLogging(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init logging: %v\n", err)
		os.Exit(1)
	}
	defer closeLogging()

	fmt.Printf("Using prompt version: %s\n", *promptVersion)

	maxAnalysesPerModel := 1

	filteredConfigs := modelConfigs
	if modelFilter != "" {
		filteredConfigs = filterModels(modelConfigs, modelFilter)
		if len(filteredConfigs) == 0 {
			fmt.Printf("No models match filter: %s\n", modelFilter)
			fmt.Printf("Available models: ")
			for i, cfg := range modelConfigs {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(cfg.Name)
			}
			fmt.Println()
			return
		}
		fmt.Printf("Filtered to %d model(s): ", len(filteredConfigs))
		for i, cfg := range filteredConfigs {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(cfg.Name)
		}
		fmt.Println()
	}

	// Get posts needing analysis for the first model in the filtered list
	modelName := ""
	if len(filteredConfigs) > 0 {
		modelName = filteredConfigs[0].ModelID
	}

	var posts []schwartz.Post
	var err error

	if *sampleFlag > 0 {
		fmt.Printf("Getting random sample of %d posts...\n", *sampleFlag)
		posts, err = db.GetRandomPosts(modelName, maxAnalysesPerModel, *sampleFlag)
	} else {
		posts, err = db.GetPostsNeedingAnalysis(modelName, maxAnalysesPerModel)
	}

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

	for i, cfg := range filteredConfigs {
		fmt.Printf("\n[%d/%d] Model: %s (%s)\n", i+1, len(filteredConfigs), cfg.Name, cfg.Provider)
		fmt.Println("----------------------------------------")

		analysisCounts, err := db.GetPostCountForModel(cfg.ModelID)
		if err != nil {
			fmt.Printf("ERROR: could not get analysis counts: %v\n", err)
			continue
		}

		// Post To Analayze
		var postsToAnalyze []schwartz.Post
		var skippedPosts []string
		for _, post := range posts {
			count := analysisCounts[post.AtURI]
			if count >= maxAnalysesPerModel {
				skippedPosts = append(skippedPosts, post.AtURI)
			} else {
				postsToAnalyze = append(postsToAnalyze, post)
			}
		}
		if len(skippedPosts) > 0 {
			fmt.Printf("Skipped %d posts already having %d analyses for this model\n", len(skippedPosts), maxAnalysesPerModel)
		}
		if len(postsToAnalyze) == 0 {
			fmt.Println("No posts to analyze for this model!")
			continue
		}
		fmt.Printf("Analyzing %d posts for this model\n", len(postsToAnalyze))

		// Get the AI Client
		var client AIClient
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

		// Process with Routines
		var wg sync.WaitGroup
		var mu sync.Mutex
		counter := 0
		workers := 5

		in := make(chan schwartz.Post, workers)

		for range workers {
			wg.Add(1)
			go runAnalysisAsync(&wg, &mu, &counter, client, in, cfg.ModelID, cfg.Provider, *promptVersion)
		}

		for _, post := range postsToAnalyze {
			in <- post
		}
		close(in)

		wg.Wait()

		fmt.Println("\n========================================")
		fmt.Println("All models processed.")
		fmt.Println("========================================")
	}
}

// runAnalysisAsync processes posts concurrently with the AI model.
func runAnalysisAsync(wg *sync.WaitGroup, mu *sync.Mutex, counter *int, client AIClient, in <-chan schwartz.Post, model string, provider string, promptVersion string) {
	defer wg.Done()

	for post := range in {
		mu.Lock()
		*counter++
		currentCounter := *counter
		mu.Unlock()

		postStart := time.Now()
		fmt.Printf("  [%d] Analyzing: %s\n", currentCounter, truncate(post.Text, 50))

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

		analysis, err := CalculateRating(ctx, client, model, &post, promptVersion)
		if err != nil {
			fmt.Printf("  ERROR analyzing post: %v\n", err)
			cancel()
			continue
		}

		fmt.Printf("  OK - Tokens: %d - Time: %v\n",
			analysis.Stats.TotalTokens,
			time.Since(postStart).Round(time.Millisecond))

		logAnalysis(post.AtURI, model, analysis.Rating, analysis.Reasoning, analysis.Stats)

		if err := db.SaveAnalysis(post.AtURI, model, provider, *analysis); err != nil {
			fmt.Printf("  ERROR saving analysis: %v\n", err)
		}

		cancel()
		time.Sleep(3 * time.Second)
	}
}

// truncate limits text length for display (adds "..." if truncated).
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// filterModels filters model configs by name (partial match).
func filterModels(configs []ModelConfig, name string) []ModelConfig {
	name = strings.ToLower(name)
	var filtered []ModelConfig
	for _, cfg := range configs {
		if strings.Contains(strings.ToLower(cfg.Name), name) || strings.Contains(strings.ToLower(cfg.ModelID), name) {
			filtered = append(filtered, cfg)
		}
	}
	if len(filtered) > 1 {
		fmt.Printf("Multiple models match '%s', using first: %s\n", name, filtered[0].Name)
		return filtered[:1]
	}
	return filtered
}

// GetOpenRouterClient creates an OpenRouter AI client.
func GetOpenRouterClient() (AIClient, error) {
	key := os.Getenv("OPEN_ROUTER_KEY")
	if key == "" {
		return nil, fmt.Errorf("OPEN_ROUTER_KEY not set")
	}
	return NewOpenRouterClient(key), nil
}

// GetSiliconFlowClient creates a SiliconFlow AI client.
func GetSiliconFlowClient() (AIClient, error) {
	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SILICONFLOW_API_KEY not set")
	}
	return NewOpenAIClient(apiKey, "https://api.siliconflow.com/v1"), nil
}
