package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	port := env("PORT", "8080")
	ingestionBase := mustURL(env("INGESTION_URL", ""))
	casesBase := mustURL(env("CASES_URL", ""))
	corsOrigin := env("CORS_ORIGIN", "*")

	client := &http.Client{Timeout: 15 * time.Second}

	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"ok\":true}"))
	})

	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		applyCORS(w, corsOrigin)
		w.WriteHeader(http.StatusNoContent)
	})

	r.Route("/api", func(api chi.Router) {
		api.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				applyCORS(w, corsOrigin)
				next.ServeHTTP(w, r)
			})
		})

		api.Post("/ingest", func(w http.ResponseWriter, r *http.Request) {
			proxy(client, w, r, ingestionBase, "/ingest")
		})

		api.Get("/cases", func(w http.ResponseWriter, r *http.Request) {
			proxy(client, w, r, casesBase, "/cases")
		})

		api.Get("/cases/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			proxy(client, w, r, casesBase, "/cases/"+url.PathEscape(id))
		})
	})

	addr := ":" + port
	log.Printf("api-gateway listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func proxy(client *http.Client, w http.ResponseWriter, r *http.Request, base *url.URL, path string) {
	// Preserve querystring for GET /cases
	u := *base
	u.Path = strings.TrimRight(u.Path, "/") + path
	u.RawQuery = r.URL.RawQuery

	bodyBytes, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()

	req, err := http.NewRequestWithContext(r.Context(), r.Method, u.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}

	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func applyCORS(w http.ResponseWriter, origin string) {
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
}

func mustURL(raw string) *url.URL {
	if raw == "" {
		log.Fatal("missing required URL env var")
	}
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("invalid URL %q: %v", raw, err)
	}
	return u
}

func env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
