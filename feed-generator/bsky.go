package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"bsky-schwartz/pkg/schwartz"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
)

func NewClient(handle, appPassword string) (*Client, error) {
	xrpcClient := &xrpc.Client{Host: "https://bsky.social"}

	session, err := atproto.ServerCreateSession(context.Background(), xrpcClient, &atproto.ServerCreateSession_Input{
		Identifier: handle,
		Password:   appPassword,
	})
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		client: &xrpc.Client{
			Host: "https://api.bsky.app",
			Auth: &xrpc.AuthInfo{
				AccessJwt:  session.AccessJwt,
				RefreshJwt: session.RefreshJwt,
			},
		},
	}, nil
}

func (c *Client) SearchPosts(ctx context.Context, q string, limit int64, lang string) ([]schwartz.Post, error) {
	if q == "" {
		return []schwartz.Post{}, nil
	}

	search, err := bsky.FeedSearchPosts(ctx,
		c.client, "", "",
		"", lang, limit, "",
		q, "", "",
		nil, "", "")
	if err != nil {
		return nil, err
	}

	var posts []schwartz.Post
	for _, postView := range search.Posts {
		handle := string(postView.Author.Handle)
		split := strings.Split(postView.Uri, "/")
		key := split[len(split)-1]

		record, ok := postView.Record.Val.(*bsky.FeedPost)
		if !ok {
			continue
		}

		var labels []string
		for _, label := range postView.Labels {
			labels = append(labels, label.Val)
		}

		authorName := postView.Author.Handle
		if postView.Author.DisplayName != nil && *postView.Author.DisplayName != "" {
			authorName = *postView.Author.DisplayName
		}

		var replyRoot, replyParent string
		if record.Reply != nil {
			replyRoot = record.Reply.Root.Uri
			replyParent = record.Reply.Parent.Uri
		}

		url := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", handle, key)
		atURI := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", postView.Author.Did, key)

		likeCount := 0
		if postView.LikeCount != nil {
			likeCount = int(*postView.LikeCount)
		}
		replyCount := 0
		if postView.ReplyCount != nil {
			replyCount = int(*postView.ReplyCount)
		}
		repostCount := 0
		if postView.RepostCount != nil {
			repostCount = int(*postView.RepostCount)
		}
		quoteCount := 0
		if postView.QuoteCount != nil {
			quoteCount = int(*postView.QuoteCount)
		}

		posts = append(posts, schwartz.Post{
			URL:         url,
			AtURI:       atURI,
			Text:        record.Text,
			CreatedAt:   record.CreatedAt,
			Labels:      labels,
			Langs:       record.Langs,
			Tags:        record.Tags,
			Images:      extractImagesView(postView),
			Links:       extractLinksView(postView),
			Facets:      extractFacets(record),
			AuthorName:  authorName,
			ReplyRoot:   replyRoot,
			ReplyParent: replyParent,
			LikeCount:   likeCount,
			ReplyCount:  replyCount,
			RepostCount: repostCount,
			QuoteCount:  quoteCount,
		})
	}

	return posts, nil
}

