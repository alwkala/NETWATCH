// Package api exposes the engine to the UI over HTTP on the loopback
// interface. Three independent checks protect it from other software on the
// machine and from web pages open in a browser:
//
//  1. it only listens on 127.0.0.1;
//  2. every request must carry the per-session bearer token;
//  3. the Host and Origin headers must match this server / the app shell
//     (blocks DNS-rebinding and cross-site requests).
package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"netwatch/internal/autostart"
	"netwatch/internal/engine"
	"netwatch/internal/model"
	"netwatch/internal/store"
)

// DefaultOrigins are the origins the Wails WebView2 shell uses.
var DefaultOrigins = []string{"http://wails.localhost", "https://wails.localhost", "wails://wails"}

type Config struct {
	Token          string
	AllowedOrigins []string // exact matches; DefaultOrigins are always included
	Version        string
	DBPath         string
	Logger         *slog.Logger
	Autostart      autostart.Manager
}

type Server struct {
	eng     *engine.Engine
	cfg     Config
	origins map[string]bool
	port    string
	log     *slog.Logger
	// hooks supplied by the host application
	OpenDataFolder func() error
}

func New(eng *engine.Engine, cfg Config) *Server {
	if cfg.Autostart == nil {
		cfg.Autostart = autostart.New()
	}
	s := &Server{eng: eng, cfg: cfg, origins: map[string]bool{}, log: cfg.Logger}
	if s.log == nil {
		s.log = slog.Default()
	}
	for _, o := range append(append([]string{}, DefaultOrigins...), cfg.AllowedOrigins...) {
		s.origins[strings.TrimRight(o, "/")] = true
	}
	return s
}

// ListenLoopback opens 127.0.0.1:port (0 = pick a free port).
func ListenLoopback(port int) (net.Listener, error) {
	return net.Listen("tcp4", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
}

// Serve serves on ln until ctx is cancelled.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	_, s.port, _ = net.SplitHostPort(ln.Addr().String())
	srv := &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		sh, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(sh)
	}()
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Handler returns the routed, protected handler (also used by tests).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.health)
	mux.HandleFunc("GET /v1/network", s.network)
	mux.HandleFunc("GET /v1/devices", s.devices)
	mux.HandleFunc("GET /v1/devices/{id}", s.device)
	mux.HandleFunc("PATCH /v1/devices/{id}", s.patchDevice)
	mux.HandleFunc("POST /v1/devices/{id}/merge", s.mergeDevice)
	mux.HandleFunc("GET /v1/devices/{id}/history", s.history)
	mux.HandleFunc("GET /v1/devices/{id}/evidence", s.deviceEvidence)
	mux.HandleFunc("POST /v1/devices/{id}/ports", s.devicePorts)
	mux.HandleFunc("GET /v1/events", s.events)
	mux.HandleFunc("POST /v1/scans", s.startScan)
	mux.HandleFunc("GET /v1/scans/{id}", s.scanSnapshot)
	mux.HandleFunc("GET /v1/scans/{id}/stream", s.scanStream)
	mux.HandleFunc("POST /v1/ping", s.ping)
	mux.HandleFunc("POST /v1/wol", s.wol)
	mux.HandleFunc("GET /v1/settings", s.getSettings)
	mux.HandleFunc("PUT /v1/settings", s.putSettings)
	mux.HandleFunc("GET /v1/data/stats", s.dataStats)
	mux.HandleFunc("POST /v1/data/vacuum", s.vacuum)
	mux.HandleFunc("POST /v1/data/integrity", s.integrity)
	mux.HandleFunc("POST /v1/data/open", s.openData)
	mux.HandleFunc("POST /v1/data/prune", s.pruneData)
	mux.HandleFunc("DELETE /v1/data", s.clearData)
	return s.protect(mux)
}

// ---- middleware -----------------------------------------------------------

func (s *Server) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Error("panic in handler", "path", r.URL.Path, "panic", rec)
				writeErr(w, http.StatusInternalServerError, "internal", "internal error")
			}
		}()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// 1. Host must be this loopback server (anti DNS-rebinding).
		if !s.hostOK(r.Host) {
			writeErr(w, http.StatusForbidden, "bad_host", "unexpected Host header")
			return
		}
		// 2. Origin, when present, must be the app shell or a configured dev origin.
		origin := r.Header.Get("Origin")
		if origin != "" {
			if !s.origins[strings.TrimRight(origin, "/")] {
				writeErr(w, http.StatusForbidden, "bad_origin", "origin not allowed")
				return
			}
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// 3. Bearer token (health is open so the shell can wait for startup).
		if r.URL.Path != "/v1/health" && !s.tokenOK(r) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeErr(w, http.StatusUnauthorized, "unauthorized", "missing or invalid token")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) hostOK(host string) bool {
	h, p, err := net.SplitHostPort(host)
	if err != nil {
		return false
	}
	if s.port != "" && p != s.port {
		return false
	}
	return h == "127.0.0.1" || h == "localhost"
}

