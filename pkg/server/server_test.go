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

func TestServer_BatchAnswers(t *testing.T) {
	srv, _ := setupTestServer(t)

	// 1. Method Not Allowed for GET
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/quiz/answers/batch", nil)
	wGet := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 Method Not Allowed, got %d", wGet.Code)
	}

	// First fetch bootstrap to get valid member, song, and group IDs from seeds
	reqBS := httptest.NewRequest(http.MethodGet, "/api/v1/sync/bootstrap", nil)
	wBS := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wBS, reqBS)
	var bs model.BootstrapResponse
	if err := json.Unmarshal(wBS.Body.Bytes(), &bs); err != nil {
		t.Fatalf("failed to decode bootstrap: %v", err)
	}
	if len(bs.Members) == 0 || len(bs.Songs) == 0 {
		t.Fatal("expected seed members and songs to be available")
	}

	validMember := bs.Members[0]
	validSong := bs.Songs[0]

	// 2. Successful batch submission with polymorphic targets (Member & Song)
	memID := validMember.ID
	songID := validSong.ID
	batchPayload := model.BatchAnswerRequest{
		Answers: []model.BatchAnswerItem{
			{
				ID:             "ans_0192534a-9b41-715a-b9c1-111111111111",
				QuizQuestionID: "quiz_0192534a-9b41-715a-b9c1-222222222222",
				TargetMemberID: &memID,
				GroupID:        validMember.GroupID,
				IsCorrect:      true,
				ResponseTimeMs: 1200,
			},
			{
				ID:             "ans_0192534a-9b41-715a-b9c1-333333333333",
				QuizQuestionID: "quiz_0192534a-9b41-715a-b9c1-444444444444",
				TargetSongID:   &songID,
				GroupID:        validSong.GroupID,
				IsCorrect:      false,
				ResponseTimeMs: 2500,
			},
		},
	}

	jsonBytes, err := json.Marshal(batchPayload)
	if err != nil {
		t.Fatalf("failed to marshal batch request: %v", err)
	}

	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/quiz/answers/batch", strings.NewReader(string(jsonBytes)))
	reqPost.Header.Set("Content-Type", "application/json")
	wPost := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wPost, reqPost)

	if wPost.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for batch answers, got %d: %s", wPost.Code, wPost.Body.String())
	}

	var batchResp model.BatchAnswerResponse
	if err := json.Unmarshal(wPost.Body.Bytes(), &batchResp); err != nil {
		t.Fatalf("failed to unmarshal batch response: %v", err)
	}
	if batchResp.SyncedCount != 2 {
		t.Fatalf("expected SyncedCount=2, got %d", batchResp.SyncedCount)
	}
	if len(batchResp.SyncedIDs) != 2 {
		t.Fatalf("expected 2 SyncedIDs, got %d", len(batchResp.SyncedIDs))
	}

	// 3. Idempotent re-submission (INSERT OR IGNORE) should succeed without error
	reqIdempotent := httptest.NewRequest(http.MethodPost, "/api/v1/quiz/answers/batch", strings.NewReader(string(jsonBytes)))
	reqIdempotent.Header.Set("Content-Type", "application/json")
	wIdempotent := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wIdempotent, reqIdempotent)

	if wIdempotent.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for idempotent batch resubmission, got %d: %s", wIdempotent.Code, wIdempotent.Body.String())
	}

	// 4. Invalid Answer ID prefix
	invalidPayload := model.BatchAnswerRequest{
		Answers: []model.BatchAnswerItem{
			{
				ID:             "mem_invalid_prefix",
				QuizQuestionID: "quiz_0192534a-9b41-715a-b9c1-222222222222",
				TargetMemberID: &memID,
				GroupID:        validMember.GroupID,
			},
		},
	}
	invalidBytes, _ := json.Marshal(invalidPayload)
	reqInvalid := httptest.NewRequest(http.MethodPost, "/api/v1/quiz/answers/batch", strings.NewReader(string(invalidBytes)))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wInvalid, reqInvalid)

	if wInvalid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid prefix, got %d", wInvalid.Code)
	}
}

