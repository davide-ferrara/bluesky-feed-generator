package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func handleWellKnown(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(cfg.ServiceDID, cfg.Hostname) {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]any{
			"@context": []string{"https://www.w3.org/ns/did/v1"},
			"id":       cfg.ServiceDID,
			"service": []map[string]any{
				{
					"id":              "#bsky_fg",
					"type":            "BskyFeedGenerator",
					"serviceEndpoint": fmt.Sprintf("https://%s", cfg.Hostname),
				},
			},
		})
	}
}

func handleDescribeFeedGenerator(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feeds := make([]map[string]string, 0, len(Algos))
		for shortname := range Algos {
			feeds = append(feeds, map[string]string{
				"uri": fmt.Sprintf("at://%s/app.bsky.feed.generator/%s", cfg.PublisherDID, shortname),
			})
		}
		writeJSON(w, map[string]any{
			"did":   cfg.ServiceDID,
			"feeds": feeds,
		})
	}
}

func handleGetFeedSkeleton(cfg *Config, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userDID string

		// TODO: IMPROVE JWT PASRING
		if auth := r.Header.Get("Authorization"); auth != "" {

			parts := strings.Split(auth, " ")
			tokenStr := parts[1]

			parser := jwt.NewParser()
			token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
			if err != nil {
				logger.Warn("Error in parsing JWT:", err)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				logger.Warn("Error in claming JWT")
			}

			// Il DID è nel campo "iss" (issuer) del token
			userDID, ok = claims["iss"].(string)
			if !ok {
				logger.Warn("Unauthenticated feed reqeust")
			}
		}

		fmt.Println("[DEBUG] REQEUST FROM:", userDID)

		feedParam := r.URL.Query().Get("feed")
		if feedParam == "" {
			writeError(w, http.StatusBadRequest, "missing feed parameter")
			return
		}

		parts := strings.Split(strings.TrimPrefix(feedParam, "at://"), "/")
		if len(parts) != 3 {
			writeError(w, http.StatusBadRequest, "invalid feed URI")
			return
		}
		did, collection, rkey := parts[0], parts[1], parts[2]

		if did != cfg.PublisherDID || collection != "app.bsky.feed.generator" {
			writeError(w, http.StatusBadRequest, "UnsupportedAlgorithm")
			return
		}

		algo, ok := Algos[rkey]
		if !ok {
			writeError(w, http.StatusBadRequest, "UnsupportedAlgorithm")
			return
		}

		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
				limit = min(parsed, 200)
			}
		}
		cursor := r.URL.Query().Get("cursor")

		// Chiamata all'algoritmo di generazione del feed
		result, err := algo(r.Context(), limit, cursor, userDID)
		if err != nil {
			logger.Error("algo error", "algo", rkey, "err", err)
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, result)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
