package sync

import (
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	// maxAuditLogRows is the maximum number of rows to keep in the sync_audit_log table.
	// Older entries are pruned during each write to keep the table bounded.
	maxAuditLogRows = 500
)

// AuditLogger writes sync events to the persistent sync_audit_log SQLite table.
// It is designed to be embedded into SyncStatus so that all Log() and RecordConflict()
// calls automatically persist for debugging.
type AuditLogger struct {
	db *sqlx.DB
}

// NewAuditLogger creates a new AuditLogger that writes to the given database.
func NewAuditLogger(db *sqlx.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// Write inserts a sync event into the audit log and prunes old entries.
func (a *AuditLogger) Write(direction, entityType, entityID, operation, status, details string, count int) {
	if a == nil || a.db == nil {
		return
	}

	_, err := a.db.Exec(
		`INSERT INTO sync_audit_log (direction, entity_type, entity_id, operation, status, details, count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		direction, entityType, entityID, operation, status, details, count,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		slog.Error("audit: failed to write sync_audit_log", "error", err)
		return
	}

	// Prune old entries to keep the table bounded
	a.prune()
}

// prune removes old entries beyond the maximum row count.
func (a *AuditLogger) prune() {
	a.db.Exec(
		`DELETE FROM sync_audit_log WHERE id NOT IN (
			SELECT id FROM sync_audit_log ORDER BY created_at DESC LIMIT ?
		)`, maxAuditLogRows,
	)
}

// AuditLogEntry represents a single row from the sync_audit_log table.
type AuditLogEntry struct {
	ID         int    `db:"id" json:"id"`
	Direction  string `db:"direction" json:"direction"`
	EntityType string `db:"entity_type" json:"entity_type"`
	EntityID   string `db:"entity_id" json:"entity_id"`
	Operation  string `db:"operation" json:"operation"`
	Status     string `db:"status" json:"status"`
	Details    string `db:"details" json:"details"`
	Count      int    `db:"count" json:"count"`
	CreatedAt  string `db:"created_at" json:"created_at"`
}

// QueryRecent returns the most recent N audit log entries.
func (a *AuditLogger) QueryRecent(limit int) ([]AuditLogEntry, error) {
	if a == nil || a.db == nil {
		return nil, nil
	}
	var entries []AuditLogEntry
	err := a.db.Select(&entries,
		`SELECT id, direction, entity_type, entity_id, operation, status, details, count, created_at
		 FROM sync_audit_log ORDER BY created_at DESC LIMIT ?`, limit)
	return entries, err
}

// QueryByEntity returns audit log entries for a specific entity.
func (a *AuditLogger) QueryByEntity(entityType, entityID string, limit int) ([]AuditLogEntry, error) {
	if a == nil || a.db == nil {
		return nil, nil
	}
	var entries []AuditLogEntry
	err := a.db.Select(&entries,
		`SELECT id, direction, entity_type, entity_id, operation, status, details, count, created_at
		 FROM sync_audit_log WHERE entity_type = ? AND entity_id = ?
		 ORDER BY created_at DESC LIMIT ?`, entityType, entityID, limit)
	return entries, err
}

// QueryConflicts returns only conflict entries from the audit log.
func (a *AuditLogger) QueryConflicts(limit int) ([]AuditLogEntry, error) {
	if a == nil || a.db == nil {
		return nil, nil
	}
	var entries []AuditLogEntry
	err := a.db.Select(&entries,
		`SELECT id, direction, entity_type, entity_id, operation, status, details, count, created_at
		 FROM sync_audit_log WHERE status LIKE 'conflict%'
		 ORDER BY created_at DESC LIMIT ?`, limit)
	return entries, err
}