func TestServer_QuizStatistics(t *testing.T) {
	srv, _ := setupTestServer(t)

	// 1. Fetch valid members and songs from bootstrap
	reqBoot := httptest.NewRequest(http.MethodGet, "/api/v1/sync/bootstrap", nil)
	wBoot := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wBoot, reqBoot)
	var boot model.BootstrapResponse
	if err := json.Unmarshal(wBoot.Body.Bytes(), &boot); err != nil {
		t.Fatalf("failed to decode bootstrap: %v", err)
	}

	validMember := boot.Members[0]
	validSong := boot.Songs[0]

	// 2. Insert batch answers: 1 correct member, 1 wrong member, 1 correct song
	batchPayload := model.BatchAnswerRequest{
		Answers: []model.BatchAnswerItem{
			{
				ID:             "ans_0192534a-9b41-715a-b9c1-333333333331",
				QuizQuestionID: "quiz_0192534a-9b41-715a-b9c1-444444444441",
				TargetMemberID: &validMember.ID,
				GroupID:        validMember.GroupID,
				IsCorrect:      true,
				ResponseTimeMs: 1200,
			},
			{
				ID:             "ans_0192534a-9b41-715a-b9c1-333333333332",
				QuizQuestionID: "quiz_0192534a-9b41-715a-b9c1-444444444442",
				TargetMemberID: &validMember.ID,
				GroupID:        validMember.GroupID,
				IsCorrect:      false,
				ResponseTimeMs: 2400,
			},
			{
				ID:             "ans_0192534a-9b41-715a-b9c1-333333333333",
				QuizQuestionID: "quiz_0192534a-9b41-715a-b9c1-444444444443",
				TargetSongID:   &validSong.ID,
				GroupID:        validSong.GroupID,
				IsCorrect:      true,
				ResponseTimeMs: 1800,
			},
		},
	}
	jsonBytes, _ := json.Marshal(batchPayload)
	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/quiz/answers/batch", strings.NewReader(string(jsonBytes)))
	reqPost.Header.Set("Content-Type", "application/json")
	wPost := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from batch insert, got %d: %s", wPost.Code, wPost.Body.String())
	}

	// 3. GET /api/v1/quiz/statistics
	reqStats := httptest.NewRequest(http.MethodGet, "/api/v1/quiz/statistics", nil)
	wStats := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wStats, reqStats)

	if wStats.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from quiz statistics, got %d: %s", wStats.Code, wStats.Body.String())
	}

	var stats model.QuizStatisticsResponse
	if err := json.Unmarshal(wStats.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to unmarshal statistics: %v", err)
	}

	if stats.TotalAnswers != 3 {
		t.Fatalf("expected TotalAnswers=3, got %d", stats.TotalAnswers)
	}
	if stats.TotalCorrect != 2 {
		t.Fatalf("expected TotalCorrect=2, got %d", stats.TotalCorrect)
	}
	expectedAccuracy := 2.0 / 3.0
	if stats.AccuracyRate < expectedAccuracy-0.01 || stats.AccuracyRate > expectedAccuracy+0.01 {
		t.Fatalf("expected AccuracyRate ~ 0.667, got %f", stats.AccuracyRate)
	}
	if len(stats.Groups) == 0 {
		t.Fatal("expected at least 1 group stat, got empty")
	}
	if len(stats.WeakTargets) == 0 {
		t.Fatal("expected weak targets, got empty")
	}

	// Member target accuracy is 50% (1/2), Song target accuracy is 100% (1/1)
	// Weakest should be the member
	if stats.WeakTargets[0].TargetID != validMember.ID {
		t.Fatalf("expected weakest target to be member %s, got %s", validMember.ID, stats.WeakTargets[0].TargetID)
	}
	if stats.WeakTargets[0].AccuracyRate != 0.5 {
		t.Fatalf("expected member accuracy 0.5, got %f", stats.WeakTargets[0].AccuracyRate)
	}

	// 4. Test target_type filtering: target_type=song
	reqSongStats := httptest.NewRequest(http.MethodGet, "/api/v1/quiz/statistics?target_type=song", nil)
	wSongStats := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wSongStats, reqSongStats)
	if wSongStats.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wSongStats.Code)
	}
	var songStats model.QuizStatisticsResponse
	if err := json.Unmarshal(wSongStats.Body.Bytes(), &songStats); err != nil {
		t.Fatalf("failed to decode song statistics: %v", err)
	}
	for _, wt := range songStats.WeakTargets {
		if wt.TargetType != model.TargetTypeSong {
			t.Fatalf("expected target_type=song, got %s", wt.TargetType)
		}
	}

	// 5. Test invalid target_type returns 400
	reqInvalidTT := httptest.NewRequest(http.MethodGet, "/api/v1/quiz/statistics?target_type=invalid", nil)
	wInvalidTT := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wInvalidTT, reqInvalidTT)
	if wInvalidTT.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", wInvalidTT.Code)
	}
}

