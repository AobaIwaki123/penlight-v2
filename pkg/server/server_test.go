package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/aobaiwaki/penlight-v2/pkg/config"
	"github.com/aobaiwaki/penlight-v2/pkg/model"
	"github.com/aobaiwaki/penlight-v2/pkg/server"
)

func setupTestServer(t *testing.T) (*server.Server, string) {
	t.Helper()
	tempDir := t.TempDir()
	cfg := &config.Config{
		DataDir:       tempDir,
		DBPath:        filepath.Join(tempDir, "test.db"),
		AssetDir:      filepath.Join(tempDir, "assets"),
		SessionSecret: "test-secret-key-32bytes-secure-val!",
	}

	// Change working dir relative paths by symlinking migrations, seeds, and data if needed
	// Or Server runs from repo root where migrations/ and seeds/ are present
	srv, err := server.NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	return srv, tempDir
}

func TestServer_Healthz(t *testing.T) {
	srv, _ := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Fatalf("expected 'ok', got %q", w.Body.String())
	}
}

func TestServer_BootstrapAndETag(t *testing.T) {
	srv, _ := setupTestServer(t)

	// 1. Initial request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/bootstrap", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header, got empty")
	}

	var resp model.BootstrapResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal bootstrap response: %v", err)
	}

	if len(resp.Groups) == 0 {
		t.Fatal("expected groups to be populated from seeds")
	}
	if len(resp.Members) == 0 {
		t.Fatal("expected members to be populated from seeds")
	}

	// 2. Conditional request with If-None-Match
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/sync/bootstrap", nil)
	req2.Header.Set("If-None-Match", etag)
	w2 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotModified {
		t.Fatalf("expected 304 Not Modified, got %d", w2.Code)
	}
}

func TestServer_ImageRedirectFallback(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Test redirect for known image_key in seeds/data/image_sources.json
	req := httptest.NewRequest(http.MethodGet, "/images/img_455e72f40ae75d5281943b482e2d2a95.webp", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302 Found redirect, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc == "" {
		t.Fatal("expected Location header in redirect")
	}
}
