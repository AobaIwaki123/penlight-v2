package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aobaiwaki/penlight-v2/frontend"
	"github.com/aobaiwaki/penlight-v2/migrations"
	"github.com/aobaiwaki/penlight-v2/pkg/config"
	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/repository"
	"github.com/aobaiwaki/penlight-v2/seeds"
)

// Server holds server dependencies and HTTP router.
type Server struct {
	cfg          *config.Config
	repo         *repository.SQLiteRepository
	imageMu      sync.RWMutex
	imageSources map[string]string // image_key -> official CDN URL
	httpClient   *http.Client
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

	// 1. Auto-apply migrations (embedded in binary, fallback to disk)
	if err := applyMigrations(repo.DB()); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// 2. Check if seed data needs to be populated or synchronized (Ref: ADR-0021)
	var currentVersion string
	_ = repo.DB().QueryRow("SELECT version FROM master_versions WHERE id = 'current';").Scan(&currentVersion)
	if currentVersion != model.CurrentMasterVersion {
		log.Printf("Master data update detected (current: %q, target: %q), syncing seeds/seed.sql...", currentVersion, model.CurrentMasterVersion)
		seedSQL, err := seeds.FS.ReadFile("seed.sql")
		if err != nil {
			seedSQL, err = resolveFile("seeds/seed.sql")
		}
		if err != nil {
			log.Printf("Warning: failed to read seeds/seed.sql: %v", err)
		} else {
			if _, err := repo.DB().Exec(string(seedSQL)); err != nil {
				log.Printf("Warning: failed to execute seeds/seed.sql: %v", err)
			} else {
				log.Println("Seed data successfully synced into SQLite.")
			}
		}
	}

	// 3. Load image sources map for dynamic CDN redirect fallback
	imageSources := make(map[string]string)
	srcBytes, err := seeds.FS.ReadFile("data/image_sources.json")
	if err != nil {
		for _, candidate := range []string{"seeds/data/image_sources.json", "data/image_sources.json"} {
			if b, rErr := resolveFile(candidate); rErr == nil {
				srcBytes = b
				break
			}
		}
	}
	if len(srcBytes) > 0 {
		var entries []imageSourceEntry
		if err := json.Unmarshal(srcBytes, &entries); err == nil {
			for _, e := range entries {
				if e.ImageKey != "" && e.URL != "" {
					imageSources[e.ImageKey] = e.URL
				}
			}
		}
	}

	s := &Server{
		cfg:          cfg,
		repo:         repo,
		imageSources: imageSources,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		mux:          http.NewServeMux(),
	}
	s.registerRoutes()
	return s, nil
}

// SetHTTPClient sets a custom HTTP client (primarily for testing mock CDN endpoints).
func (s *Server) SetHTTPClient(client *http.Client) {
	s.httpClient = client
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
	s.mux.HandleFunc("/api/v1/quiz/answers/batch", s.handleBatchAnswers)
	s.mux.HandleFunc("/images/", s.handleImage)

	// Embedded frontend static SPA handler (Ref: ADR-0002, ADR-0011)
	if assets, err := frontend.Assets(); err == nil {
		s.mux.Handle("/", s.spaHandler(assets))
	} else {
		log.Printf("Warning: failed to initialize embedded frontend assets: %v", err)
	}
}

