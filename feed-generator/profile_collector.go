package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"bsky-schwartz/db"
	"bsky-schwartz/pkg/schwartz"

	"golang.org/x/time/rate"
)

type ProfileConfig struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	URL   string `json:"url"`
	Limit int    `json:"limit"`
}

type ProfileCollector struct {
	client   *Client
	profiles []ProfileConfig
	limiter  *rate.Limiter
}

func NewProfileCollector(client *Client, profilesFile string) (*ProfileCollector, error) {
	data, err := os.ReadFile(profilesFile)
	if err != nil {
		return nil, fmt.Errorf("read profiles file: %w", err)
	}

	var config struct {
		Profiles []ProfileConfig `json:"profiles"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse profiles file: %w", err)
	}

	return &ProfileCollector{
		client:   client,
		profiles: config.Profiles,
		limiter:  rate.NewLimiter(10, 10),
	}, nil
}

func (pc *ProfileCollector) Collect(ctx context.Context) (int, error) {
	totalPosts := 0

	fmt.Printf("Collecting posts from %d profiles...\n", len(pc.profiles))
	fmt.Printf("Filters: lang=%s, engagement>=%d\n\n", TargetLang, MinEngagement)

	for i, profile := range pc.profiles {
		fmt.Printf("[%d/%d] %s (limit: %d)\n", i+1, len(pc.profiles), profile.Name, profile.Limit)

		// Extract handle from URL
		handle := extractHandleFromURL(profile.URL)
		if handle == "" {
			fmt.Printf("  ERROR: invalid profile URL: %s\n", profile.URL)
			continue
		}

		posts, err := pc.collectProfile(ctx, profile.Name, handle, profile.Limit)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		filtered := pc.filterPosts(posts)

		fmt.Printf("  Fetched: %d, Filtered: %d\n", len(posts), len(filtered))

		saved := 0
		for _, post := range filtered {
			if err := db.SavePost(post); err != nil {
				if strings.Contains(err.Error(), "Post already exists") {
					continue
				}
				fmt.Printf("  ERROR saving post: %v\n", err)
				continue
			}
			saved++
		}

		fmt.Printf("  Saved: %d posts\n\n", saved)
		totalPosts += saved
	}

	fmt.Printf("Total: %d posts saved from %d profiles\n", totalPosts, len(pc.profiles))
	return totalPosts, nil
}

func (pc *ProfileCollector) collectProfile(ctx context.Context, name, handle string, limit int) ([]schwartz.Post, error) {
	allPosts := []schwartz.Post{}
	cursor := ""
	batchSize := 100

	for len(allPosts) < limit {
		if err := pc.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		posts, nextCursor, err := pc.client.GetAuthorPosts(ctx, handle, batchSize, cursor)
		if err != nil {
			if isRateLimitError(err) {
				fmt.Printf("  Rate limited, waiting 5 minutes...\n")
				time.Sleep(5 * time.Minute)
				continue
			}
			return allPosts, fmt.Errorf("get author posts: %w", err)
		}

		if len(posts) == 0 {
			break
		}

		allPosts = append(allPosts, posts...)

		if nextCursor == "" || nextCursor == cursor {
			break
		}
		cursor = nextCursor
	}

	if len(allPosts) > limit {
		allPosts = allPosts[:limit]
	}

	return allPosts, nil
}

func (pc *ProfileCollector) filterPosts(posts []schwartz.Post) []schwartz.Post {
	filtered := []schwartz.Post{}

	for _, post := range posts {
		// No language filter - accept all posts

		engagement := post.LikeCount + post.ReplyCount + post.RepostCount + post.QuoteCount
		if engagement < 0 {
			continue
		}

		filtered = append(filtered, post)
	}

	return filtered
}

func (pc *ProfileCollector) hasLang(post schwartz.Post, lang string) bool {
	for _, l := range post.Langs {
		if l == lang {
			return true
		}
	}
	return false
}

func extractHandleFromURL(url string) string {
	re := regexp.MustCompile(`^https://bsky\.app/profile/([^/]+)$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func isProfileRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "429") ||
		strings.Contains(err.Error(), "rate limit")
}

func CollectFromProfiles() {
	if err := db.Init("../data.db"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	client, err := NewClient(GetEnv("BSKY_HANDLE"), GetEnv("BSKY_APP_PASSWORD"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not init Bluesky Client: %v\n", err)
		os.Exit(1)
	}

	profilesFile := "profiles.json"
	if flag.NArg() > 1 {
		profilesFile = flag.Arg(1)
	}

	collector, err := NewProfileCollector(client, profilesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not create profile collector: %v\n", err)
		os.Exit(1)
	}

	if _, err := collector.Collect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