// GetPost from hanlde and key
func (c *Client) GetPost(ctx context.Context, handle string, key string) (schwartz.Post, error) {
	atURI, err := c.GetAtURI(ctx, handle, key)
	if err != nil {
		return schwartz.Post{}, err
	}

	result, err := bsky.FeedGetPosts(ctx, c.client, []string{atURI})
	if err != nil {
		return schwartz.Post{}, err
	}

	if len(result.Posts) == 0 {
		return schwartz.Post{}, fmt.Errorf("post not found: %s", atURI)
	}

	postView := result.Posts[0]
	record := postView.Record.Val.(*bsky.FeedPost)

	var labels []string
	for _, label := range postView.Labels {
		labels = append(labels, label.Val)
	}

	authorName := postView.Author.Handle
	if postView.Author.DisplayName != nil && *postView.Author.DisplayName != "" {
		authorName = *postView.Author.DisplayName
	}

	var replyRoot, replyParent string
	if record.Reply != nil {
		replyRoot = record.Reply.Root.Uri
		replyParent = record.Reply.Parent.Uri
	}

	url, err := BuildURL(handle, key)
	if err != nil {
		fmt.Println("Error in converting handle and key to URL")
		url = ""
	}

	likeCount := 0
	if postView.LikeCount != nil {
		likeCount = int(*postView.LikeCount)
	}
	replyCount := 0
	if postView.ReplyCount != nil {
		replyCount = int(*postView.ReplyCount)
	}
	repostCount := 0
	if postView.RepostCount != nil {
		repostCount = int(*postView.RepostCount)
	}
	quoteCount := 0
	if postView.QuoteCount != nil {
		quoteCount = int(*postView.QuoteCount)
	}

	return schwartz.Post{
		URL:         url,
		AtURI:       atURI,
		Text:        record.Text,
		CreatedAt:   record.CreatedAt,
		Labels:      labels,
		Langs:       record.Langs,
		Tags:        record.Tags,
		Images:      extractImagesView(postView),
		Links:       extractLinksView(postView),
		Facets:      extractFacets(record),
		AuthorName:  authorName,
		ReplyRoot:   replyRoot,
		ReplyParent: replyParent,
		LikeCount:   likeCount,
		ReplyCount:  replyCount,
		RepostCount: repostCount,
		QuoteCount:  quoteCount,
	}, nil
}

// GetPostURL https://bsky.app/profile/{handle or DID}/post/{rkey}
func (c *Client) GetPostURL(ctx context.Context, url string) (schwartz.Post, error) {
	re := regexp.MustCompile(`^https://bsky\.app/profile/([^/]+)/post/([^/]+)$`)
	matches := re.FindStringSubmatch(url)

	handle := matches[1]
	key := matches[2]

	post, err := c.GetPost(ctx, handle, key)
	if err != nil {
		return schwartz.Post{}, err
	}

	return post, nil
}

// GetAtURI at://did:plc:vwzwgnygau7ed7b7wt5ux7y2/app.bsky.feed.post/3k5nobkf2w72g
func (c *Client) GetAtURI(ctx context.Context, handle string, key string) (string, error) {
	result, err := atproto.IdentityResolveHandle(ctx, c.client, handle)
	if err != nil {
		return "", err
	}

	atURI := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", result.Did, key)

	return atURI, nil
}

