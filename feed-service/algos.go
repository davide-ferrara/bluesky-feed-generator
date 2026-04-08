package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "bsky-schwartz/db"
	"bsky-schwartz/internal/models"
	"bsky-schwartz/pkg/schwartz"

	"github.com/bluesky-social/indigo/api/bsky"
)

// AlgoHandler - Firma funzione per un algoritmo di feed
type AlgoHandler func(ctx context.Context, limit int, cursor string, userDID string) (*bsky.FeedGetFeedSkeleton_Output, error)

// Algos - Registro degli algoritmi disponibili
var Algos = map[string]AlgoHandler{
	// "values":     feedValueBased,
	// "engagement": feedEngagement,
	"values": feedEngagement,
}

// =============================================================================
// POST DATABASE
// =============================================================================

// GetFeedPosts returns posts, optionally including unanalyzed posts
// includeUnanalyzed: if true, returns ALL posts (analyzed + unanalyzed)
//
//	if false, returns only posts with analysis
func GetFeedPosts(includeUnanalyzed bool) ([]schwartz.Post, error) {
	if includeUnanalyzed {
		return dbpkg.GetAllPosts()
	}
	return dbpkg.GetPostsWithAnalysis("openai/gpt-4.1-mini")
}

// =============================================================================
// SCORING
// =============================================================================

func jsonKeyToWeightKey(jsonKey string) string {
	return strings.ReplaceAll(jsonKey, " ", "_")
}

// CalculateScore - Calcola lo score di un post rispetto ai weights
func CalculateScore(posts []schwartz.Post, weights map[string]float64) {
	for i := range posts {
		var score int
		for key, rating := range posts[i].ValueAnalysis.Rating {
			score += rating * int(weights[strings.ToLower(key)])
		}
		posts[i].ValueAnalysis.Score = score
	}
}

// calculateEngagementScore - Calcola lo score di engagement
// score = log(engagement+1) / log(age_hours+2)
func calculateEngagementScore(post schwartz.Post) int {
	total := post.LikeCount + post.ReplyCount + post.RepostCount + post.QuoteCount
	if total == 0 {
		return 0
	}

	created, err := time.Parse(time.RFC3339, post.CreatedAt)
	if err != nil {
		return 0
	}
	ageHours := time.Since(created).Hours()
	if ageHours < 1 {
		ageHours = 1
	}

	score := math.Log(float64(total)+1) / math.Log(ageHours+2)
	return int(score * 1000)
}

// =============================================================================
// FEED GENERATORS
// =============================================================================

// feedEngagement - Feed basato su engagement score (log scale con decay temporale)
func feedEngagement(ctx context.Context, limit int, cursor string, userDID string) (*bsky.FeedGetFeedSkeleton_Output, error) {
	posts, err := GetFeedPosts(true) // tutti i post
	if err != nil {
		return nil, fmt.Errorf("loading feed: %w", err)
	}

	for i := range posts {
		posts[i].ValueAnalysis.Score = calculateEngagementScore(posts[i])
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ValueAnalysis.Score > posts[j].ValueAnalysis.Score
	})

	startIdx := 0
	if cursor != "" {
		idx, err := strconv.Atoi(cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		startIdx = idx
	}

	endIdx := startIdx + limit
	if endIdx > len(posts) {
		endIdx = len(posts)
	}

	postSlice := posts[startIdx:endIdx]

	feed := make([]*bsky.FeedDefs_SkeletonFeedPost, len(postSlice))
	for i, sp := range postSlice {
		feed[i] = &bsky.FeedDefs_SkeletonFeedPost{
			Post: sp.AtURI,
		}
	}

	var nextCursor *string
	if endIdx < len(posts) {
		c := strconv.Itoa(endIdx)
		nextCursor = &c
	}

	return &bsky.FeedGetFeedSkeleton_Output{
		Cursor: nextCursor,
		Feed:   feed,
	}, nil
}

// feedValueBased - Feed basato sui values Schwartz
func feedValueBased(ctx context.Context, limit int, cursor string, userDID string) (*bsky.FeedGetFeedSkeleton_Output, error) {
	weights := GlobalUserWeights.Get(userDID)
	if weights == nil {
		weights = make(map[string]float64)
		for _, v := range models.SwartzValues {
			weights[v.ID] = 0.0
		}
	}

	posts, err := GetFeedPosts(false) // solo post analizzati
	if err != nil {
		return nil, fmt.Errorf("loading feed: %w", err)
	}

	CalculateScore(posts, weights)

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ValueAnalysis.Score > posts[j].ValueAnalysis.Score
	})

	startIdx := 0
	if cursor != "" {
		idx, err := strconv.Atoi(cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		startIdx = idx
	}

	endIdx := startIdx + limit
	if endIdx > len(posts) {
		endIdx = len(posts)
	}

	postSlice := posts[startIdx:endIdx]

	feed := make([]*bsky.FeedDefs_SkeletonFeedPost, len(postSlice))
	for i, sp := range postSlice {
		feed[i] = &bsky.FeedDefs_SkeletonFeedPost{
			Post: sp.AtURI,
		}
	}

	var nextCursor *string
	if endIdx < len(posts) {
		c := strconv.Itoa(endIdx)
		nextCursor = &c
	}

	return &bsky.FeedGetFeedSkeleton_Output{
		Cursor: nextCursor,
		Feed:   feed,
	}, nil
}
