package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"bsky-schwartz/db"
	"bsky-schwartz/pkg/schwartz"
)

// Openness to Change: Query focalizzate su esplorazione, indipendenza e novità.
var opennessToChangeQueries = []string{
	"creatività", "innovazione", "libertà", "autonomia", "curiosità",
	"nuove idee", "avventura", "scoperta", "novità", "viaggi",
	"benessere", "cambiamento", "sperimentazione", "indipendenza",
	"originalità", "esplorazione", "fuori dagli schemi", "nuove prospettive",
	"fai da te", "pensiero critico", "nuovi orizzonti", "cambiare tutto",
}

// Self-Enhancement: Query focalizzate su ambizione, status, potere e risorse.
var selfEnhancementQueries = []string{
	"successo", "carriera", "obiettivi", "ambizione", "leadership",
	"influenza", "business", "finanza", "investimenti", "reputazione",
	"prestigio", "eccellenza", "risultati", "merito", "competizione",
	"vincere", "traguardo", "crescita professionale", "status sociale",
	"prendere il comando", "autorità", "gestione", "profitto",
}

// Conservation: Query focalizzate su ordine, sicurezza, tradizione e norme.
var conservationQueries = []string{
	"sicurezza", "famiglia", "stabilità", "ordine", "tradizioni",
	"storia", "radici", "dovere", "legalità", "rispetto",
	"umiltà", "protezione", "istituzioni", "educazione", "valori",
	"integrità", "fedeltà", "buone maniere", "appartenenza", "casa",
	"norme", "rispettare le regole", "preservare",
}

// Self-Transcendence: Query focalizzate su giustizia sociale, empatia e ambiente.
var selfTranscendenceQueries = []string{
	"altruismo", "empatia", "volontariato", "solidarietà", "uguaglianza",
	"diritti umani", "giustizia", "pace", "natura", "ambiente",
	"ecologia", "sostenibilità", "inclusione", "tolleranza", "umanità",
	"bene comune", "salvaguardia", "cambiamento climatico", "diritti civili",
	"aiutare gli altri", "fratellanza", "rispetto universale",
}

var valuesQueries = [][]string{
	opennessToChangeQueries,
	selfEnhancementQueries,
	conservationQueries,
	selfTranscendenceQueries,
}

func GenerateFeed(postsPerCluster int, lang string) {
	if err := db.Init("../data.db"); err != nil {
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

	fmt.Printf("Target: %d total posts, ~%d per cluster\n", *collectLimit, postsPerCluster)

	var mu sync.Mutex
	var allPosts []schwartz.Post

	numOfRoutines := 3
	for clusterIdx, cluster := range valuesQueries {
		fmt.Printf("\n=== Cluster %d/%d (%d queries) ===\n", clusterIdx+1, 4, len(cluster))

		postsPerQuery := postsPerCluster / len(cluster)
		if postsPerQuery < 1 {
			postsPerQuery = 1
		}

		clusterPosts := 0
		for _, q := range cluster {
			if clusterPosts >= postsPerCluster {
				break
			}

			fmt.Printf("Query [%d/%d]: %s (limit: %d)\n", clusterIdx+1, len(cluster), q, postsPerQuery)

			var wg sync.WaitGroup
			sem := make(chan struct{}, numOfRoutines)

			wg.Add(numOfRoutines)
			for range numOfRoutines {
				go func(query string) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					posts, err := bskyClient.SearchPosts(ctx, query, int64(postsPerQuery), lang)
					if err != nil {
						fmt.Printf("  ERROR: %v\n", err)
						return
					}

					mu.Lock()
					clusterPosts += len(posts)
					allPosts = append(allPosts, posts...)
					mu.Unlock()

					fmt.Printf("  Found %d posts (total in cluster: %d)\n", len(posts), clusterPosts)
				}(q)
			}
			wg.Wait()
		}
	}

	fmt.Printf("\n=== Total collected: %d posts ===\n", len(allPosts))

	fmt.Println("Saving posts to database...")
	for _, post := range allPosts {
		if err := db.SavePost(post); err != nil {
			fmt.Printf("ERROR saving post: %v\n", err)
		}
	}

	fmt.Printf("Saved %d posts to database\n", len(allPosts))
}
