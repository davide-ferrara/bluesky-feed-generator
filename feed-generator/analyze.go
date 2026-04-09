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

func AnalyzePosts(modelFilter string) {
	if err := db.Init("../data.db"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := initLogging(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init logging: %v\n", err)
		os.Exit(1)
	}
	defer closeLogging()

	maxAnalysesPerModel := 5

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
		fmt.Printf("Filtering models by: %s\n", modelFilter)
	}

	// Get posts needing analysis for the first model in the filtered list
	modelName := ""
	if len(filteredConfigs) > 0 {
		modelName = filteredConfigs[0].ModelID
	}

	posts, err := db.GetPostsNeedingAnalysis(modelName, maxAnalysesPerModel)
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

	fmt.Printf("Image analysis: %v\n", *analyzeImages)

	for i, cfg := range filteredConfigs {
		fmt.Printf("\n[%d/%d] Model: %s (%s)\n", i+1, len(filteredConfigs), cfg.Name, cfg.Provider)
		fmt.Println("----------------------------------------")

		analysisCounts, err := db.GetPostCountForModel(cfg.ModelID)
		if err != nil {
			fmt.Printf("ERROR: could not get analysis counts: %v\n", err)
			continue
		}

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

		if err := runAnalysisForModel(client, postsToAnalyze, cfg.ModelID, cfg.Provider, *analyzeImages); err != nil {
			fmt.Printf("ERROR: Analysis failed: %v\n", err)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("All models processed.")
	fmt.Println("========================================")
}

func runAnalysisForModel(client AIClient, posts []schwartz.Post, model string, provider string, analyzeImages bool) error {
	startTime := time.Now()

	for i := range posts {
		postStart := time.Now()
		fmt.Printf("  [Post %d/%d] Analyzing: %s\n", i+1, len(posts), truncate(posts[i].Text, 50))

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

		analysis, err := CalculateRating(ctx, client, model, &posts[i], analyzeImages)
		if err != nil {
			posts[i].ValueAnalysis.Error = err.Error()
			fmt.Printf("  [Post %d/%d] ERROR: %v\n", i+1, len(posts), err)
		} else {
			posts[i].ValueAnalysis = *analysis
			fmt.Printf("  [Post %d/%d] OK - Tokens: %d - Time: %v\n",
				i+1, len(posts),
				analysis.Stats.TotalTokens,
				time.Since(postStart).Round(time.Millisecond))

			logAnalysis(posts[i].AtURI, model, analysis.Rating, analysis.Reasoning, analysis.Stats)

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
