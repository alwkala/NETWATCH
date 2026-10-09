package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"netwatch/internal/api"
	"netwatch/internal/engine"
	"netwatch/internal/model"
	"netwatch/internal/netenv"
	"netwatch/internal/store"
)

type mockEnv struct {
	netenv.Portable
}

func (m *mockEnv) Info(ctx context.Context) (netenv.Info, error) {
	return netenv.Info{
		Adapters: []netenv.Adapter{
			{
				Name:    "Ethernet",
				Type:    "Ethernet",
				MAC:     "00:11:22:33:44:55",
				Up:      true,
				IP:      netip.MustParsePrefix("192.168.1.10/24"),
				Gateway: netip.MustParseAddr("192.168.1.1"),
			},
		},
	}, nil
}

func (m *mockEnv) Neighbors(ctx context.Context) ([]netenv.Neighbor, error) {
	return []netenv.Neighbor{}, nil
}

func (m *mockEnv) Ping(ctx context.Context, ip netip.Addr, timeout time.Duration) (time.Duration, bool) {
	return 5 * time.Millisecond, true
}

func (m *mockEnv) SendARP(ctx context.Context, ip netip.Addr) (string, error) {
	return "00:11:22:33:44:55", nil
}

func setupTestServer(t *testing.T) (*api.Server, string, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test_api.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("store.Open failed: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	eng := engine.New(engine.Options{
		Env:   &mockEnv{},
		Store: st,
	})

	token := "test-secret-bearer-token-12345"
	cfg := api.Config{
		Token:          token,
		AllowedOrigins: []string{"http://localhost:5173"},
		Version:        "0.1.0-test",
		DBPath:         dbPath,
	}
	srv := api.New(eng, cfg)
	return srv, token, dbPath
}

func TestAPI_HealthOpen(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on /v1/health without token, got %d", rec.Code)
	}

	var res struct {
		OK      bool   `json:"ok"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !res.OK || res.Version != "0.1.0-test" {
		t.Fatalf("unexpected health body: %+v", res)
	}
}

func TestAPI_SecurityBoundaries(t *testing.T) {
	srv, token, _ := setupTestServer(t)
	handler := srv.Handler()

	// 1. Missing Token on protected endpoint
	t.Run("missing_token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/network", nil)
		req.Host = "127.0.0.1:8080"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	// 2. Invalid Token on protected endpoint
	t.Run("invalid_token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/network", nil)
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Authorization", "Bearer wrong-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
		}
	})

	// 3. Spoofed Host Header (DNS-Rebinding protection)
	t.Run("host_protection", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/network", nil)
		req.Host = "attacker.example.com:8080"
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for external Host, got %d", rec.Code)
		}
	})

	// 4. Disallowed Origin Header (Cross-Site Request protection)
	t.Run("origin_protection_disallowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/network", nil)
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Origin", "http://malicious-website.com")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for disallowed Origin, got %d", rec.Code)
		}
	})

	// 5. Allowed Origin Header (Wails or configured dev server)
	t.Run("origin_protection_allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/network", nil)
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for allowed Origin, got %d: %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
			t.Fatalf("CORS headers missing or incorrect: %+v", rec.Header())
		}
	})
}

func TestAPI_SettingsAndDatabaseEndpoints(t *testing.T) {
	srv, token, dbPath := setupTestServer(t)
	opened := false
	openedURL := ""
	srv.OpenDataFolder = func() error {
		opened = true
		return nil
	}
	srv.OpenURL = func(targetURL string) error {
		openedURL = targetURL
		return nil
	}
	handler := srv.Handler()

	doReq := func(method, path string, body []byte) *httptest.ResponseRecorder {
		var r *http.Request
		if body != nil {
			r = httptest.NewRequest(method, path, bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
		} else {
			r = httptest.NewRequest(method, path, nil)
		}
		r.Host = "127.0.0.1:8080"
		r.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		return rec
	}

	// 1. GET /v1/settings returns defaults
	rec := doReq(http.MethodGet, "/v1/settings", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/settings failed: %d", rec.Code)
	}
	var st model.Settings
	if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if st.ScanInterval != "5m" {
		t.Fatalf("expected default scanInterval 5m, got %s", st.ScanInterval)
	}

	// 2. PUT /v1/settings updates preferences
	st.ScanInterval = "15m"
	st.NotifyNewDevice = false
	b, _ := json.Marshal(st)
	rec = doReq(http.MethodPut, "/v1/settings", b)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /v1/settings failed: %d", rec.Code)
	}
	var updated model.Settings
	_ = json.Unmarshal(rec.Body.Bytes(), &updated)
	if updated.ScanInterval != "15m" || updated.NotifyNewDevice {
		t.Fatalf("settings not updated: %+v", updated)
	}

	// 3. GET /v1/data/stats
	rec = doReq(http.MethodGet, "/v1/data/stats", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/data/stats failed: %d", rec.Code)
	}
	var stats model.DatabaseStats
	_ = json.Unmarshal(rec.Body.Bytes(), &stats)
	if stats.DBPath != dbPath || stats.FileSizeBytes == 0 {
		t.Fatalf("unexpected data stats: %+v", stats)
	}

	// 4. POST /v1/data/vacuum
	rec = doReq(http.MethodPost, "/v1/data/vacuum", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /v1/data/vacuum failed: %d", rec.Code)
	}
	var mRes model.MaintenanceResult
	_ = json.Unmarshal(rec.Body.Bytes(), &mRes)
	if !mRes.Success {
		t.Fatalf("vacuum failed: %+v", mRes)
	}

	// 5. POST /v1/data/integrity
	rec = doReq(http.MethodPost, "/v1/data/integrity", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /v1/data/integrity failed: %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &mRes)
	if !mRes.Success || !strings.Contains(mRes.Message, "ok") {
		t.Fatalf("integrity check failed: %+v", mRes)
	}

	// 6. POST /v1/data/open calls hook
	rec = doReq(http.MethodPost, "/v1/data/open", nil)
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("POST /v1/data/open failed: %d", rec.Code)
	}
	if !opened {
		t.Fatalf("OpenDataFolder hook was not invoked")
	}

	// 7. DELETE /v1/data clears history
	rec = doReq(http.MethodDelete, "/v1/data", nil)
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("DELETE /v1/data failed: %d", rec.Code)
	}

	// 8. Invalid PUT /v1/settings returns 400 Bad Request
	st.ScanInterval = "invalid_interval_value"
	badJSON, _ := json.Marshal(st)
	rec = doReq(http.MethodPut, "/v1/settings", badJSON)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid scanInterval, got %d", rec.Code)
	}

	// 9. POST /v1/system/open
	rec = doReq(http.MethodPost, "/v1/system/open", []byte(`{"url":"http://192.168.1.1"}`))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for valid private LAN URL, got %d", rec.Code)
	}
	if openedURL != "http://192.168.1.1" {
		t.Fatalf("expected openedURL to be http://192.168.1.1, got %s", openedURL)
	}

	// 10. POST /v1/system/open with public IP rejected
	rec = doReq(http.MethodPost, "/v1/system/open", []byte(`{"url":"http://8.8.8.8"}`))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for public IP URL, got %d", rec.Code)
	}

	// 11. POST /v1/system/open with invalid scheme rejected
	rec = doReq(http.MethodPost, "/v1/system/open", []byte(`{"url":"ftp://192.168.1.1"}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-http scheme, got %d", rec.Code)
	}
}