// GetFeed retrieves posts from a custom feed generator
func (c *Client) GetFeed(ctx context.Context, handle, feedName string, limit int, cursor string) ([]schwartz.Post, string, error) {
	// Use PDS endpoint for feed requests (bsky.social instead of api.bsky.app)
	pdsClient := &xrpc.Client{
		Host: "https://bsky.social",
		Auth: c.client.Auth,
	}

	result, err := atproto.IdentityResolveHandle(ctx, pdsClient, handle)
	if err != nil {
		return nil, "", fmt.Errorf("resolve handle: %w", err)
	}

	feedURI := fmt.Sprintf("at://%s/app.bsky.feed.generator/%s", result.Did, feedName)

	feed, err := bsky.FeedGetFeed(ctx, pdsClient, cursor, feedURI, int64(limit))
	if err != nil {
		return nil, "", fmt.Errorf("get feed: %w", err)
	}

	posts := []schwartz.Post{}
	for _, item := range feed.Feed {
		postView := item.Post
		if postView == nil {
			continue
		}

		record, ok := postView.Record.Val.(*bsky.FeedPost)
		if !ok {
			continue
		}

		var labels []string
		for _, label := range postView.Labels {
			labels = append(labels, label.Val)
		}

		authorName := string(postView.Author.Handle)
		if postView.Author.DisplayName != nil && *postView.Author.DisplayName != "" {
			authorName = *postView.Author.DisplayName
		}

		split := strings.Split(postView.Uri, "/")
		key := split[len(split)-1]

		var replyRoot, replyParent string
		if record.Reply != nil {
			replyRoot = record.Reply.Root.Uri
			replyParent = record.Reply.Parent.Uri
		}

		likeCount := 0
		if postView.LikeCount != nil {
			likeCount = int(*postView.LikeCount)
		}
		replyCount := 0
		if postView.ReplyCount != nil {
			replyCount = int(*postView.ReplyCount)
		}
		repostCount := 0
		if postView.RepostCount != nil {
			repostCount = int(*postView.RepostCount)
		}
		quoteCount := 0
		if postView.QuoteCount != nil {
			quoteCount = int(*postView.QuoteCount)
		}

		postURL := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", postView.Author.Handle, key)
		atURI := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", postView.Author.Did, key)

		posts = append(posts, schwartz.Post{
			URL:         postURL,
			AtURI:       atURI,
			Text:        record.Text,
			CreatedAt:   record.CreatedAt,
			Labels:      labels,
			Langs:       record.Langs,
			Tags:        record.Tags,
			Images:      extractImagesView(postView),
			Links:       extractLinksView(postView),
			Facets:      extractFacets(record),
			AuthorName:  authorName,
			ReplyRoot:   replyRoot,
			ReplyParent: replyParent,
			LikeCount:   likeCount,
			ReplyCount:  replyCount,
			RepostCount: repostCount,
			QuoteCount:  quoteCount,
		})
	}

	nextCursor := ""
	if feed.Cursor != nil {
		nextCursor = *feed.Cursor
	}

	return posts, nextCursor, nil
}

// GetAuthorPosts retrieves posts from a specific profile/author
func (c *Client) GetAuthorPosts(ctx context.Context, handle string, limit int, cursor string) ([]schwartz.Post, string, error) {
	pdsClient := &xrpc.Client{
		Host: "https://bsky.social",
		Auth: c.client.Auth,
	}

	result, err := bsky.FeedGetAuthorFeed(ctx, pdsClient, handle, cursor, "posts_with_replies", false, int64(limit))
	if err != nil {
		return nil, "", fmt.Errorf("get author feed: %w", err)
	}

	posts := []schwartz.Post{}
	for _, item := range result.Feed {
		postView := item.Post
		if postView == nil {
			continue
		}

		record, ok := postView.Record.Val.(*bsky.FeedPost)
		if !ok {
			continue
		}

		var labels []string
		for _, label := range postView.Labels {
			labels = append(labels, label.Val)
		}

		authorName := string(postView.Author.Handle)
		if postView.Author.DisplayName != nil && *postView.Author.DisplayName != "" {
			authorName = *postView.Author.DisplayName
		}

		split := strings.Split(postView.Uri, "/")
		key := split[len(split)-1]

		var replyRoot, replyParent string
		if record.Reply != nil {
			replyRoot = record.Reply.Root.Uri
			replyParent = record.Reply.Parent.Uri
		}

		likeCount := 0
		if postView.LikeCount != nil {
			likeCount = int(*postView.LikeCount)
		}
		replyCount := 0
		if postView.ReplyCount != nil {
			replyCount = int(*postView.ReplyCount)
		}
		repostCount := 0
		if postView.RepostCount != nil {
			repostCount = int(*postView.RepostCount)
		}
		quoteCount := 0
		if postView.QuoteCount != nil {
			quoteCount = int(*postView.QuoteCount)
		}

		postURL := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", postView.Author.Handle, key)
		atURI := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", postView.Author.Did, key)

		posts = append(posts, schwartz.Post{
			URL:         postURL,
			AtURI:       atURI,
			Text:        record.Text,
			CreatedAt:   record.CreatedAt,
			Labels:      labels,
			Langs:       record.Langs,
			Tags:        record.Tags,
			Images:      extractImagesView(postView),
			Links:       extractLinksView(postView),
			Facets:      extractFacets(record),
			AuthorName:  authorName,
			ReplyRoot:   replyRoot,
			ReplyParent: replyParent,
			LikeCount:   likeCount,
			ReplyCount:  replyCount,
			RepostCount: repostCount,
			QuoteCount:  quoteCount,
		})
	}

	nextCursor := ""
	if result.Cursor != nil {
		nextCursor = *result.Cursor
	}

	return posts, nextCursor, nil
}

