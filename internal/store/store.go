// Package store persists the device inventory, events and scan history in a
// local SQLite file. SQLite is the single source of truth; nothing is cached
// elsewhere and nothing leaves the machine.
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"netwatch/internal/model"
)

var ErrNotFound = errors.New("not found")

// Known is a stored device plus engine bookkeeping that is not sent to the UI.
type Known struct {
	model.Device
	Net    string // network key the device belongs to
	Missed int    // consecutive scans in which the device was not seen
}

// HistoryRow is a per-device timeline entry before it has an ID.
type HistoryRow struct {
	DeviceID string
	Time     time.Time
	Type     string
	Desc     string
}

// ScanRecord is one completed scan.
type ScanRecord struct {
	ID         string
	Type       string
	StartedAt  time.Time
	DurationMs int64
	Scanned    int
	Total      int
	Found      int
	NewDevices int
	Errors     int
}

// Writeset is applied atomically by Commit.
type Writeset struct {
	Devices []Known
	Events  []model.NetworkEvent
	History []HistoryRow
	// Services replaces the stored open-port set of each listed device.
	Services map[string][]model.DeviceService
	Scan     *ScanRecord
}

type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the database at path. ":memory:" is allowed.
func Open(path string) (*Store, error) {
	dsn := path
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
		dsn = "file:" + filepath.ToSlash(path)
	} else {
		dsn = "file::memory:"
	}
	dsn += "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // one writer; avoids lock contention entirely
	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

var migrations = []string{
	// v1
	`CREATE TABLE devices (
		id           TEXT PRIMARY KEY,
		net          TEXT NOT NULL,
		mac          TEXT NOT NULL,
		ip           TEXT NOT NULL,
		hostname     TEXT NOT NULL DEFAULT '',
		vendor       TEXT NOT NULL DEFAULT '',
		type         TEXT NOT NULL DEFAULT 'Unknown',
		name         TEXT NOT NULL DEFAULT '',
		custom_alias TEXT NOT NULL DEFAULT '',
		notes        TEXT NOT NULL DEFAULT '',
		os           TEXT NOT NULL DEFAULT '',
		status       TEXT NOT NULL,
		is_new       INTEGER NOT NULL DEFAULT 0,
		missed       INTEGER NOT NULL DEFAULT 0,
		latency_ms   INTEGER,
		first_seen   INTEGER NOT NULL,
		last_seen    INTEGER NOT NULL
	);
	CREATE INDEX idx_devices_net ON devices(net);
	CREATE TABLE services (
		device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
		port      INTEGER NOT NULL,
		protocol  TEXT NOT NULL,
		service   TEXT NOT NULL,
		status    TEXT NOT NULL,
		PRIMARY KEY (device_id, port, protocol)
	);
	CREATE TABLE device_events (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
		ts        INTEGER NOT NULL,
		type      TEXT NOT NULL,
		description TEXT NOT NULL
	);
	CREATE INDEX idx_device_events ON device_events(device_id, ts DESC);
	CREATE TABLE events (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		ts          INTEGER NOT NULL,
		type        TEXT NOT NULL,
		title       TEXT NOT NULL,
		device_id   TEXT NOT NULL DEFAULT '',
		device_name TEXT NOT NULL DEFAULT '',
		ip          TEXT NOT NULL DEFAULT '',
		mac         TEXT NOT NULL DEFAULT '',
		details     TEXT NOT NULL DEFAULT ''
	);
	CREATE INDEX idx_events_ts ON events(ts DESC);
	CREATE TABLE scans (
		id          TEXT PRIMARY KEY,
		type        TEXT NOT NULL,
		started_at  INTEGER NOT NULL,
		duration_ms INTEGER NOT NULL,
		scanned     INTEGER NOT NULL,
		total       INTEGER NOT NULL,
		found       INTEGER NOT NULL,
		new_devices INTEGER NOT NULL,
		errors      INTEGER NOT NULL
	);
	CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);`,
}

