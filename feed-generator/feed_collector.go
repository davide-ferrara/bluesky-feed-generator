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

const (
	MinEngagement   = 0
	TargetLang      = "en"
	IncludeRetweets = false
	RateLimitPerSec = 10
)

type FeedConfig struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Limit    int    `json:"limit"`
	Handle   string
	FeedName string
}

type FeedCollector struct {
	client  *Client
	feeds   []FeedConfig
	limiter *rate.Limiter
}

func NewFeedCollector(client *Client, feedsFile string) (*FeedCollector, error) {
	data, err := os.ReadFile(feedsFile)
	if err != nil {
		return nil, fmt.Errorf("read feeds file: %w", err)
	}

	var config struct {
		Feeds []FeedConfig `json:"feeds"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse feeds file: %w", err)
	}

	for i := range config.Feeds {
		if err := parseFeedURL(&config.Feeds[i]); err != nil {
			return nil, fmt.Errorf("parse feed URL %s: %w", config.Feeds[i].Name, err)
		}
	}

	return &FeedCollector{
		client:  client,
		feeds:   config.Feeds,
		limiter: rate.NewLimiter(RateLimitPerSec, RateLimitPerSec),
	}, nil
}

func parseFeedURL(feed *FeedConfig) error {
	re := regexp.MustCompile(`^https://bsky\.app/profile/([^/]+)/feed/([^/]+)$`)
	matches := re.FindStringSubmatch(feed.URL)

	if len(matches) != 3 {
		return fmt.Errorf("invalid feed URL format: %s", feed.URL)
	}

	feed.Handle = matches[1]
	feed.FeedName = matches[2]
	return nil
}

func (fc *FeedCollector) Collect(ctx context.Context) (int, error) {
	totalPosts := 0

	fmt.Printf("Collecting posts from %d feeds...\n", len(fc.feeds))
	fmt.Printf("Filters: lang=%s, engagement>=%d, retweets=%v\n\n",
		TargetLang, MinEngagement, IncludeRetweets)

	for i, feed := range fc.feeds {
		fmt.Printf("[%d/%d] %s (limit: %d)\n", i+1, len(fc.feeds), feed.Name, feed.Limit)

		posts, err := fc.collectFeed(ctx, feed)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		filtered := fc.filterPosts(posts)

		fmt.Printf("  Fetched: %d, Filtered: %d\n", len(posts), len(filtered))

		saved := 0
		for _, post := range filtered {
			if err := db.SavePost(post); err != nil {
				fmt.Printf("  ERROR saving post: %v\n", err)
				continue
			}
			saved++
		}

		fmt.Printf("  Saved: %d posts\n\n", saved)
		totalPosts += saved
	}

	fmt.Printf("Total: %d posts saved from %d feeds\n", totalPosts, len(fc.feeds))
	return totalPosts, nil
}

func (fc *FeedCollector) collectFeed(ctx context.Context, feed FeedConfig) ([]schwartz.Post, error) {
	allPosts := []schwartz.Post{}
	cursor := ""
	batchSize := 100

	for len(allPosts) < feed.Limit {
		if err := fc.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		posts, nextCursor, err := fc.client.GetFeed(ctx, feed.Handle, feed.FeedName, batchSize, cursor)
		if err != nil {
			if isRateLimitError(err) {
				fmt.Printf("  Rate limited, waiting 5 minutes...\n")
				time.Sleep(5 * time.Minute)
				continue
			}
			return allPosts, fmt.Errorf("get feed: %w", err)
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

	if len(allPosts) > feed.Limit {
		allPosts = allPosts[:feed.Limit]
	}

	return allPosts, nil
}

func (fc *FeedCollector) filterPosts(posts []schwartz.Post) []schwartz.Post {
	filtered := []schwartz.Post{}

	for _, post := range posts {
		if !fc.hasLang(post, TargetLang) {
			continue
		}

		engagement := post.LikeCount + post.ReplyCount + post.RepostCount + post.QuoteCount
		if engagement < MinEngagement {
			continue
		}

		// TODO: Implement repost detection when Post struct has Reason field
		// if !IncludeRetweets && fc.isReasonRepost(post) {
		//	continue
		// }

		filtered = append(filtered, post)
	}

	return filtered
}

func (fc *FeedCollector) hasLang(post schwartz.Post, lang string) bool {
	for _, l := range post.Langs {
		if l == lang {
			return true
		}
	}
	return false
}

func isRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "429") ||
		strings.Contains(err.Error(), "rate limit")
}

func CollectFromFeeds() {
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

	feedsFile := "feeds.json"
	if flag.NArg() > 1 {
		feedsFile = flag.Arg(1)
	}

	collector, err := NewFeedCollector(client, feedsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not create feed collector: %v\n", err)
		os.Exit(1)
	}

	if _, err := collector.Collect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