func extractImagesView(postView *bsky.FeedDefs_PostView) []schwartz.PostImage {
	var images []schwartz.PostImage
	if postView.Embed == nil || postView.Embed.EmbedImages_View == nil {
		return images
	}
	for _, img := range postView.Embed.EmbedImages_View.Images {
		images = append(images, schwartz.PostImage{
			Alt:   img.Alt,
			Image: img.Fullsize,
		})
	}
	return images
}

func extractLinksView(postView *bsky.FeedDefs_PostView) []schwartz.PostLink {
	var links []schwartz.PostLink
	if postView.Embed == nil || postView.Embed.EmbedExternal_View == nil {
		return links
	}
	ext := postView.Embed.EmbedExternal_View.External
	thumb := ""
	if ext.Thumb != nil {
		thumb = *ext.Thumb
	}
	links = append(links, schwartz.PostLink{
		Uri:         ext.Uri,
		Title:       ext.Title,
		Description: ext.Description,
		Thumb:       thumb,
	})
	return links
}

func extractImages(post *bsky.FeedPost) []schwartz.PostImage {
	var images []schwartz.PostImage
	if post.Embed == nil || post.Embed.EmbedImages == nil {
		return images
	}
	for _, img := range post.Embed.EmbedImages.Images {
		images = append(images, schwartz.PostImage{
			Alt:   img.Alt,
			Image: img.Image.Ref.String(),
		})
	}
	return images
}

func extractLinks(post *bsky.FeedPost) []schwartz.PostLink {
	var links []schwartz.PostLink
	if post.Embed == nil || post.Embed.EmbedExternal == nil {
		return links
	}
	ext := post.Embed.EmbedExternal.External
	thumb := ""
	if ext.Thumb != nil {
		thumb = ext.Thumb.Ref.String()
	}
	links = append(links, schwartz.PostLink{
		Uri:         ext.Uri,
		Title:       ext.Title,
		Description: ext.Description,
		Thumb:       thumb,
	})
	return links
}

func extractFacets(post *bsky.FeedPost) []schwartz.PostFacet {
	var facets []schwartz.PostFacet
	for _, facet := range post.Facets {
		for _, feature := range facet.Features {
			switch {
			case feature.RichtextFacet_Link != nil:
				facets = append(facets, schwartz.PostFacet{
					Type:  "link",
					Value: feature.RichtextFacet_Link.Uri,
				})
			case feature.RichtextFacet_Mention != nil:
				facets = append(facets, schwartz.PostFacet{
					Type:  "mention",
					Value: feature.RichtextFacet_Mention.Did,
				})
			case feature.RichtextFacet_Tag != nil:
				facets = append(facets, schwartz.PostFacet{
					Type:  "tag",
					Value: feature.RichtextFacet_Tag.Tag,
				})
			}
		}
	}
	return facets
}

func SavePostsToJSON(filename string, data interface{}) error {
	filename = fmt.Sprintf("%s_%s.json", filename, time.Now().Format("20060102150405"))
	bytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	if err := os.WriteFile(filename, bytes, 0o644); err != nil {
		return fmt.Errorf("write file error: %w", err)
	}
	return nil
}

// "https://bsky.app/profile/pietrosalvatori.bsky.social/post/3mfqzcs2wck2y",
func BuildURL(handle string, key string) (string, error) {
	if handle == "" || key == "" {
		return "", fmt.Errorf("empty hanlde or key")
	}

	url := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", handle, key)

	return url, nil
}
