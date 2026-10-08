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

func TestServer_MetadataEditProposalWorkflow(t *testing.T) {
	srv, _ := setupTestServer(t)

	requestBootstrap := func(path string, etag string) (*httptest.ResponseRecorder, model.BootstrapResponse) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if etag != "" {
			req.Header.Set("If-None-Match", etag)
		}
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)
		var response model.BootstrapResponse
		if w.Code == http.StatusOK {
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to decode bootstrap: %v", err)
			}
		}
		return w, response
	}

	initial, bootstrap := requestBootstrap("/api/v1/sync/bootstrap", "")
	if initial.Code != http.StatusOK || len(bootstrap.Members) == 0 || len(bootstrap.PhotoTypes) == 0 {
		t.Fatalf("bootstrap is incomplete: status=%d members=%d photo_types=%d", initial.Code, len(bootstrap.Members), len(bootstrap.PhotoTypes))
	}
	graduated, graduatedBootstrap := requestBootstrap("/api/v1/sync/bootstrap?include_graduated=true", "")
	if graduated.Code != http.StatusOK || graduated.Header().Get("ETag") == initial.Header().Get("ETag") || len(graduatedBootstrap.Members) <= len(bootstrap.Members) {
		t.Fatalf("include_graduated must have its own ETag: default=%q graduated=%q", initial.Header().Get("ETag"), graduated.Header().Get("ETag"))
	}
	conditional, _ := requestBootstrap("/api/v1/sync/bootstrap", initial.Header().Get("ETag"))
	if conditional.Code != http.StatusNotModified {
		t.Fatalf("expected conditional bootstrap to return 304, got %d", conditional.Code)
	}

	target := bootstrap.Members[0]
	updatedPenlight := target.Penlight
	updatedPenlight.Ordered = !updatedPenlight.Ordered
	proposalID := model.ID("prp_test_server_workflow")
	request := model.SubmitMetadataEditProposalRequest{
		ID:           proposalID,
		BaseRevision: target.MetadataRevision,
		Changes: model.MetadataEditChanges{
			Penlight: &model.PenlightChange{Before: target.Penlight, After: updatedPenlight},
		},
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	postProposal := func(payload []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/members/"+string(target.ID)+"/metadata-edit-proposals", strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)
		return w
	}

	submitted := postProposal(requestJSON)
	if submitted.Code != http.StatusOK {
		t.Fatalf("expected proposal submission 200, got %d: %s", submitted.Code, submitted.Body.String())
	}
	var proposal model.MetadataEditProposal
	if err := json.Unmarshal(submitted.Body.Bytes(), &proposal); err != nil {
		t.Fatal(err)
	}
	if proposal.Status != model.ProposalPending || proposal.ID != proposalID {
		t.Fatalf("unexpected submitted proposal: %+v", proposal)
	}

	// A pending proposal is visible to the admin queue but does not alter the public member or ETag.
	pendingList := httptest.NewRecorder()
	srv.Handler().ServeHTTP(pendingList, httptest.NewRequest(http.MethodGet, "/api/v1/admin/metadata-edit-proposals", nil))
	if pendingList.Code != http.StatusOK || !strings.Contains(pendingList.Body.String(), string(proposalID)) {
		t.Fatalf("pending proposal was not listed: status=%d body=%s", pendingList.Code, pendingList.Body.String())
	}
	unchanged, _ := requestBootstrap("/api/v1/sync/bootstrap", initial.Header().Get("ETag"))
	if unchanged.Code != http.StatusNotModified {
		t.Fatalf("pending proposal changed bootstrap ETag: status=%d", unchanged.Code)
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/metadata-edit-proposals/"+string(proposalID)+"/approve", nil)
	approvedResponse := httptest.NewRecorder()
	srv.Handler().ServeHTTP(approvedResponse, approveReq)
	if approvedResponse.Code != http.StatusOK {
		t.Fatalf("expected approval 200, got %d: %s", approvedResponse.Code, approvedResponse.Body.String())
	}
	if err := json.Unmarshal(approvedResponse.Body.Bytes(), &proposal); err != nil {
		t.Fatal(err)
	}
	if proposal.Status != model.ProposalApproved || proposal.ApprovedAt == nil {
		t.Fatalf("unexpected approved proposal: %+v", proposal)
	}

	afterApproval, approvedBootstrap := requestBootstrap("/api/v1/sync/bootstrap", "")
	if afterApproval.Code != http.StatusOK || afterApproval.Header().Get("ETag") == initial.Header().Get("ETag") {
		t.Fatalf("approval did not update bootstrap ETag: before=%q after=%q", initial.Header().Get("ETag"), afterApproval.Header().Get("ETag"))
	}
	var approvedMember *model.Member
	for i := range approvedBootstrap.Members {
		if approvedBootstrap.Members[i].ID == target.ID {
			approvedMember = &approvedBootstrap.Members[i]
			break
		}
	}
	if approvedMember == nil || approvedMember.Penlight.Ordered != updatedPenlight.Ordered || approvedMember.MetadataRevision != target.MetadataRevision+1 || approvedMember.VerifiedAt == nil {
		t.Fatalf("approval did not update public member: %+v", approvedMember)
	}

	// Re-approval and same-ID resubmission are idempotent.
	approvedAgain := httptest.NewRecorder()
	srv.Handler().ServeHTTP(approvedAgain, httptest.NewRequest(http.MethodPost, "/api/v1/admin/metadata-edit-proposals/"+string(proposalID)+"/approve", nil))
	if approvedAgain.Code != http.StatusOK || approvedAgain.Body.String() == "" {
		t.Fatalf("re-approval was not idempotent: status=%d", approvedAgain.Code)
	}
	resubmitted := postProposal(requestJSON)
	if resubmitted.Code != http.StatusOK || !strings.Contains(resubmitted.Body.String(), `"status":"approved"`) {
		t.Fatalf("same-ID resubmission was not idempotent: status=%d body=%s", resubmitted.Code, resubmitted.Body.String())
	}

	// A different payload under the same ID is rejected.
	request.Changes.Penlight.After.Ordered = !request.Changes.Penlight.After.Ordered
	differentJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	different := postProposal(differentJSON)
	if different.Code != http.StatusBadRequest {
		t.Fatalf("expected same-ID content mismatch 400, got %d", different.Code)
	}

	// Rejection records the decision without changing the public revision or ETag.
	rejectID := model.ID("prp_test_server_rejection")
	rejectRequest := model.SubmitMetadataEditProposalRequest{
		ID:           rejectID,
		BaseRevision: approvedMember.MetadataRevision,
		Changes: model.MetadataEditChanges{
			Generation: &model.GenerationChange{Before: approvedMember.Generation, After: approvedMember.Generation + 1},
		},
	}
	rejectJSON, err := json.Marshal(rejectRequest)
	if err != nil {
		t.Fatal(err)
	}
	if rejectedSubmission := postProposal(rejectJSON); rejectedSubmission.Code != http.StatusOK {
		t.Fatalf("expected second proposal submission 200, got %d: %s", rejectedSubmission.Code, rejectedSubmission.Body.String())
	}
	reason := `{"reason":"not enough evidence"}`
	rejectHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/admin/metadata-edit-proposals/"+string(rejectID)+"/reject", strings.NewReader(reason))
	rejectHTTP.Header.Set("Content-Type", "application/json")
	rejectedResponse := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rejectedResponse, rejectHTTP)
	if rejectedResponse.Code != http.StatusOK || !strings.Contains(rejectedResponse.Body.String(), `"status":"rejected"`) {
		t.Fatalf("expected rejection 200, got %d: %s", rejectedResponse.Code, rejectedResponse.Body.String())
	}
	emptyReject := httptest.NewRecorder()
	srv.Handler().ServeHTTP(emptyReject, httptest.NewRequest(http.MethodPost, "/api/v1/admin/metadata-edit-proposals/"+string(rejectID)+"/reject", nil))
	if emptyReject.Code != http.StatusOK {
		t.Fatalf("empty rejection body should be accepted: status=%d body=%s", emptyReject.Code, emptyReject.Body.String())
	}
	unchangedAfterReject, _ := requestBootstrap("/api/v1/sync/bootstrap", afterApproval.Header().Get("ETag"))
	if unchangedAfterReject.Code != http.StatusNotModified {
		t.Fatalf("rejection changed bootstrap ETag: status=%d", unchangedAfterReject.Code)
	}
	invalidGraduated, _ := requestBootstrap("/api/v1/sync/bootstrap?include_graduated=1", "")
	if invalidGraduated.Code != http.StatusBadRequest {
		t.Fatalf("invalid include_graduated should return 400, got %d", invalidGraduated.Code)
	}
}
