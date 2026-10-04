package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aobaiwaki/penlight-v2/pkg/config"
	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/repository"
)

// Server holds server dependencies and HTTP router.
type Server struct {
	cfg          *config.Config
	repo         *repository.SQLiteRepository
	imageSources map[string]string // image_key -> official CDN URL
	mux          *http.ServeMux
}

type imageSourceEntry struct {
	ImageKey string `json:"dest_key"`
	URL      string `json:"url"`
}

func resolveFile(relPath string) ([]byte, error) {
	candidates := []string{
		relPath,
		filepath.Join("..", relPath),
		filepath.Join("..", "..", relPath),
	}
	for _, p := range candidates {
		if bytes, err := os.ReadFile(p); err == nil {
			return bytes, nil
		}
	}
	return nil, fmt.Errorf("file not found in search paths: %s", relPath)
}

// NewServer initializes dependencies, applies migrations and seeds if needed, and returns Server.
func NewServer(cfg *config.Config) (*Server, error) {
	repo, err := repository.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	// 1. Auto-apply migrations
	migrationSQL, err := resolveFile("migrations/000001_init.up.sql")
	if err != nil {
		return nil, fmt.Errorf("failed to read migration SQL: %w", err)
	}
	if _, err := repo.DB().Exec(string(migrationSQL)); err != nil {
		return nil, fmt.Errorf("failed to execute migration: %w", err)
	}

	// 2. Check if seed data needs to be populated
	var groupCount int
	_ = repo.DB().QueryRow("SELECT COUNT(*) FROM groups;").Scan(&groupCount)
	if groupCount == 0 {
		log.Println("Database is empty, loading seeds/seed.sql...")
		seedSQL, err := resolveFile("seeds/seed.sql")
		if err != nil {
			log.Printf("Warning: failed to read seeds/seed.sql: %v", err)
		} else {
			if _, err := repo.DB().Exec(string(seedSQL)); err != nil {
				log.Printf("Warning: failed to execute seeds/seed.sql: %v", err)
			} else {
				log.Println("Seed data successfully loaded into SQLite.")
			}
		}
	}

	// 3. Load image sources map for dynamic CDN redirect fallback
	imageSources := make(map[string]string)
	for _, candidate := range []string{"seeds/data/image_sources.json", "data/image_sources.json"} {
		if srcBytes, err := resolveFile(candidate); err == nil {
			var entries []imageSourceEntry
			if err := json.Unmarshal(srcBytes, &entries); err == nil {
				for _, e := range entries {
					if e.ImageKey != "" && e.URL != "" {
						imageSources[e.ImageKey] = e.URL
					}
				}
				if len(imageSources) > 0 {
					break
				}
			}
		}
	}

	s := &Server{
		cfg:          cfg,
		repo:         repo,
		imageSources: imageSources,
		mux:          http.NewServeMux(),
	}
	s.registerRoutes()
	return s, nil
}

// Handler returns the HTTP handler with global middleware.
func (s *Server) Handler() http.Handler {
	return s.corsMiddleware(s.mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, If-None-Match")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/api/v1/sync/bootstrap", s.handleBootstrap)
	s.mux.HandleFunc("/images/", s.handleImage)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// ETag checking via master_versions
	version, _ := s.repo.GetMasterVersion(ctx)
	var etag string
	if version != nil {
		etag = fmt.Sprintf(`"%s"`, version.Version)
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	groups, err := s.repo.ListGroups(ctx)
	if err != nil {
		http.Error(w, "failed to list groups", http.StatusInternalServerError)
		return
	}

	colors, err := s.repo.ListColors(ctx)
	if err != nil {
		http.Error(w, "failed to list colors", http.StatusInternalServerError)
		return
	}

	members, err := s.repo.ListMembers(ctx)
	if err != nil {
		http.Error(w, "failed to list members", http.StatusInternalServerError)
		return
	}

	// Populate images for each member
	for i := range members {
		imgs, err := s.repo.ListMemberImages(ctx, members[i].ID)
		if err == nil && len(imgs) > 0 {
			members[i].Images = imgs
		}
	}

	resp := model.BootstrapResponse{
		Groups:      groups,
		Colors:      colors,
		Members:     members,
		GeneratedAt: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if etag != "" {
		w.Header().Set("ETag", etag)
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/images/")
	if key == "" {
		http.NotFound(w, r)
		return
	}

	// 1. Check if local asset file exists
	localPath := filepath.Join(s.cfg.AssetDir, key)
	if _, err := os.Stat(localPath); err == nil {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeFile(w, r, localPath)
		return
	}

	// 2. Fallback: redirect to official CDN URL from image_sources.json
	if cdnURL, ok := s.imageSources[key]; ok {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.Redirect(w, r, cdnURL, http.StatusFound)
		return
	}

	http.NotFound(w, r)
}