func TestAPI_ScanAndSSEStream(t *testing.T) {
	srv, token, _ := setupTestServer(t)
	handler := srv.Handler()

	// 1. Start a scan
	startReq := httptest.NewRequest(http.MethodPost, "/v1/scans", bytes.NewReader([]byte(`{"type":"quick"}`)))
	startReq.Header.Set("Content-Type", "application/json")
	startReq.Header.Set("Authorization", "Bearer "+token)
	startReq.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, startReq)

	if rec.Code != http.StatusAccepted && rec.Code != http.StatusOK {
		t.Fatalf("POST /v1/scans failed: %d body=%s", rec.Code, rec.Body.String())
	}
	var startRes struct {
		ScanID string `json:"scanId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &startRes); err != nil || startRes.ScanID == "" {
		t.Fatalf("unexpected start scan response: %s", rec.Body.String())
	}

	// 2. Query scan snapshot
	snapReq := httptest.NewRequest(http.MethodGet, "/v1/scans/"+startRes.ScanID, nil)
	snapReq.Header.Set("Authorization", "Bearer "+token)
	snapReq.Host = "127.0.0.1:8080"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, snapReq)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/scans/{id} failed: %d", rec.Code)
	}

	// 3. Connect to SSE stream
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	streamReq := httptest.NewRequest(http.MethodGet, "/v1/scans/"+startRes.ScanID+"/stream", nil).WithContext(ctx)
	streamReq.Header.Set("Authorization", "Bearer "+token)
	streamReq.Host = "127.0.0.1:8080"
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, streamReq)

	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("expected text/event-stream content type, got %s", ct)
	}
	if body := rec.Body.String(); !strings.Contains(body, "event: ") {
		t.Fatalf("expected SSE events in body, got: %s", body)
	}

	// 4. Connect to SSE stream using ?token= query param (EventSource compatibility)
	streamQueryReq := httptest.NewRequest(http.MethodGet, "/v1/scans/"+startRes.ScanID+"/stream?token="+token, nil).WithContext(ctx)
	streamQueryReq.Host = "127.0.0.1:8080"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, streamQueryReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for SSE stream with valid ?token= query param, got %d", rec.Code)
	}

	// 5. Connect to SSE stream with invalid ?token=
	streamWrongTokenReq := httptest.NewRequest(http.MethodGet, "/v1/scans/"+startRes.ScanID+"/stream?token=invalid-secret", nil).WithContext(ctx)
	streamWrongTokenReq.Host = "127.0.0.1:8080"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, streamWrongTokenReq)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for SSE stream with invalid ?token=, got %d", rec.Code)
	}

	// 6. Non-stream REST endpoint must reject ?token= query param (strictly enforce Authorization header)
	restQueryReq := httptest.NewRequest(http.MethodGet, "/v1/devices?token="+token, nil)
	restQueryReq.Host = "127.0.0.1:8080"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, restQueryReq)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for REST endpoint with ?token= query param, got %d", rec.Code)
	}
}

func TestAPI_PruneData(t *testing.T) {
	srv, token, _ := setupTestServer(t)
	handler := srv.Handler()

	body := bytes.NewReader([]byte(`{"olderThanDays": 30}`))
	req := httptest.NewRequest(http.MethodPost, "/v1/data/prune", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /v1/data/prune failed: %d, body=%s", rec.Code, rec.Body.String())
	}
	var res struct {
		DeletedCount  int64 `json:"deletedCount"`
		OlderThanDays int   `json:"olderThanDays"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal prune response: %v", err)
	}
	if res.OlderThanDays != 30 {
		t.Fatalf("expected olderThanDays=30, got %d", res.OlderThanDays)
	}
}

