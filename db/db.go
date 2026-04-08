package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"bsky-schwartz/pkg/schwartz"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("could not create tables: %w", err)
	}

	return nil
}

func createTables() error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		handle TEXT PRIMARY KEY,
		did TEXT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	weightsTable := `
	CREATE TABLE IF NOT EXISTS weights (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_handle TEXT NOT NULL,
		value_id TEXT NOT NULL,
		weight REAL NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_handle) REFERENCES users(handle) ON DELETE CASCADE,
		UNIQUE(user_handle, value_id)
	);
	`

	if _, err := DB.Exec(usersTable); err != nil {
		return fmt.Errorf("could not create users table: %w", err)
	}

	if _, err := DB.Exec(weightsTable); err != nil {
		return fmt.Errorf("could not create weights table: %w", err)
	}

	postsTable := `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		at_uri TEXT UNIQUE NOT NULL,
		url TEXT NOT NULL,
		text TEXT,
		created_at TEXT,
		langs TEXT,
		tags TEXT,
		images TEXT,
		links TEXT,
		facets TEXT,
		author_name TEXT,
		reply_root TEXT,
		reply_parent TEXT,
		likes INTEGER DEFAULT 0,
		replies INTEGER DEFAULT 0,
		reposts INTEGER DEFAULT 0,
		quotes INTEGER DEFAULT 0,
		created_at_db TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := DB.Exec(postsTable); err != nil {
		return fmt.Errorf("could not create posts table: %w", err)
	}

	analysesTable := `
	CREATE TABLE IF NOT EXISTS analyses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_at_uri TEXT NOT NULL,
		model TEXT NOT NULL,
		provider TEXT NOT NULL,
		reasoning TEXT,
		score INTEGER DEFAULT 0,
		stats TEXT,
		reputation INTEGER DEFAULT 0,
		power INTEGER DEFAULT 0,
		wealth INTEGER DEFAULT 0,
		achievement INTEGER DEFAULT 0,
		pleasure INTEGER DEFAULT 0,
		independent_thoughts INTEGER DEFAULT 0,
		independent_actions INTEGER DEFAULT 0,
		stimulation INTEGER DEFAULT 0,
		personal_security INTEGER DEFAULT 0,
		societal_security INTEGER DEFAULT 0,
		tradition INTEGER DEFAULT 0,
		lawfulness INTEGER DEFAULT 0,
		respect INTEGER DEFAULT 0,
		humility INTEGER DEFAULT 0,
		responsibility INTEGER DEFAULT 0,
		caring INTEGER DEFAULT 0,
		equality INTEGER DEFAULT 0,
		nature INTEGER DEFAULT 0,
		tolerance INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_at_uri) REFERENCES posts(at_uri) ON DELETE CASCADE,
		UNIQUE(post_at_uri, model)
	);
	`

	if _, err := DB.Exec(analysesTable); err != nil {
		return fmt.Errorf("could not create analyses table: %w", err)
	}

	return nil
}

func SavePost(post schwartz.Post) error {
	langsJSON, _ := json.Marshal(post.Langs)
	tagsJSON, _ := json.Marshal(post.Tags)
	imagesJSON, _ := json.Marshal(post.Images)
	linksJSON, _ := json.Marshal(post.Links)
	facetsJSON, _ := json.Marshal(post.Facets)

	_, err := DB.Exec(`
		INSERT INTO posts (at_uri, url, text, created_at, langs, tags, images, links, facets, author_name, reply_root, reply_parent, likes, replies, reposts, quotes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(at_uri) DO UPDATE SET 
			text = excluded.text,
			author_name = excluded.author_name,
			created_at = excluded.created_at,
			langs = excluded.langs,
			tags = excluded.tags,
			images = excluded.images,
			links = excluded.links,
			facets = excluded.facets,
			reply_root = excluded.reply_root,
			reply_parent = excluded.reply_parent,
			likes = excluded.likes,
			replies = excluded.replies,
			reposts = excluded.reposts,
			quotes = excluded.quotes
	`, post.AtURI, post.URL, post.Text, post.CreatedAt, string(langsJSON), string(tagsJSON), string(imagesJSON), string(linksJSON), string(facetsJSON), post.AuthorName, post.ReplyRoot, post.ReplyParent, post.LikeCount, post.ReplyCount, post.RepostCount, post.QuoteCount)
	return err
}

func SaveAnalysis(postAtURI string, model string, provider string, analysis schwartz.ValueAnalysis) error {
	statsJSON, _ := json.Marshal(analysis.Stats)

	rating := analysis.Rating
	_, err := DB.Exec(`
		INSERT INTO analyses (post_at_uri, model, provider, reasoning, score, stats,
			reputation, power, wealth, achievement, pleasure,
			independent_thoughts, independent_actions, stimulation,
			personal_security, societal_security, tradition, lawfulness,
			respect, humility, responsibility, caring, equality, nature, tolerance)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(post_at_uri, model) DO UPDATE SET
			reasoning = excluded.reasoning,
			score = excluded.score,
			stats = excluded.stats,
			reputation = excluded.reputation,
			power = excluded.power,
			wealth = excluded.wealth,
			achievement = excluded.achievement,
			pleasure = excluded.pleasure,
			independent_thoughts = excluded.independent_thoughts,
			independent_actions = excluded.independent_actions,
			stimulation = excluded.stimulation,
			personal_security = excluded.personal_security,
			societal_security = excluded.societal_security,
			tradition = excluded.tradition,
			lawfulness = excluded.lawfulness,
			respect = excluded.respect,
			humility = excluded.humility,
			responsibility = excluded.responsibility,
			caring = excluded.caring,
			equality = excluded.equality,
			nature = excluded.nature,
			tolerance = excluded.tolerance
	`, postAtURI, model, provider, analysis.Reasoning, analysis.Score, string(statsJSON),
		rating["reputation"], rating["power"], rating["wealth"], rating["achievement"], rating["pleasure"],
		rating["independent thoughts"], rating["independent actions"], rating["stimulation"],
		rating["personal security"], rating["societal security"], rating["tradition"], rating["lawfulness"],
		rating["respect"], rating["humility"], rating["responsibility"], rating["caring"], rating["equality"], rating["nature"], rating["tolerance"])
	return err
}

func SaveUser(handle, did string) error {
	_, err := DB.Exec(`
		INSERT INTO users (handle, did, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(handle) DO UPDATE SET did = excluded.did, updated_at = excluded.updated_at
	`, handle, did)
	return err
}

func SaveWeight(userHandle, valueID string, weight float64) error {
	_, err := DB.Exec(`
		INSERT INTO weights (user_handle, value_id, weight, updated_at)
		VALUES (?, ?, ?, datetime('now'))
		ON CONFLICT(user_handle, value_id) DO UPDATE SET weight = excluded.weight, updated_at = excluded.updated_at
	`, userHandle, valueID, weight)
	return err
}

func SaveWeights(userHandle string, weights map[string]float64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO weights (user_handle, value_id, weight, updated_at)
		VALUES (?, ?, ?, datetime('now'))
		ON CONFLICT(user_handle, value_id) DO UPDATE SET weight = excluded.weight, updated_at = excluded.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for valueID, weight := range weights {
		if _, err := stmt.Exec(userHandle, valueID, weight); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetUserFromDB(handle string) (string, map[string]float64, error) {
	var did string
	err := DB.QueryRow("SELECT did FROM users WHERE handle = ?", handle).Scan(&did)
	if err != nil {
		return "", nil, fmt.Errorf("user not found: %w", err)
	}

	rows, err := DB.Query("SELECT value_id, weight FROM weights WHERE user_handle = ?", handle)
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()

	weights := make(map[string]float64)
	for rows.Next() {
		var valueID string
		var weight float64
		if err := rows.Scan(&valueID, &weight); err != nil {
			return "", nil, err
		}
		weights[valueID] = weight
	}

	return did, weights, nil
}

func GetUnanalyzedPosts() ([]schwartz.Post, error) {
	rows, err := DB.Query(`
		SELECT p.at_uri, p.url, p.text, p.created_at, p.langs, p.tags, p.images, p.links, p.facets, p.author_name, p.reply_root, p.reply_parent, p.likes, p.replies, p.reposts, p.quotes
		FROM posts p
		WHERE NOT EXISTS (SELECT 1 FROM analyses a WHERE a.post_at_uri = p.at_uri)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []schwartz.Post
	for rows.Next() {
		var post schwartz.Post
		var langsJSON, tagsJSON, imagesJSON, linksJSON, facetsJSON string
		var likes, replies, reposts, quotes int
		err := rows.Scan(&post.AtURI, &post.URL, &post.Text, &post.CreatedAt, &langsJSON, &tagsJSON, &imagesJSON, &linksJSON, &facetsJSON, &post.AuthorName, &post.ReplyRoot, &post.ReplyParent, &likes, &replies, &reposts, &quotes)
		if err != nil {
			return nil, err
		}
		post.LikeCount = likes
		post.ReplyCount = replies
		post.RepostCount = reposts
		post.QuoteCount = quotes
		json.Unmarshal([]byte(langsJSON), &post.Langs)
		json.Unmarshal([]byte(tagsJSON), &post.Tags)
		json.Unmarshal([]byte(imagesJSON), &post.Images)
		json.Unmarshal([]byte(linksJSON), &post.Links)
		json.Unmarshal([]byte(facetsJSON), &post.Facets)
		posts = append(posts, post)
	}

	return posts, nil
}

func GetAllPosts() ([]schwartz.Post, error) {
	query := `
		SELECT p.at_uri, p.url, p.text, p.created_at, p.langs, p.tags, p.images, p.links, p.facets, p.author_name, p.reply_root, p.reply_parent, p.likes, p.replies, p.reposts, p.quotes
		FROM posts p
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []schwartz.Post
	for rows.Next() {
		var post schwartz.Post
		var langsJSON, tagsJSON, imagesJSON, linksJSON, facetsJSON string
		var likes, replies, reposts, quotes int
		err := rows.Scan(&post.AtURI, &post.URL, &post.Text, &post.CreatedAt, &langsJSON, &tagsJSON, &imagesJSON, &linksJSON, &facetsJSON, &post.AuthorName, &post.ReplyRoot, &post.ReplyParent, &likes, &replies, &reposts, &quotes)
		if err != nil {
			return nil, err
		}
		post.LikeCount = likes
		post.ReplyCount = replies
		post.RepostCount = reposts
		post.QuoteCount = quotes
		json.Unmarshal([]byte(langsJSON), &post.Langs)
		json.Unmarshal([]byte(tagsJSON), &post.Tags)
		json.Unmarshal([]byte(imagesJSON), &post.Images)
		json.Unmarshal([]byte(linksJSON), &post.Links)
		json.Unmarshal([]byte(facetsJSON), &post.Facets)
		posts = append(posts, post)
	}

	return posts, nil
}

func GetPostsWithAnalysis(model string) ([]schwartz.Post, error) {
	query := `
		SELECT p.at_uri, p.url, p.text, p.created_at, p.langs, p.tags, p.images, p.links, p.facets, p.author_name, p.reply_root, p.reply_parent, p.likes, p.replies, p.reposts, p.quotes,
			   a.model, a.provider, a.reasoning, a.score, a.stats,
			   a.reputation, a.power, a.wealth, a.achievement, a.pleasure,
			   a.independent_thoughts, a.independent_actions, a.stimulation,
			   a.personal_security, a.societal_security, a.tradition, a.lawfulness,
			   a.respect, a.humility, a.responsibility, a.caring, a.equality, a.nature, a.tolerance
		FROM posts p
		JOIN analyses a ON p.at_uri = a.post_at_uri
		WHERE a.model = ?
	`

	rows, err := DB.Query(query, model)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []schwartz.Post
	for rows.Next() {
		var post schwartz.Post
		var langsJSON, tagsJSON, imagesJSON, linksJSON, facetsJSON, statsJSON string
		var model, provider, reasoning string
		var score int
		var reputation, power, wealth, achievement, pleasure int
		var independentThoughts, independentActions, stimulation int
		var personalSecurity, societalSecurity, tradition, lawfulness int
		var respect, humility, responsibility, caring, equality, nature, tolerance int
		var likes, replies, reposts, quotes int

		err := rows.Scan(&post.AtURI, &post.URL, &post.Text, &post.CreatedAt, &langsJSON, &tagsJSON, &imagesJSON, &linksJSON, &facetsJSON, &post.AuthorName, &post.ReplyRoot, &post.ReplyParent, &likes, &replies, &reposts, &quotes,
			&model, &provider, &reasoning, &score, &statsJSON,
			&reputation, &power, &wealth, &achievement, &pleasure,
			&independentThoughts, &independentActions, &stimulation,
			&personalSecurity, &societalSecurity, &tradition, &lawfulness,
			&respect, &humility, &responsibility, &caring, &equality, &nature, &tolerance)
		if err != nil {
			return nil, err
		}

		post.LikeCount = likes
		post.ReplyCount = replies
		post.RepostCount = reposts
		post.QuoteCount = quotes

		json.Unmarshal([]byte(langsJSON), &post.Langs)
		json.Unmarshal([]byte(tagsJSON), &post.Tags)
		json.Unmarshal([]byte(imagesJSON), &post.Images)
		json.Unmarshal([]byte(linksJSON), &post.Links)
		json.Unmarshal([]byte(facetsJSON), &post.Facets)

		rating := schwartz.SchwartzValues{
			"reputation":           reputation,
			"power":                power,
			"wealth":               wealth,
			"achievement":          achievement,
			"pleasure":             pleasure,
			"independent thoughts": independentThoughts,
			"independent actions":  independentActions,
			"stimulation":          stimulation,
			"personal security":    personalSecurity,
			"societal security":    societalSecurity,
			"tradition":            tradition,
			"lawfulness":           lawfulness,
			"respect":              respect,
			"humility":             humility,
			"responsibility":       responsibility,
			"caring":               caring,
			"equality":             equality,
			"nature":               nature,
			"tolerance":            tolerance,
		}

		var stats schwartz.AIStats
		json.Unmarshal([]byte(statsJSON), &stats)

		analysis := schwartz.ValueAnalysis{
			Model:     model,
			Provider:  provider,
			Reasoning: reasoning,
			Score:     score,
			Stats:     stats,
			Rating:    rating,
		}

		post.ValueAnalysis = analysis
		posts = append(posts, post)
	}

	return posts, nil
}

func GetPostAnalysis(postAtURI string) ([]schwartz.ValueAnalysis, error) {
	query := `
		SELECT model, provider, reasoning, score, stats,
			   reputation, power, wealth, achievement, pleasure,
			   independent_thoughts, independent_actions, stimulation,
			   personal_security, societal_security, tradition, lawfulness,
			   respect, humility, responsibility, caring, equality, nature, tolerance
		FROM analyses
		WHERE post_at_uri = ?
	`

	rows, err := DB.Query(query, postAtURI)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyses []schwartz.ValueAnalysis
	for rows.Next() {
		var analysis schwartz.ValueAnalysis
		var statsJSON string
		var model, provider, reasoning string
		var score int
		var reputation, power, wealth, achievement, pleasure int
		var independentThoughts, independentActions, stimulation int
		var personalSecurity, societalSecurity, tradition, lawfulness int
		var respect, humility, responsibility, caring, equality, nature, tolerance int

		err := rows.Scan(&model, &provider, &reasoning, &score, &statsJSON,
			&reputation, &power, &wealth, &achievement, &pleasure,
			&independentThoughts, &independentActions, &stimulation,
			&personalSecurity, &societalSecurity, &tradition, &lawfulness,
			&respect, &humility, &responsibility, &caring, &equality, &nature, &tolerance)
		if err != nil {
			return nil, err
		}

		rating := schwartz.SchwartzValues{
			"reputation":           reputation,
			"power":                power,
			"wealth":               wealth,
			"achievement":          achievement,
			"pleasure":             pleasure,
			"independent thoughts": independentThoughts,
			"independent actions":  independentActions,
			"stimulation":          stimulation,
			"personal security":    personalSecurity,
			"societal security":    societalSecurity,
			"tradition":            tradition,
			"lawfulness":           lawfulness,
			"respect":              respect,
			"humility":             humility,
			"responsibility":       responsibility,
			"caring":               caring,
			"equality":             equality,
			"nature":               nature,
			"tolerance":            tolerance,
		}

		json.Unmarshal([]byte(statsJSON), &analysis.Stats)
		analysis.Model = model
		analysis.Provider = provider
		analysis.Reasoning = reasoning
		analysis.Score = score
		analysis.Rating = rating
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
