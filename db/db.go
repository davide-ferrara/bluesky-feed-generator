package db

import (
	"bsky-schwartz/pkg/schwartz"
	"database/sql"
	"encoding/json"
	"fmt"

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
		did TEXT PRIMARY KEY,
		handle TEXT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	weightsTable := `
	CREATE TABLE IF NOT EXISTS weights (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		did TEXT NOT NULL,
		value_id TEXT NOT NULL,
		weight REAL NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (did) REFERENCES users(did) ON DELETE CASCADE,
		UNIQUE(did, value_id)
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
		FOREIGN KEY (post_at_uri) REFERENCES posts(at_uri) ON DELETE CASCADE
	);
	`

	if _, err := DB.Exec(analysesTable); err != nil {
		return fmt.Errorf("could not create analyses table: %w", err)
	}

	return nil
}

func SavePost(post schwartz.Post) error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM posts WHERE at_uri = ?", post.AtURI).Scan(&count)
	if err != nil {
		return fmt.Errorf("could not check post existence: %w", err)
	}
	if count > 0 {
		fmt.Printf("Post already exists: %s\n", post.AtURI)
		return nil
	}

	langsJSON, _ := json.Marshal(post.Langs)
	tagsJSON, _ := json.Marshal(post.Tags)
	imagesJSON, _ := json.Marshal(post.Images)
	linksJSON, _ := json.Marshal(post.Links)
	facetsJSON, _ := json.Marshal(post.Facets)

	_, err = DB.Exec(`INSERT INTO posts (at_uri, url, text, created_at, langs, tags, images, links, facets, author_name, reply_root, reply_parent, likes, replies, reposts, quotes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, post.AtURI, post.URL, post.Text, post.CreatedAt, string(langsJSON), string(tagsJSON), string(imagesJSON), string(linksJSON), string(facetsJSON), post.AuthorName, post.ReplyRoot, post.ReplyParent, post.LikeCount, post.ReplyCount, post.RepostCount, post.QuoteCount)
	return err
}

func UpdatePostStats(atURI string, likes, replies, reposts, quotes int) error {
	_, err := DB.Exec(`UPDATE posts SET likes = ?, replies = ?, reposts = ?, quotes = ? WHERE at_uri = ?`, likes, replies, reposts, quotes, atURI)
	return err
}

func SaveAnalysis(postAtURI string, model string, provider string, analysis schwartz.ValueAnalysis) error {
	statsJSON, _ := json.Marshal(analysis.Stats)
	rating := analysis.Rating

	query := `INSERT INTO analyses (post_at_uri, model, provider, reasoning, score, stats, reputation, power, wealth, achievement, pleasure, independent_thoughts, independent_actions, stimulation, personal_security, societal_security, tradition, lawfulness, respect, humility, responsibility, caring, equality, nature, tolerance) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := DB.Exec(query,
		postAtURI,
		model,
		provider,
		analysis.Reasoning,
		analysis.Score,
		string(statsJSON),
		rating["reputation"],
		rating["power"],
		rating["wealth"],
		rating["achievement"],
		rating["pleasure"],
		rating["independent_thoughts"],
		rating["independent_actions"],
		rating["stimulation"],
		rating["personal_security"],
		rating["societal_security"],
		rating["tradition"],
		rating["lawfulness"],
		rating["respect"],
		rating["humility"],
		rating["responsibility"],
		rating["caring"],
		rating["equality"],
		rating["nature"],
		rating["tolerance"],
	)
	return err
}

func SaveUser(handle string, did string) error {
	_, err := DB.Exec(`INSERT INTO users (did, handle, updated_at) VALUES (?, ?, datetime('now')) ON CONFLICT(did) DO UPDATE SET handle = excluded.handle, updated_at = excluded.updated_at`, did, handle)
	return err
}

func SaveWeight(did string, valueID string, weight float64) error {
	_, err := DB.Exec(`INSERT INTO weights (did, value_id, weight, updated_at) VALUES (?, ?, ?, datetime('now')) ON CONFLICT(did, value_id) DO UPDATE SET weight = excluded.weight, updated_at = excluded.updated_at`, did, valueID, weight)
	return err
}