func (s *Server) spaHandler(assets fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(assets))

	return func(w http.ResponseWriter, r *http.Request) {
		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		// 1. Try serving requested static file from embedded assets
		if f, err := assets.Open(cleanPath); err == nil {
			stat, err := f.Stat()
			_ = f.Close()
			if err == nil && !stat.IsDir() {
				if strings.HasPrefix(cleanPath, "_next/static/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "public, max-age=3600")
				}
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// 2. SPA Fallback: serve index.html for client-side routing
		indexFile, err := assets.Open("index.html")
		if err != nil {
			// Fallback when frontend static export is not yet built (e.g. clean checkout before npm run build)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<!DOCTYPE html><html><head><meta charset=\"UTF-8\"><title>Penlight v2</title></head><body><div id=\"root\">Penlight v2 (Frontend not built yet. Run npm run build)</div></body></html>"))
			return
		}
		defer indexFile.Close()

		indexBytes, err := io.ReadAll(indexFile)
		if err != nil {
			http.Error(w, "failed to read index.html", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexBytes)
	}
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

	series, err := s.repo.ListSeries(ctx)
	if err != nil {
		http.Error(w, "failed to list series", http.StatusInternalServerError)
		return
	}
	if series == nil {
		series = make([]model.Series, 0)
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

	songs, err := s.repo.ListSongs(ctx)
	if err != nil {
		http.Error(w, "failed to list songs", http.StatusInternalServerError)
		return
	}
	if songs == nil {
		songs = make([]model.Song, 0)
	}

	resp := model.BootstrapResponse{
		Series:      series,
		Groups:      groups,
		Colors:      colors,
		Members:     members,
		Songs:       songs,
		GeneratedAt: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if etag != "" {
		w.Header().Set("ETag", etag)
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func writeProblemDetails(w http.ResponseWriter, status int, code, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.AppError{
		Status: status,
		Code:   code,
		Title:  title,
		Detail: detail,
	})
}

func (s *Server) handleBatchAnswers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeProblemDetails(w, http.StatusMethodNotAllowed, model.CodeInvalidParams, "Method Not Allowed", "POST is required")
		return
	}

	var req model.BatchAnswerRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024*1024)).Decode(&req); err != nil {
		writeProblemDetails(w, http.StatusBadRequest, model.CodeInvalidParams, "Invalid JSON", err.Error())
		return
	}

	if len(req.Answers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.BatchAnswerResponse{
			SyncedCount: 0,
			SyncedIDs:   []model.ID{},
		})
		return
	}

	logs := make([]model.AnswerLog, 0, len(req.Answers))
	syncedIDs := make([]model.ID, 0, len(req.Answers))

	for _, item := range req.Answers {
		if !strings.HasPrefix(string(item.ID), string(model.PrefixAnswer)+"_") {
			writeProblemDetails(w, http.StatusBadRequest, model.CodeInvalidParams, "Invalid Answer ID", "ID must have prefix "+string(model.PrefixAnswer)+"_")
			return
		}
		if (item.TargetMemberID == nil && item.TargetSongID == nil) || (item.TargetMemberID != nil && item.TargetSongID != nil) {
			writeProblemDetails(w, http.StatusBadRequest, model.CodeInvalidParams, "Invalid Target", "Exactly one of target_member_id or target_song_id must be provided")
			return
		}

		answeredAt := item.AnsweredAt
		if answeredAt.IsZero() {
			answeredAt = time.Now().UTC()
		}

		logs = append(logs, model.AnswerLog{
			ID:             item.ID,
			UserID:         item.UserID,
			QuizQuestionID: item.QuizQuestionID,
			TargetMemberID: item.TargetMemberID,
			TargetSongID:   item.TargetSongID,
			GroupID:        item.GroupID,
			IsCorrect:      item.IsCorrect,
			ResponseTimeMs: item.ResponseTimeMs,
			AnsweredAt:     answeredAt,
		})
		syncedIDs = append(syncedIDs, item.ID)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.repo.BatchInsertAnswerLogs(ctx, logs); err != nil {
		writeProblemDetails(w, http.StatusInternalServerError, model.CodeInternalError, "Database Error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.BatchAnswerResponse{
		SyncedCount: len(logs),
		SyncedIDs:   syncedIDs,
	})
}

func (s *Server) getImageSourceURL(key string) (string, bool) {
	s.imageMu.RLock()
	url, ok := s.imageSources[key]
	s.imageMu.RUnlock()
	if ok {
		return url, true
	}

	// Dynamic reload from disk if missing in memory (e.g. seed data added while running)
	s.imageMu.Lock()
	defer s.imageMu.Unlock()
	if url, ok := s.imageSources[key]; ok {
		return url, true
	}

	for _, candidate := range []string{"seeds/data/image_sources.json", "data/image_sources.json"} {
		if srcBytes, err := resolveFile(candidate); err == nil {
			var entries []imageSourceEntry
			if err := json.Unmarshal(srcBytes, &entries); err == nil {
				for _, e := range entries {
					if e.ImageKey != "" && e.URL != "" {
						s.imageSources[e.ImageKey] = e.URL
					}
				}
				if len(s.imageSources) > 0 {
					break
				}
			}
		}
	}
	url, ok = s.imageSources[key]
	return url, ok
}

func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/images/")
	if key == "" || strings.Contains(key, "..") || strings.Contains(key, "/") {
		http.NotFound(w, r)
		return
	}

	// 1. Check if local asset file already exists
	localPath := filepath.Join(s.cfg.AssetDir, key)
	if _, err := os.Stat(localPath); err == nil {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeFile(w, r, localPath)
		return
	}

	// 2. On-demand cache-through proxy: fetch from official CDN URL without Referer (Ref: ADR-0033)
	cdnURL, ok := s.getImageSourceURL(key)
	if !ok {
		http.NotFound(w, r)
		return
	}

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	// Create request with context and ensure no Referer is sent
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, cdnURL, nil)
	if err != nil {
		http.Error(w, "failed to create upstream request", http.StatusInternalServerError)
		return
	}
	req.Header.Del("Referer")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "upstream image fetch error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("upstream returned %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read upstream image", http.StatusBadGateway)
		return
	}

	// 3. Atomically save to local storage (AssetDir) for subsequent 0ms requests
	if err := os.MkdirAll(s.cfg.AssetDir, 0o755); err == nil {
		tmpFile, err := os.CreateTemp(s.cfg.AssetDir, "img-tmp-*")
		if err == nil {
			tmpPath := tmpFile.Name()
			if _, err := tmpFile.Write(imgBytes); err == nil {
				_ = tmpFile.Close()
				_ = os.Rename(tmpPath, localPath)
			} else {
				_ = tmpFile.Close()
				_ = os.Remove(tmpPath)
			}
		}
	}

	// 4. Serve image directly from backend with immutable cache headers
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		if strings.HasSuffix(key, ".webp") {
			contentType = "image/webp"
		} else if strings.HasSuffix(key, ".jpg") || strings.HasSuffix(key, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(key, ".png") {
			contentType = "image/png"
		} else {
			contentType = "application/octet-stream"
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(imgBytes)
}

func applyMigrations(db *sql.DB) error {
	// Create schema_migrations table if not exists
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL
	);`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// If groups table already exists from pre-schema_migrations era, ensure 000001 is recorded
	var groupsExists int
	_ = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='groups';").Scan(&groupsExists)
	if groupsExists > 0 {
		_, _ = db.Exec("INSERT OR IGNORE INTO schema_migrations (version, applied_at) VALUES ('000001_init.up.sql', ?)", time.Now().UTC().Format(time.RFC3339))
	}

	migrationFiles := []string{
		"000001_init.up.sql",
		"000002_add_series_and_songs.up.sql",
	}

	for _, filename := range migrationFiles {
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", filename).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", filename, err)
		}
		if exists > 0 {
			continue
		}

		migrationSQL, err := migrations.FS.ReadFile(filename)
		if err != nil {
			migrationSQL, err = resolveFile("migrations/" + filename)
			if err != nil {
				return fmt.Errorf("failed to read migration SQL (%s): %w", filename, err)
			}
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin tx for migration %s: %w", filename, err)
		}

		if _, err := tx.Exec(string(migrationSQL)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		now := time.Now().UTC().Format(time.RFC3339)
		if _, err := tx.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", filename, now); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}
		log.Printf("Applied migration: %s", filename)
	}

	return nil
}