func TestAPI_CORSPreflightIncludesAllMethods(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodOptions, "/v1/settings", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for preflight OPTIONS, got %d", rec.Code)
	}
	methods := rec.Header().Get("Access-Control-Allow-Methods")
	for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
		if !strings.Contains(methods, m) {
			t.Errorf("expected CORS Allow-Methods to contain %q, got: %s", m, methods)
		}
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected nosniff header on OPTIONS response")
	}
}

func TestAPI_StrictBearerAuth(t *testing.T) {
	srv, token, _ := setupTestServer(t)
	handler := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/network", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for raw token without 'Bearer ', got %d", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected nosniff header on 401 response")
	}
}

func TestAPI_ParseOriginLoopbackValidation(t *testing.T) {
	valid := []string{
		"http://localhost:3000",
		"http://127.0.0.1:5173",
		"http://app.localhost",
	}
	for _, v := range valid {
		if _, err := api.ParseOrigin(v); err != nil {
			t.Errorf("expected %q to be valid, got: %v", v, err)
		}
	}

	invalid := []string{
		"https://evil.example.com",
		"http://192.168.1.50:3000",
		"not-a-url",
	}
	for _, inv := range invalid {
		if _, err := api.ParseOrigin(inv); err == nil {
			t.Errorf("expected %q to fail loopback check", inv)
		}
	}
}

func TestAPI_DeviceMergeAndTrustStatus(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_merge.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now().UTC()
	ws := store.Writeset{
		Devices: []store.Known{
			{Device: model.Device{ID: "d-target", MAC: "00:11:22:33:44:55", IP: "192.168.1.50", Name: "Laptop", TrustStatus: model.TrustKnown, FirstSeen: now, LastSeen: now}},
			{Device: model.Device{ID: "d-source", MAC: "02:11:22:33:44:55", IP: "192.168.1.51", Name: "Laptop Private", IsRandomizedMAC: true, TrustStatus: model.TrustUnknown, FirstSeen: now, LastSeen: now}},
		},
	}
	if err := st.Commit(context.Background(), ws); err != nil {
		t.Fatal(err)
	}

	eng := engine.New(engine.Options{
		Env:   &mockEnv{},
		Store: st,
	})
	token := "test-secret-bearer-token-12345"
	srv := api.New(eng, api.Config{
		Token:   token,
		Version: "0.1.0-test",
		DBPath:  dbPath,
	})
	handler := srv.Handler()

	doReq := func(method, path string, body []byte) *httptest.ResponseRecorder {
		var r *http.Request
		if body != nil {
			r = httptest.NewRequest(method, path, bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
		} else {
			r = httptest.NewRequest(method, path, nil)
		}
		r.Host = "127.0.0.1:8080"
		r.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		return rec
	}

	// 1. PATCH trust status to guest
	patchBody, _ := json.Marshal(map[string]any{"trustStatus": "guest"})
	rec := doReq(http.MethodPatch, "/v1/devices/d-target", patchBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH trustStatus failed: %d, %s", rec.Code, rec.Body.String())
	}
	var patched model.Device
	_ = json.Unmarshal(rec.Body.Bytes(), &patched)
	if patched.TrustStatus != "guest" {
		t.Fatalf("expected trustStatus 'guest', got %s", patched.TrustStatus)
	}

	// 2. Reject invalid trust status
	badPatch, _ := json.Marshal(map[string]any{"trustStatus": "invalid-status"})
	rec = doReq(http.MethodPatch, "/v1/devices/d-target", badPatch)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad trust status, got %d", rec.Code)
	}

	// 3. POST /v1/devices/{id}/merge
	mergeBody, _ := json.Marshal(map[string]any{"sourceId": "d-source"})
	rec = doReq(http.MethodPost, "/v1/devices/d-target/merge", mergeBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST merge failed: %d, %s", rec.Code, rec.Body.String())
	}

	// Verify source device is now gone
	rec = doReq(http.MethodGet, "/v1/devices/d-source", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for merged source device, got %d", rec.Code)
	}
}
