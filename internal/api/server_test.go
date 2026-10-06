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
	srv.OpenDataFolder = func() error {
		opened = true
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
}
