package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var (
	collectLimit  = flag.Int("n", 40, "Total posts to collect for feed (balanced across 4 clusters)")
	limitFlag     = flag.Int("limit", 0, "Number of posts to analyze per model")
	modelFlag     = flag.String("model", "", "Run analysis for specific model (partial name match)")
	limitAnalyze  = flag.Int("l", 0, "Limit posts to analyze")
	langFlag      = flag.String("lang", "it", "Language filter for posts (e.g., it, en)")
	analyzeImages = flag.Bool("images", true, "Include images in AI analysis")
)

func GetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing env: " + key)
	}
	return v
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: .env not found: %v\n", err)
	}

	flag.Parse()

	switch flag.Arg(0) {

	case "":
		postsPerCluster := *collectLimit / 4
		GenerateFeed(postsPerCluster, *langFlag)
		os.Exit(0)

	case "analyze":
		AnalyzePosts(*modelFlag)
		os.Exit(0)

	case "from-file":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: feedgen from-file <urls.json>\n")
			os.Exit(1)
		}
		FetchFromFile(flag.Arg(1))
		os.Exit(0)

	default:
		fmt.Printf("Unknown command: %s\n", flag.Arg(0))
		os.Exit(0)
	}
}