func (s *Server) tokenOK(r *http.Request) bool {
	got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		// Fallback for Server-Sent Events (SSE) endpoints where native EventSource cannot send Authorization headers.
		// Allowed strictly on /stream paths, verified via constant-time comparison.
		if strings.HasSuffix(r.URL.Path, "/stream") {
			qToken := r.URL.Query().Get("token")
			if qToken != "" && s.cfg.Token != "" && subtle.ConstantTimeCompare([]byte(qToken), []byte(s.cfg.Token)) == 1 {
				return true
			}
		}
		return false
	}
	return s.cfg.Token != "" && subtle.ConstantTimeCompare([]byte(got), []byte(s.cfg.Token)) == 1
}

// ---- helpers --------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type errBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	var b errBody
	b.Error.Code, b.Error.Message = code, msg
	writeJSON(w, status, b)
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, engine.ErrInvalid):
		writeErr(w, http.StatusBadRequest, "invalid", err.Error())
	case errors.Is(err, engine.ErrNoNetwork):
		writeErr(w, http.StatusServiceUnavailable, "no_network", "No active network connection. Connect to a network and try again.")
	default:
		s.log.Error("request failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal", "Internal server error")
	}
}

func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%w: %v", engine.ErrInvalid, err)
	}
	return nil
}

// ---- handlers -------------------------------------------------------------

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "version": s.cfg.Version})
}