func SaveWeights(did string, weights map[string]float64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO weights (did, value_id, weight, updated_at) VALUES (?, ?, ?, datetime('now')) ON CONFLICT(did, value_id) DO UPDATE SET weight = excluded.weight, updated_at = excluded.updated_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for valueID, weight := range weights {
		if _, err := stmt.Exec(did, valueID, weight); err != nil {
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

	rows, err := DB.Query("SELECT value_id, weight FROM weights WHERE did = ?", did)
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

func GetWeightsByDID(did string) (map[string]float64, error) {
	rows, err := DB.Query("SELECT value_id, weight FROM weights WHERE did = ?", did)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	weights := make(map[string]float64)
	for rows.Next() {
		var valueID string
		var weight float64
		if err := rows.Scan(&valueID, &weight); err != nil {
			return nil, err
		}
		weights[valueID] = weight
	}

	return weights, nil
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

func GetPostCountForModel(model string) (map[string]int, error) {
	query := `
		SELECT post_at_uri, COUNT(*) as count
		FROM analyses
		WHERE model LIKE ?
		GROUP BY post_at_uri
	`
	rows, err := DB.Query(query, model+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var atURI string
		var count int
		if err := rows.Scan(&atURI, &count); err != nil {
			return nil, err
		}
		counts[atURI] = count
	}
	return counts, nil
}

func GetPostsNeedingAnalysis(model string, maxAnalyses int) ([]schwartz.Post, error) {
	query := `
		SELECT p.at_uri, p.url, p.text, p.created_at, p.langs, p.tags, p.images, p.links, p.facets, p.author_name, p.reply_root, p.reply_parent, p.likes, p.replies, p.reposts, p.quotes
		FROM posts p
		LEFT JOIN analyses a ON a.post_at_uri = p.at_uri AND a.model = ?
		GROUP BY p.at_uri
		HAVING COUNT(a.id) < ?
	`
	rows, err := DB.Query(query, model, maxAnalyses)
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

func GetRandomPosts(model string, maxAnalyses int, limit int) ([]schwartz.Post, error) {
	query := `
		SELECT p.at_uri, p.url, p.text, p.created_at, p.langs, p.tags, p.images, p.links, p.facets, p.author_name, p.reply_root, p.reply_parent, p.likes, p.replies, p.reposts, p.quotes
		FROM posts p
		LEFT JOIN analyses a ON a.post_at_uri = p.at_uri AND a.model = ?
		GROUP BY p.at_uri
		HAVING COUNT(a.id) < ?
		ORDER BY RANDOM()
		LIMIT ?
	`
	rows, err := DB.Query(query, model, maxAnalyses, limit)
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
		var score float64
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

func GetPostsWithAnalysisAveraged(model string) ([]schwartz.Post, error) {
	query := `
		SELECT p.at_uri, p.url, p.text, p.created_at, p.langs, p.tags, p.images, p.links, p.facets, p.author_name, p.reply_root, p.reply_parent, p.likes, p.replies, p.reposts, p.quotes,
			   AVG(a.reputation) as reputation,
			   AVG(a.power) as power,
			   AVG(a.wealth) as wealth,
			   AVG(a.achievement) as achievement,
			   AVG(a.pleasure) as pleasure,
			   AVG(a.independent_thoughts) as independent_thoughts,
			   AVG(a.independent_actions) as independent_actions,
			   AVG(a.stimulation) as stimulation,
			   AVG(a.personal_security) as personal_security,
			   AVG(a.societal_security) as societal_security,
			   AVG(a.tradition) as tradition,
			   AVG(a.lawfulness) as lawfulness,
			   AVG(a.respect) as respect,
			   AVG(a.humility) as humility,
			   AVG(a.responsibility) as responsibility,
			   AVG(a.caring) as caring,
			   AVG(a.equality) as equality,
			   AVG(a.nature) as nature,
			   AVG(a.tolerance) as tolerance
FROM posts p
		JOIN analyses a ON p.at_uri = a.post_at_uri
		WHERE a.model LIKE ?
		GROUP BY p.at_uri
	`
	rows, err := DB.Query(query, "%"+model+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []schwartz.Post
	for rows.Next() {
		var post schwartz.Post
		var langsJSON, tagsJSON, imagesJSON, linksJSON, facetsJSON string
		var reputation, power, wealth, achievement, pleasure float64
		var independentThoughts, independentActions, stimulation float64
		var personalSecurity, societalSecurity, tradition, lawfulness float64
		var respect, humility, responsibility, caring, equality, nature, tolerance float64
		var likes, replies, reposts, quotes int

		err := rows.Scan(&post.AtURI, &post.URL, &post.Text, &post.CreatedAt, &langsJSON, &tagsJSON, &imagesJSON, &linksJSON, &facetsJSON, &post.AuthorName, &post.ReplyRoot, &post.ReplyParent, &likes, &replies, &reposts, &quotes,
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
			"reputation":           int(reputation),
			"power":                int(power),
			"wealth":               int(wealth),
			"achievement":          int(achievement),
			"pleasure":             int(pleasure),
			"independent_thoughts": int(independentThoughts),
			"independent_actions":  int(independentActions),
			"stimulation":          int(stimulation),
			"personal_security":    int(personalSecurity),
			"societal_security":    int(societalSecurity),
			"tradition":            int(tradition),
			"lawfulness":           int(lawfulness),
			"respect":              int(respect),
			"humility":             int(humility),
			"responsibility":       int(responsibility),
			"caring":               int(caring),
			"equality":             int(equality),
			"nature":               int(nature),
			"tolerance":            int(tolerance),
		}

		analysis := schwartz.ValueAnalysis{
			Model:  model,
			Rating: rating,
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
		var score float64
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