func TestServer_AdminMetadataOperations(t *testing.T) {
	srv, _ := setupTestServer(t)

	// Fetch a member from bootstrap endpoint to test with
	reqBootstrap := httptest.NewRequest(http.MethodGet, "/api/v1/sync/bootstrap", nil)
	wBootstrap := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wBootstrap, reqBootstrap)
	if wBootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap failed with %d", wBootstrap.Code)
	}
	var bootstrapData struct {
		Members []model.Member    `json:"members"`
		Colors  []model.Color     `json:"colors"`
		Photos  []model.PhotoType `json:"photo_types"`
	}
	if err := json.Unmarshal(wBootstrap.Body.Bytes(), &bootstrapData); err != nil {
		t.Fatalf("failed to decode bootstrap: %v", err)
	}
	if len(bootstrapData.Members) == 0 {
		t.Fatal("no members found in seed")
	}
	targetMember := bootstrapData.Members[0]

	// 1. GET /api/v1/admin/members/{id}
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/admin/members/"+string(targetMember.ID), nil)
	wGet := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wGet.Code)
	}
	var mGet model.Member
	if err := json.Unmarshal(wGet.Body.Bytes(), &mGet); err != nil {
		t.Fatalf("failed to decode member: %v", err)
	}
	if mGet.ID != targetMember.ID {
		t.Fatalf("expected ID=%s, got %s", targetMember.ID, mGet.ID)
	}

	// 2. POST /api/v1/admin/members/{id}/verify
	reqVerify := httptest.NewRequest(http.MethodPost, "/api/v1/admin/members/"+string(targetMember.ID)+"/verify", nil)
	wVerify := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wVerify, reqVerify)
	if wVerify.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wVerify.Code)
	}
	var mVerified model.Member
	if err := json.Unmarshal(wVerify.Body.Bytes(), &mVerified); err != nil {
		t.Fatalf("failed to decode verified member: %v", err)
	}
	if mVerified.VerifiedAt == nil {
		t.Fatal("expected VerifiedAt to be set after verify")
	}

	// 3. PATCH /api/v1/admin/members/{id}/penlight
	if len(bootstrapData.Colors) >= 2 {
		c1 := bootstrapData.Colors[0].ID
		c2 := bootstrapData.Colors[1].ID
		payload := `{"left_color_id":"` + string(c1) + `","right_color_id":"` + string(c2) + `","ordered":true}`
		reqPenlight := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/members/"+string(targetMember.ID)+"/penlight", strings.NewReader(payload))
		reqPenlight.Header.Set("Content-Type", "application/json")
		wPenlight := httptest.NewRecorder()
		srv.Handler().ServeHTTP(wPenlight, reqPenlight)
		if wPenlight.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", wPenlight.Code, wPenlight.Body.String())
		}
		var mPenlight model.Member
		if err := json.Unmarshal(wPenlight.Body.Bytes(), &mPenlight); err != nil {
			t.Fatalf("failed to decode updated penlight: %v", err)
		}
		if mPenlight.Penlight.LeftColorID != c1 || mPenlight.Penlight.RightColorID != c2 || !mPenlight.Penlight.Ordered {
			t.Fatalf("unexpected penlight update: %+v", mPenlight.Penlight)
		}
	}

	// 4. PATCH /api/v1/admin/members/{id}/status
	statusPayload := `{"status":"hiatus"}`
	reqStatus := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/members/"+string(targetMember.ID)+"/status", strings.NewReader(statusPayload))
	reqStatus.Header.Set("Content-Type", "application/json")
	wStatus := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wStatus, reqStatus)
	if wStatus.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", wStatus.Code, wStatus.Body.String())
	}
	var mStatus model.Member
	if err := json.Unmarshal(wStatus.Body.Bytes(), &mStatus); err != nil {
		t.Fatalf("failed to decode updated status: %v", err)
	}
	if mStatus.Status != model.StatusHiatus {
		t.Fatalf("expected hiatus, got %s", mStatus.Status)
	}

	// 5. PUT /api/v1/admin/members/{id}/images/primary
	if len(targetMember.Images) > 0 {
		imgID := targetMember.Images[0].ID
		imgPayload := `{"image_id":"` + string(imgID) + `"}`
		reqImg := httptest.NewRequest(http.MethodPut, "/api/v1/admin/members/"+string(targetMember.ID)+"/images/primary", strings.NewReader(imgPayload))
		reqImg.Header.Set("Content-Type", "application/json")
		wImg := httptest.NewRecorder()
		srv.Handler().ServeHTTP(wImg, reqImg)
		if wImg.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", wImg.Code, wImg.Body.String())
		}

		// 6. PATCH /api/v1/admin/images/{id}/photo-type
		if len(bootstrapData.Photos) > 0 {
			phtID := bootstrapData.Photos[0].ID
			ptPayload := `{"photo_type_id":"` + string(phtID) + `"}`
			reqPT := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/images/"+string(imgID)+"/photo-type", strings.NewReader(ptPayload))
			reqPT.Header.Set("Content-Type", "application/json")
			wPT := httptest.NewRecorder()
			srv.Handler().ServeHTTP(wPT, reqPT)
			if wPT.Code != http.StatusOK {
				t.Fatalf("expected 200 OK, got %d: %s", wPT.Code, wPT.Body.String())
			}
		}
	}

	// 7. Test CORS preflight allows PATCH and PUT
	reqOptions := httptest.NewRequest(http.MethodOptions, "/api/v1/admin/members/"+string(targetMember.ID)+"/penlight", nil)
	wOptions := httptest.NewRecorder()
	srv.Handler().ServeHTTP(wOptions, reqOptions)
	if wOptions.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", wOptions.Code)
	}
	allowMethods := wOptions.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowMethods, "PATCH") || !strings.Contains(allowMethods, "PUT") {
		t.Fatalf("expected PATCH and PUT in Access-Control-Allow-Methods, got %q", allowMethods)
	}
}