func (s *Server) network(w http.ResponseWriter, r *http.Request) {
	info, err := s.eng.NetworkInfo(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) devices(w http.ResponseWriter, r *http.Request) {
	ds, err := s.eng.Devices(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ds)
}

func (s *Server) device(w http.ResponseWriter, r *http.Request) {
	d, err := s.eng.Device(r.Context(), r.PathValue("id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) patchDevice(w http.ResponseWriter, r *http.Request) {
	var p model.DevicePatch
	if err := decode(r, &p); err != nil {
		s.fail(w, err)
		return
	}
	d, err := s.eng.UpdateDevice(r.Context(), r.PathValue("id"), p)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) mergeDevice(w http.ResponseWriter, r *http.Request) {
	var req model.MergeDevicesRequest
	if err := decode(r, &req); err != nil {
		s.fail(w, err)
		return
	}
	targetID := r.PathValue("id")
	if req.SourceID == "" {
		writeErr(w, http.StatusBadRequest, "invalid_argument", "sourceId is required")
		return
	}
	if err := s.eng.MergeDevices(r.Context(), targetID, req.SourceID); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) history(w http.ResponseWriter, r *http.Request) {
	h, err := s.eng.History(r.Context(), r.PathValue("id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h)
}

func (s *Server) deviceEvidence(w http.ResponseWriter, r *http.Request) {
	evs, err := s.eng.DeviceEvidence(r.Context(), r.PathValue("id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, evs)
}

func (s *Server) devicePorts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	svcs, err := s.eng.ScanDevicePorts(ctx, r.PathValue("id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, svcs)
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	ev, err := s.eng.Events(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ev)
}

func (s *Server) startScan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type string `json:"type"`
	}
	if r.ContentLength != 0 {
		if err := decode(r, &body); err != nil {
			s.fail(w, err)
			return
		}
	}
	id, err := s.eng.StartScan(body.Type)
	switch {
	case errors.Is(err, engine.ErrScanRunning):
		// Not an error for the UI: attach to the scan already in progress.
		writeJSON(w, http.StatusOK, map[string]any{"scanId": id, "alreadyRunning": true})
	case err != nil:
		s.fail(w, err)
	default:
		writeJSON(w, http.StatusAccepted, map[string]any{"scanId": id})
	}
}

func (s *Server) scanSnapshot(w http.ResponseWriter, r *http.Request) {
	sc := s.eng.Scan(r.PathValue("id"))
	if sc == nil {
		s.fail(w, store.ErrNotFound)
		return
	}
	snap, _ := sc.Watch()
	writeJSON(w, http.StatusOK, snap)
}

// scanStream emits Server-Sent Events:
//
//	event: progress  data: {scanId,type,progress,scanned,total,found,done:false}
//	event: done      data: <ScanResult>
//	event: scan-error data: {"message": "..."}
func (s *Server) scanStream(w http.ResponseWriter, r *http.Request) {
	sc := s.eng.Scan(r.PathValue("id"))
	if sc == nil {
		s.fail(w, store.ErrNotFound)
		return
	}
	fl, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "internal", "streaming unsupported")
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	send := func(event string, v any) {
		b, _ := json.Marshal(v)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		fl.Flush()
	}
	hb := time.NewTicker(15 * time.Second)
	defer hb.Stop()
	for {
		snap, changed := sc.Watch()
		switch {
		case snap.Done && snap.Error != "":
			send("scan-error", map[string]string{"message": snap.Error})
			return
		case snap.Done:
			send("done", snap.Result)
			return
		default:
			send("progress", snap)
		}
		// Coalesce bursts: at most ~8 progress frames per second.
		select {
		case <-r.Context().Done():
			return
		case <-changed:
			select {
			case <-r.Context().Done():
				return
			case <-time.After(120 * time.Millisecond):
			}
		case <-hb.C:
			fmt.Fprint(w, ": keep-alive\n\n")
			fl.Flush()
		}
	}
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IP string `json:"ip"`
	}
	if err := decode(r, &body); err != nil {
		s.fail(w, err)
		return
	}
	res, err := s.eng.Ping(r.Context(), body.IP)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) wol(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MAC string `json:"mac"`
	}
	if err := decode(r, &body); err != nil {
		s.fail(w, err)
		return
	}
	res, err := s.eng.WakeOnLAN(r.Context(), body.MAC)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	st, err := s.eng.Settings(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *Server) putSettings(w http.ResponseWriter, r *http.Request) {
	var st model.Settings
	if err := decode(r, &st); err != nil {
		s.fail(w, err)
		return
	}
	switch st.ScanInterval {
	case "1m", "5m", "15m", "1h", "manual":
		// valid
	default:
		writeErr(w, http.StatusBadRequest, "invalid_settings", "scanInterval must be one of: 1m, 5m, 15m, 1h, manual")
		return
	}
	if err := s.eng.UpdateSettings(r.Context(), st); err != nil {
		s.fail(w, err)
		return
	}
	if s.cfg.Autostart != nil {
		if err := s.cfg.Autostart.Set(st.LaunchAtStartup); err != nil {
			s.log.Warn("failed to update autostart setting", "error", err)
		}
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *Server) dataStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.eng.DatabaseStats(r.Context(), s.cfg.DBPath)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) vacuum(w http.ResponseWriter, r *http.Request) {
	if err := s.eng.Vacuum(r.Context()); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, model.MaintenanceResult{Success: true, Message: "Database defragmented and compacted successfully."})
}

func (s *Server) integrity(w http.ResponseWriter, r *http.Request) {
	res, err := s.eng.IntegrityCheck(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	msg := fmt.Sprintf("Integrity check result: %s", res)
	writeJSON(w, http.StatusOK, model.MaintenanceResult{Success: res == "ok", Message: msg})
}

func (s *Server) openData(w http.ResponseWriter, r *http.Request) {
	if s.OpenDataFolder == nil {
		writeErr(w, http.StatusNotImplemented, "unsupported", "opening the data folder is not available in this host")
		return
	}
	if err := s.OpenDataFolder(); err != nil {
		s.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) clearData(w http.ResponseWriter, r *http.Request) {
	if err := s.eng.ClearHistory(r.Context()); err != nil {
		s.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) pruneData(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OlderThanDays int `json:"olderThanDays"`
	}
	if err := decode(r, &body); err != nil {
		s.fail(w, err)
		return
	}
	if body.OlderThanDays < 1 {
		body.OlderThanDays = 30
	}
	deleted, err := s.eng.PruneEvents(r.Context(), body.OlderThanDays)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"deletedCount":  deleted,
		"olderThanDays": body.OlderThanDays,
	})
}

// ParseOrigin validates a user-supplied origin flag.
func ParseOrigin(o string) (string, error) {
	u, err := url.Parse(o)
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Path != "" && u.Path != "/") {
		return "", fmt.Errorf("invalid origin %q (expected scheme://host[:port])", o)
	}
	hostname := u.Hostname()
	if hostname != "localhost" && hostname != "127.0.0.1" && hostname != "::1" && !strings.HasSuffix(hostname, ".localhost") {
		return "", fmt.Errorf("origin host %q must be a loopback address", hostname)
	}
	return u.Scheme + "://" + u.Host, nil
}
