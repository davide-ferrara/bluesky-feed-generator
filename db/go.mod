module bsky-schwartz/db

go 1.26

require (
	bsky-schwartz/pkg/schwartz v0.0.0
	github.com/mattn/go-sqlite3 v1.14.40
)

replace bsky-schwartz/pkg/schwartz => ../pkg/schwartz
