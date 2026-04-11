package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var (
	modelFlag     = flag.String("model", "", "Run analysis for specific model")
	limitAnalyze  = flag.Int("l", 0, "Limit posts")
	sampleFlag    = flag.Int("sample", 0, "Random sample")
	promptVersion = flag.String("prompt", "v4", "Prompt version")
	dbPath        = flag.String("db", "", "DB path")
	profilesPath  = flag.String("profiles", "", "Profiles path")
	collectTotal  = flag.Int("total", 0, "Total posts to collect")
)

func GetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "Missing .env file\n")
		os.Exit(1)
	}
	return v
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: .env not found: %v\n", err)
	}

	flag.Parse()

	if *dbPath != "" {
		DBPath = *dbPath
	}
	if *profilesPath != "" {
		ProfilesFile = *profilesPath
	}

	switch flag.Arg(0) {

	case "analyze":
		AnalyzePosts(*modelFlag)

	case "from-file":
		if flag.NArg() < 2 {
			fmt.Fprintf(os.Stderr, "Usage: feedgen from-file <urls.json>\n")
			os.Exit(1)
		}
		FetchFromFile(flag.Arg(1))

	case "collect-profiles":
		CollectFromProfiles()

	default:
		fmt.Printf("Unknown command: %s\n", flag.Arg(0))
		os.Exit(1)
	}
}
