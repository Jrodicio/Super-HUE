package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"superhue/internal/domain"
)

type SettingsRepository struct{ db *sql.DB }

type HueRepository struct{ db *sql.DB }

type RuleRepository struct{ db *sql.DB }

type DeviceRepository struct{ db *sql.DB }

type LogRepository struct{ db *sql.DB }

func NewSettingsRepository(db *sql.DB) *SettingsRepository { return &SettingsRepository{db: db} }
func NewHueRepository(db *sql.DB) *HueRepository           { return &HueRepository{db: db} }
func NewRuleRepository(db *sql.DB) *RuleRepository         { return &RuleRepository{db: db} }
func NewDeviceRepository(db *sql.DB) *DeviceRepository     { return &DeviceRepository{db: db} }
func NewLogRepository(db *sql.DB) *LogRepository           { return &LogRepository{db: db} }

func (r *SettingsRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (r *SettingsRepository) Set(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO settings(key, value, updated_at) VALUES(?, ?, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`, key, value)
	return err
}

func (r *SettingsRepository) List(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, rows.Err()
}

func (r *HueRepository) SaveLights(ctx context.Context, lights []domain.Light) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM lights_cache`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO lights_cache(id, name, room_id, room_name, on_state, brightness, color_hex, reachable, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, light := range lights {
		if _, err := stmt.ExecContext(ctx, light.ID, light.Name, light.RoomID, light.RoomName, boolToInt(light.On), light.Brightness, light.ColorHex, boolToInt(light.Reachable)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *HueRepository) SaveRooms(ctx context.Context, rooms []domain.Room) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM rooms`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO rooms(id, name, type, updated_at) VALUES(?, ?, ?, CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, room := range rooms {
		if _, err := stmt.ExecContext(ctx, room.ID, room.Name, room.Type); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *HueRepository) SaveScenes(ctx context.Context, scenes []domain.Scene) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM scenes`); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO scenes(id, name, group_id, group_name, updated_at) VALUES(?, ?, ?, ?, CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, scene := range scenes {
		if _, err := stmt.ExecContext(ctx, scene.ID, scene.Name, scene.GroupID, scene.Group); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *HueRepository) ListLights(ctx context.Context) ([]domain.Light, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, room_id, room_name, on_state, brightness, color_hex, reachable FROM lights_cache ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Light
	for rows.Next() {
		var light domain.Light
		var onState, reachable int
		if err := rows.Scan(&light.ID, &light.Name, &light.RoomID, &light.RoomName, &onState, &light.Brightness, &light.ColorHex, &reachable); err != nil {
			return nil, err
		}
		light.On = onState == 1
		light.Reachable = reachable == 1
		result = append(result, light)
	}
	return result, rows.Err()
}

func (r *HueRepository) ListRooms(ctx context.Context) ([]domain.Room, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, type FROM rooms ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Room
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Type); err != nil {
			return nil, err
		}
		result = append(result, room)
	}
	return result, rows.Err()
}

func (r *HueRepository) ListScenes(ctx context.Context) ([]domain.Scene, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, group_id, group_name FROM scenes ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Scene
	for rows.Next() {
		var scene domain.Scene
		if err := rows.Scan(&scene.ID, &scene.Name, &scene.GroupID, &scene.Group); err != nil {
			return nil, err
		}
		result = append(result, scene)
	}
	return result, rows.Err()
}

func (r *HueRepository) SaveRoom(ctx context.Context, room *domain.Room) error {
	room.ID = strings.TrimSpace(room.ID)
	room.Name = strings.TrimSpace(room.Name)
	if room.ID == "" {
		room.ID = fmt.Sprintf("local-%d", time.Now().UTC().UnixNano())
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO rooms(id, name, type, updated_at) VALUES(?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(id) DO UPDATE SET name = excluded.name, type = excluded.type, updated_at = CURRENT_TIMESTAMP`, room.ID, room.Name, room.Type)
	return err
}

func (r *HueRepository) DeleteRoom(ctx context.Context, roomID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE lights_cache SET room_id = '', room_name = '' WHERE room_id = ?`, roomID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM rooms WHERE id = ?`, roomID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *HueRepository) AssignLightRoom(ctx context.Context, lightID, roomID, roomName string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE lights_cache SET room_id = ?, room_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, roomID, roomName, lightID)
	return err
}

func (r *RuleRepository) List(ctx context.Context) ([]domain.Rule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, trigger_type, enabled, created_at, updated_at FROM rules ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []domain.Rule
	for rows.Next() {
		var rule domain.Rule
		var enabled int
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Trigger, &enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rule.Enabled = enabled == 1
		rule.Conditions, err = r.loadConditions(ctx, rule.ID)
		if err != nil {
			return nil, err
		}
		rule.Actions, err = r.loadActions(ctx, rule.ID)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *RuleRepository) Save(ctx context.Context, rule *domain.Rule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	if rule.ID == 0 {
		res, err := tx.ExecContext(ctx, `INSERT INTO rules(name, trigger_type, enabled, created_at, updated_at) VALUES(?, ?, ?, ?, ?)`, rule.Name, rule.Trigger, boolToInt(rule.Enabled), now, now)
		if err != nil {
			return err
		}
		rule.ID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE rules SET name = ?, trigger_type = ?, enabled = ?, updated_at = ? WHERE id = ?`, rule.Name, rule.Trigger, boolToInt(rule.Enabled), now, rule.ID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM rule_actions WHERE rule_id = ?`, rule.ID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM rule_conditions WHERE rule_id = ?`, rule.ID); err != nil {
			return err
		}
	}
	for _, cond := range rule.Conditions {
		if _, err := tx.ExecContext(ctx, `INSERT INTO rule_conditions(rule_id, condition_type, key_name, value, negate) VALUES(?, ?, ?, ?, ?)`, rule.ID, cond.Type, cond.Key, cond.Value, boolToInt(cond.Negate)); err != nil {
			return err
		}
	}
	for _, action := range rule.Actions {
		if _, err := tx.ExecContext(ctx, `INSERT INTO rule_actions(rule_id, action_type, target, value) VALUES(?, ?, ?, ?)`, rule.ID, action.Type, action.Target, action.Value); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *RuleRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM rules WHERE id = ?`, id)
	return err
}

func (r *RuleRepository) loadConditions(ctx context.Context, ruleID int64) ([]domain.Condition, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, condition_type, key_name, value, negate FROM rule_conditions WHERE rule_id = ? ORDER BY id`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Condition
	for rows.Next() {
		var cond domain.Condition
		var negate int
		if err := rows.Scan(&cond.ID, &cond.Type, &cond.Key, &cond.Value, &negate); err != nil {
			return nil, err
		}
		cond.RuleID = ruleID
		cond.Negate = negate == 1
		result = append(result, cond)
	}
	return result, rows.Err()
}

func (r *RuleRepository) loadActions(ctx context.Context, ruleID int64) ([]domain.Action, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, action_type, target, value FROM rule_actions WHERE rule_id = ? ORDER BY id`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Action
	for rows.Next() {
		var action domain.Action
		if err := rows.Scan(&action.ID, &action.Type, &action.Target, &action.Value); err != nil {
			return nil, err
		}
		action.RuleID = ruleID
		result = append(result, action)
	}
	return result, rows.Err()
}

func (r *DeviceRepository) List(ctx context.Context) ([]domain.Device, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, ip, present, failure_count, consecutive_oks, COALESCE(last_seen_at, ''), COALESCE(last_checked_at, '') FROM devices ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Device
	for rows.Next() {
		var d domain.Device
		var present int
		var lastSeen, lastChecked string
		if err := rows.Scan(&d.ID, &d.Name, &d.IP, &present, &d.FailureCount, &d.ConsecutiveOKs, &lastSeen, &lastChecked); err != nil {
			return nil, err
		}
		d.Present = present == 1
		d.LastSeenAt = parseDBTime(lastSeen)
		d.LastCheckedAt = parseDBTime(lastChecked)
		result = append(result, d)
	}
	return result, rows.Err()
}

func (r *DeviceRepository) Save(ctx context.Context, device *domain.Device) error {
	if device.ID == 0 {
		res, err := r.db.ExecContext(ctx, `INSERT INTO devices(name, ip, present, failure_count, consecutive_oks, last_seen_at, last_checked_at) VALUES(?, ?, ?, ?, ?, ?, ?)`, device.Name, device.IP, boolToInt(device.Present), device.FailureCount, device.ConsecutiveOKs, nullableTime(device.LastSeenAt), nullableTime(device.LastCheckedAt))
		if err != nil {
			return err
		}
		device.ID, err = res.LastInsertId()
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE devices SET name = ?, ip = ?, present = ?, failure_count = ?, consecutive_oks = ?, last_seen_at = ?, last_checked_at = ? WHERE id = ?`, device.Name, device.IP, boolToInt(device.Present), device.FailureCount, device.ConsecutiveOKs, nullableTime(device.LastSeenAt), nullableTime(device.LastCheckedAt), device.ID)
	return err
}

func (r *DeviceRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM devices WHERE id = ?`, id)
	return err
}

func (r *DeviceRepository) UpdateStatus(ctx context.Context, device domain.Device) error {
	_, err := r.db.ExecContext(ctx, `UPDATE devices SET present = ?, failure_count = ?, consecutive_oks = ?, last_seen_at = ?, last_checked_at = ? WHERE id = ?`, boolToInt(device.Present), device.FailureCount, device.ConsecutiveOKs, nullableTime(device.LastSeenAt), nullableTime(device.LastCheckedAt), device.ID)
	return err
}

func (r *LogRepository) Add(ctx context.Context, entry *domain.LogEntry) error {
	res, err := r.db.ExecContext(ctx, `INSERT INTO logs(level, source, message, created_at) VALUES(?, ?, ?, ?)`, entry.Level, entry.Source, entry.Message, entry.CreatedAt.UTC())
	if err != nil {
		return err
	}
	entry.ID, _ = res.LastInsertId()
	return nil
}

func (r *LogRepository) List(ctx context.Context, limit int) ([]domain.LogEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, level, source, message, created_at FROM logs ORDER BY created_at DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.LogEntry
	for rows.Next() {
		var entry domain.LogEntry
		if err := rows.Scan(&entry.ID, &entry.Level, &entry.Source, &entry.Message, &entry.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, rows.Err()
}

func parseDBTime(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	formats := []string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05-07:00", "2006-01-02 15:04:05"}
	for _, format := range formats {
		if t, err := time.Parse(format, raw); err == nil {
			return t
		}
	}
	return time.Time{}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

var _ domain.SettingsRepository = (*SettingsRepository)(nil)
var _ domain.HueRepository = (*HueRepository)(nil)
var _ domain.RuleRepository = (*RuleRepository)(nil)
var _ domain.DeviceRepository = (*DeviceRepository)(nil)
var _ domain.LogRepository = (*LogRepository)(nil)

func ensure(condition bool, format string, args ...any) error {
	if !condition {
		return fmt.Errorf(format, args...)
	}
	return nil
}