func (s *Store) migrate(ctx context.Context) error {
	var ver int
	if err := s.db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&ver); err != nil {
		return err
	}
	for i := ver; i < len(migrations); i++ {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration v%d: %w", i+1, err)
		}
		if _, err := tx.ExecContext(ctx, `PRAGMA user_version = `+strconv.Itoa(i+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func ms(t time.Time) int64     { return t.UnixMilli() }
func fromMs(v int64) time.Time { return time.UnixMilli(v).UTC() }
func nullInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

const deviceCols = `id, net, mac, ip, hostname, vendor, type, name, custom_alias, notes, os, status, is_new, missed, latency_ms, first_seen, last_seen`

func scanKnown(sc interface{ Scan(...any) error }) (Known, error) {
	var (
		k          Known
		isNew      int
		lat        sql.NullInt64
		first, lst int64
	)
	err := sc.Scan(&k.ID, &k.Net, &k.MAC, &k.IP, &k.Hostname, &k.Vendor, &k.Type, &k.Name, &k.CustomAlias,
		&k.Notes, &k.OS, &k.Status, &isNew, &k.Missed, &lat, &first, &lst)
	if err != nil {
		return k, err
	}
	k.IsNew = isNew != 0
	if lat.Valid {
		v := int(lat.Int64)
		k.LatencyMs = &v
	}
	k.FirstSeen, k.LastSeen = fromMs(first), fromMs(lst)
	return k, nil
}

// ListKnown returns the devices of one network with services attached (no history).
func (s *Store) ListKnown(ctx context.Context, net string) ([]Known, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+deviceCols+` FROM devices WHERE net = ?`, net)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Known
	idx := map[string]int{}
	for rows.Next() {
		k, err := scanKnown(rows)
		if err != nil {
			return nil, err
		}
		idx[k.ID] = len(out)
		out = append(out, k)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	srows, err := s.db.QueryContext(ctx, `SELECT s.device_id, s.port, s.protocol, s.service, s.status
		FROM services s JOIN devices d ON d.id = s.device_id WHERE d.net = ? ORDER BY s.device_id, s.port`, net)
	if err != nil {
		return nil, err
	}
	defer srows.Close()
	for srows.Next() {
		var id string
		var sv model.DeviceService
		if err := srows.Scan(&id, &sv.Port, &sv.Protocol, &sv.Service, &sv.Status); err != nil {
			return nil, err
		}
		if i, ok := idx[id]; ok {
			out[i].Services = append(out[i].Services, sv)
		}
	}
	return out, srows.Err()
}

// GetDevice returns one device with services and its latest history.
func (s *Store) GetDevice(ctx context.Context, id string) (*model.Device, error) {
	k, err := scanKnown(s.db.QueryRowContext(ctx, `SELECT `+deviceCols+` FROM devices WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	d := k.Device
	if d.Services, err = s.services(ctx, id); err != nil {
		return nil, err
	}
	if d.History, err = s.DeviceHistory(ctx, id, 100); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) services(ctx context.Context, id string) ([]model.DeviceService, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT port, protocol, service, status FROM services WHERE device_id = ? ORDER BY port`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.DeviceService
	for rows.Next() {
		var sv model.DeviceService
		if err := rows.Scan(&sv.Port, &sv.Protocol, &sv.Service, &sv.Status); err != nil {
			return nil, err
		}
		out = append(out, sv)
	}
	return out, rows.Err()
}

func (s *Store) DeviceHistory(ctx context.Context, id string, limit int) ([]model.DeviceEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, ts, type, description FROM device_events WHERE device_id = ? ORDER BY ts DESC, id DESC LIMIT ?`, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.DeviceEvent{}
	for rows.Next() {
		var e model.DeviceEvent
		var rid, ts int64
		if err := rows.Scan(&rid, &ts, &e.Type, &e.Description); err != nil {
			return nil, err
		}
		e.ID, e.Timestamp = "h-"+strconv.FormatInt(rid, 10), fromMs(ts)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) Events(ctx context.Context, limit int) ([]model.NetworkEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, ts, type, title, device_id, device_name, ip, mac, details FROM events ORDER BY ts DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.NetworkEvent{}
	for rows.Next() {
		var e model.NetworkEvent
		var rid, ts int64
		if err := rows.Scan(&rid, &ts, &e.Type, &e.Title, &e.DeviceID, &e.DeviceName, &e.IP, &e.MAC, &e.Details); err != nil {
			return nil, err
		}
		e.ID, e.Timestamp = "evt-"+strconv.FormatInt(rid, 10), fromMs(ts)
		out = append(out, e)
	}
	return out, rows.Err()
}

// UpdateDevice applies user edits. Returns ErrNotFound for an unknown id.
func (s *Store) UpdateDevice(ctx context.Context, id string, p model.DevicePatch) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM devices WHERE id = ?`, id).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	if p.CustomAlias != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE devices SET custom_alias = ? WHERE id = ?`, *p.CustomAlias, id); err != nil {
			return err
		}
	}
	if p.Notes != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE devices SET notes = ? WHERE id = ?`, *p.Notes, id); err != nil {
			return err
		}
	}
	if p.IsNew != nil {
		v := 0
		if *p.IsNew {
			v = 1
		}
		if _, err := tx.ExecContext(ctx, `UPDATE devices SET is_new = ? WHERE id = ?`, v, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Commit applies a Writeset in one transaction.
func (s *Store) Commit(ctx context.Context, w Writeset) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, k := range w.Devices {
		isNew := 0
		if k.IsNew {
			isNew = 1
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO devices (`+deviceCols+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET mac=excluded.mac, ip=excluded.ip, hostname=excluded.hostname, vendor=excluded.vendor,
				type=excluded.type, name=excluded.name, os=excluded.os, status=excluded.status, is_new=excluded.is_new,
				missed=excluded.missed, latency_ms=excluded.latency_ms, last_seen=excluded.last_seen`,
			k.ID, k.Net, k.MAC, k.IP, k.Hostname, k.Vendor, k.Type, k.Name, k.CustomAlias, k.Notes, k.OS, k.Status,
			isNew, k.Missed, nullInt(k.LatencyMs), ms(k.FirstSeen), ms(k.LastSeen))
		if err != nil {
			return fmt.Errorf("upsert %s: %w", k.ID, err)
		}
	}
	for id, svcs := range w.Services {
		if _, err := tx.ExecContext(ctx, `DELETE FROM services WHERE device_id = ?`, id); err != nil {
			return err
		}
		for _, sv := range svcs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO services(device_id, port, protocol, service, status) VALUES (?,?,?,?,?)`,
				id, sv.Port, sv.Protocol, sv.Service, sv.Status); err != nil {
				return err
			}
		}
	}
	for _, h := range w.History {
		if _, err := tx.ExecContext(ctx, `INSERT INTO device_events(device_id, ts, type, description) VALUES (?,?,?,?)`,
			h.DeviceID, ms(h.Time), h.Type, h.Desc); err != nil {
			return err
		}
	}
	for _, e := range w.Events {
		if _, err := tx.ExecContext(ctx, `INSERT INTO events(ts, type, title, device_id, device_name, ip, mac, details) VALUES (?,?,?,?,?,?,?,?)`,
			ms(e.Timestamp), e.Type, e.Title, e.DeviceID, e.DeviceName, e.IP, e.MAC, e.Details); err != nil {
			return err
		}
	}
	if sc := w.Scan; sc != nil {
		if _, err := tx.ExecContext(ctx, `INSERT INTO scans(id, type, started_at, duration_ms, scanned, total, found, new_devices, errors) VALUES (?,?,?,?,?,?,?,?,?)`,
			sc.ID, sc.Type, ms(sc.StartedAt), sc.DurationMs, sc.Scanned, sc.Total, sc.Found, sc.NewDevices, sc.Errors); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Meta(ctx context.Context, key string) (string, error) {
	var v string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

func (s *Store) SetMeta(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO meta(key, value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// ClearHistory removes devices, events and scans (the "Clear History" action).
func (s *Store) ClearHistory(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, t := range []string{"events", "device_events", "services", "devices", "scans"} {
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+t); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// PruneEvents removes audit log events and device_events older than olderThanDays.
// Returns the total number of deleted rows.
func (s *Store) PruneEvents(ctx context.Context, olderThanDays int) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -olderThanDays)
	cutoffMs := ms(cutoff)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res1, err := tx.ExecContext(ctx, `DELETE FROM events WHERE ts < ?`, cutoffMs)
	if err != nil {
		return 0, fmt.Errorf("prune events: %w", err)
	}
	n1, _ := res1.RowsAffected()

	res2, err := tx.ExecContext(ctx, `DELETE FROM device_events WHERE ts < ?`, cutoffMs)
	if err != nil {
		return 0, fmt.Errorf("prune device_events: %w", err)
	}
	n2, _ := res2.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n1 + n2, nil
}

const settingsMetaKey = "app_settings"

// GetSettings retrieves persistent user settings or returns default settings if unset.
func (s *Store) GetSettings(ctx context.Context) (model.Settings, error) {
	val, err := s.Meta(ctx, settingsMetaKey)
	if err != nil {
		return model.DefaultSettings(), err
	}
	if val == "" {
		return model.DefaultSettings(), nil
	}
	st := model.DefaultSettings()
	if err := json.Unmarshal([]byte(val), &st); err != nil {
		return model.DefaultSettings(), nil
	}
	return st, nil
}

// SaveSettings persists user settings as JSON in the meta table.
func (s *Store) SaveSettings(ctx context.Context, st model.Settings) error {
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return s.SetMeta(ctx, settingsMetaKey, string(b))
}

// Stats returns database metrics, record counts, and WAL journal status.
func (s *Store) Stats(ctx context.Context, dbPath string) (model.DatabaseStats, error) {
	res := model.DatabaseStats{DBPath: dbPath}
	if dbPath != "" && dbPath != ":memory:" {
		if fi, err := os.Stat(dbPath); err == nil {
			res.FileSizeBytes = fi.Size()
		}
	}
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM devices`).Scan(&res.DeviceCount); err != nil {
		return res, fmt.Errorf("count devices: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM events`).Scan(&res.EventCount); err != nil {
		return res, fmt.Errorf("count events: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM scans`).Scan(&res.ScanCount); err != nil {
		return res, fmt.Errorf("count scans: %w", err)
	}
	var jMode string
	if err := s.db.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&jMode); err != nil {
		return res, fmt.Errorf("query journal_mode: %w", err)
	}
	res.WALEnabled = strings.EqualFold(jMode, "wal")
	return res, nil
}

// Vacuum defragments and compacts the SQLite database file.
func (s *Store) Vacuum(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `VACUUM`)
	return err
}

// IntegrityCheck verifies the structural consistency of the SQLite database.
func (s *Store) IntegrityCheck(ctx context.Context) (string, error) {
	var res string
	err := s.db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}
