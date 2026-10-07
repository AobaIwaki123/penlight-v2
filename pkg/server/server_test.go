package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
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

func TestServer_ImageCacheProxyAndNoReferer(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Spin up a mock official CDN server to verify:
	// 1. Backend fetches image directly
	// 2. No Referer header is sent to the upstream CDN
	var receivedReferer string
	var upstreamHitCount int
	mockCDN := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHitCount++
		receivedReferer = r.Header.Get("Referer")
		w.Header().Set("Content-Type", "image/webp")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fake-webp-image-binary-data"))
	}))
	defer mockCDN.Close()

	// Redirect client requests to mock server via custom transport
	customClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			// Proxy request to mockCDN
			req, err := http.NewRequest(r.Method, mockCDN.URL, r.Body)
			if err != nil {
				return nil, err
			}
			req.Header = r.Header.Clone()
			return http.DefaultClient.Do(req)
		}),
	}
	srv.SetHTTPClient(customClient)

	// 1. Initial request: should fetch from upstream, return 200 OK directly (no 302 redirect), and cache locally
	key := "img_455e72f40ae75d5281943b482e2d2a95.webp"
	req := httptest.NewRequest(http.MethodGet, "/images/"+key, nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d (body: %s)", w.Code, w.Body.String())
	}
	if w.Header().Get("Location") != "" {
		t.Fatalf("expected no Location header (direct 200 OK, no 302), got %s", w.Header().Get("Location"))
	}
	if w.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatalf("expected immutable Cache-Control, got %s", w.Header().Get("Cache-Control"))
	}
	if w.Body.String() != "fake-webp-image-binary-data" {
		t.Fatalf("expected image body, got %q", w.Body.String())
	}

	// Verify upstream received NO Referer header (Ref: ADR-0033)
	if receivedReferer != "" {
		t.Fatalf("expected empty Referer header sent to upstream CDN, got %q", receivedReferer)
	}
	if upstreamHitCount != 1 {
		t.Fatalf("expected upstream to be hit 1 time, got %d", upstreamHitCount)
	}

	// 2. Second request: should hit local cache directly without contacting upstream CDN again
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/images/"+key, nil)
	srv.Handler().ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on cached request, got %d", w2.Code)
	}
	if upstreamHitCount != 1 {
		t.Fatalf("expected upstreamHitCount to remain 1 after second request (served from local cache), got %d", upstreamHitCount)
	}
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestServer_EmbeddedFrontendSPA(t *testing.T) {
	srv, _ := setupTestServer(t)

	// 1. Root / should serve index.html
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /, got %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected text/html for /, got %s", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), "<html") && !strings.Contains(w.Body.String(), "<!DOCTYPE") {
		t.Fatalf("expected HTML content for /, got: %s", w.Body.String())
	}

	// 2. Client-side route (SPA fallback) e.g. /quiz
	reqSPA := httptest.NewRequest(http.MethodGet, "/quiz", nil)
	wSPA := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wSPA, reqSPA)

	if wSPA.Code != http.StatusOK {
		t.Fatalf("expected 200 for SPA route /quiz, got %d", wSPA.Code)
	}
	if !strings.Contains(wSPA.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected text/html for /quiz fallback, got %s", wSPA.Header().Get("Content-Type"))
	}

	// 3. API route is not masked by SPA handler
	reqAPI := httptest.NewRequest(http.MethodGet, "/api/v1/sync/bootstrap", nil)
	wAPI := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wAPI, reqAPI)
	if wAPI.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/v1/sync/bootstrap, got %d", wAPI.Code)
	}
	if !strings.Contains(wAPI.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected application/json for API route, got %s", wAPI.Header().Get("Content-Type"))
	}
}
