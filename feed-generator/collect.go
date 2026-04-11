package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"bsky-schwartz/db"
)

type URLsFile struct {
	URLs []string `json:"urls"`
}

func FetchFromFile(filepath string) {
	if err := db.Init(DBPath); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	bskyClient, err := NewClient(GetEnv("BSKY_HANDLE"), GetEnv("BSKY_APP_PASSWORD"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init Bluesky Client: %v\n", err)
		os.Exit(1)
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not read file: %v\n", err)
		os.Exit(1)
	}

	var file URLsFile
	if err := json.Unmarshal(data, &file); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not parse JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetching %d posts from file...\n", len(file.URLs))

	for i, url := range file.URLs {
		post, err := bskyClient.GetPostURL(ctx, url)
		if err != nil {
			fmt.Printf("  [%d/%d] ERROR: %v\n", i+1, len(file.URLs), err)
			continue
		}
		if err := db.SavePost(post); err != nil {
			fmt.Printf("  [%d/%d] ERROR saving: %v\n", i+1, len(file.URLs), err)
			continue
		}
		fmt.Printf("  [%d/%d] Saved: %s\n", i+1, len(file.URLs), truncate(post.Text, 50))
	}

	fmt.Println("Done!")
}
