package schwartz

type SchwartzValues map[string]int

type AIStats struct {
	Model            string  `json:"model"`
	ResponseTimeMs   int64   `json:"response_time_ms"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	CostUsd          float64 `json:"cost_usd"`
	Provider         string  `json:"provider"`
}

type ValueAnalysis struct {
	Rating    SchwartzValues `json:"Rating"`
	Reasoning string         `json:"Reasoning"`
	Score     int            `json:"Score"`
	Stats     AIStats        `json:"Stats"`
	Model     string         `json:"Model,omitempty"`
	Provider  string         `json:"Provider,omitempty"`
	Error     string         `json:"error,omitempty"`
}

type PostImage struct {
	Alt   string `json:"Alt"`
	Image string `json:"Image"`
}

type PostLink struct {
	Uri         string `json:"Uri"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Thumb       string `json:"Thumb"`
}

type PostFacet struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

type Post struct {
	URL           string        `json:"URL"`
	AtURI         string        `json:"AtURI"`
	Text          string        `json:"Text"`
	CreatedAt     string        `json:"CreatedAt"`
	Labels        []string      `json:"Labels"`
	Langs         []string      `json:"Langs"`
	Tags          []string      `json:"Tags"`
	Images        []PostImage   `json:"Images"`
	Links         []PostLink    `json:"Links"`
	Facets        []PostFacet   `json:"Facets"`
	AuthorName    string        `json:"AuthorName"`
	ReplyRoot     string        `json:"ReplyRoot"`
	ReplyParent   string        `json:"ReplyParent"`
	LikeCount     int           `json:"likeCount"`
	ReplyCount    int           `json:"replyCount"`
	RepostCount   int           `json:"repostCount"`
	QuoteCount    int           `json:"quoteCount"`
	ValueAnalysis ValueAnalysis `json:"ValueAnalysis"`
}
