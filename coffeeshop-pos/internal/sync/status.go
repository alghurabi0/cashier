package sync

import (
	"sync"
	"time"
)

const maxLogEntries = 50
const maxConflictRecords = 20

// ConflictRecord represents a detected conflict during sync.
type ConflictRecord struct {
	Time       string `json:"time"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Resolution string `json:"resolution"` // "lww", "manual" (future)
}

// SyncLogEntry represents a single sync operation log entry.
type SyncLogEntry struct {
	Time      string `json:"time"`
	Operation string `json:"operation"` // "pull", "push", "health", "retry"
	Entity    string `json:"entity"`    // "categories", "orders", etc.
	Status    string `json:"status"`    // "ok", "error", "skipped"
	Message   string `json:"message"`
	Count     int    `json:"count"` // items synced/pushed
}

// SyncStatusSnapshot is the JSON-serializable snapshot returned to the frontend.
type SyncStatusSnapshot struct {
	// Per-table timestamps
	TableSyncTimes map[string]string `json:"table_sync_times"`

	// Connection state
	IsConnected      bool   `json:"is_connected"`
	LastHealthCheckAt string `json:"last_health_check_at"`
	LastConnectError  string `json:"last_connect_error"`

	// Order push state
	PendingOrders int `json:"pending_orders"`
	FailedOrders  int `json:"failed_orders"`

	// Sync cycle tracking
	LastPullAt        string `json:"last_pull_at"`
	LastPushAt        string `json:"last_push_at"`
	ConsecutiveErrors int    `json:"consecutive_errors"`
	IsSyncing         bool   `json:"is_syncing"`

	// Recent log entries
	RecentLogs []SyncLogEntry `json:"recent_logs"`

	// Conflict history
	Conflicts []ConflictRecord `json:"conflicts"`
}

// SyncStatus is a thread-safe in-memory tracker for sync health metrics.
type SyncStatus struct {
	mu sync.RWMutex

	tableSyncTimes map[string]string

	isConnected      bool
	lastHealthCheckAt string
	lastConnectError  string

	pendingOrders int
	failedOrders  int

	lastPullAt        string
	lastPushAt        string
	consecutiveErrors int
	isSyncing         bool

	recentLogs []SyncLogEntry
	conflicts  []ConflictRecord

	auditLogger *AuditLogger
}

// NewSyncStatus creates a new SyncStatus tracker.
func NewSyncStatus() *SyncStatus {
	return &SyncStatus{
		tableSyncTimes: make(map[string]string),
		recentLogs:     make([]SyncLogEntry, 0, maxLogEntries),
		conflicts:      make([]ConflictRecord, 0, maxConflictRecords),
	}
}

// SetAuditLogger attaches a persistent audit logger to the sync status tracker.
// Once set, all Log() and RecordConflict() calls also write to the SQLite audit table.
func (s *SyncStatus) SetAuditLogger(logger *AuditLogger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLogger = logger
}

// Snapshot returns a read-only copy of the current sync status.
func (s *SyncStatus) Snapshot() SyncStatusSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Copy table sync times
	times := make(map[string]string, len(s.tableSyncTimes))
	for k, v := range s.tableSyncTimes {
		times[k] = v
	}

	// Copy logs
	logs := make([]SyncLogEntry, len(s.recentLogs))
	copy(logs, s.recentLogs)

	// Copy conflicts
	conflicts := make([]ConflictRecord, len(s.conflicts))
	copy(conflicts, s.conflicts)

	return SyncStatusSnapshot{
		TableSyncTimes:    times,
		IsConnected:       s.isConnected,
		LastHealthCheckAt: s.lastHealthCheckAt,
		LastConnectError:  s.lastConnectError,
		PendingOrders:     s.pendingOrders,
		FailedOrders:      s.failedOrders,
		LastPullAt:        s.lastPullAt,
		LastPushAt:        s.lastPushAt,
		ConsecutiveErrors: s.consecutiveErrors,
		IsSyncing:         s.isSyncing,
		RecentLogs:        logs,
		Conflicts:         conflicts,
	}
}

// SetConnected updates the connectivity state.
func (s *SyncStatus) SetConnected(connected bool, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isConnected = connected
	s.lastHealthCheckAt = time.Now().UTC().Format(time.RFC3339)
	if !connected {
		s.lastConnectError = errMsg
	} else {
		s.lastConnectError = ""
	}
}

// SetSyncing updates the syncing flag.
func (s *SyncStatus) SetSyncing(syncing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isSyncing = syncing
}

// SetOrderCounts updates the pending and failed order counts.
func (s *SyncStatus) SetOrderCounts(pending, failed int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingOrders = pending
	s.failedOrders = failed
}

// RecordPullSuccess records a successful pull and resets error count.
func (s *SyncStatus) RecordPullSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastPullAt = time.Now().UTC().Format(time.RFC3339)
	s.consecutiveErrors = 0
}

// RecordPushSuccess records a successful push.
func (s *SyncStatus) RecordPushSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastPushAt = time.Now().UTC().Format(time.RFC3339)
}

// RecordError increments the consecutive error count.
func (s *SyncStatus) RecordError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consecutiveErrors++
}

// SetTableSyncTime updates the last sync time for a specific table.
func (s *SyncStatus) SetTableSyncTime(tableName, syncTime string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tableSyncTimes[tableName] = syncTime
}

// Log adds an entry to the recent log buffer (circular, capped at maxLogEntries).
func (s *SyncStatus) Log(operation, entity, status, message string, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := SyncLogEntry{
		Time:      time.Now().UTC().Format(time.RFC3339),
		Operation: operation,
		Entity:    entity,
		Status:    status,
		Message:   message,
		Count:     count,
	}

	if len(s.recentLogs) >= maxLogEntries {
		// Shift left by 1 to make room
		copy(s.recentLogs, s.recentLogs[1:])
		s.recentLogs[len(s.recentLogs)-1] = entry
	} else {
		s.recentLogs = append(s.recentLogs, entry)
	}

	// Persist to audit log (fire-and-forget, non-blocking)
	if s.auditLogger != nil {
		go s.auditLogger.Write(operation, entity, "", operation, status, message, count)
	}
}

// RecordConflict records a detected conflict during sync push.
func (s *SyncStatus) RecordConflict(entityType, entityID, resolution string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := ConflictRecord{
		Time:       time.Now().UTC().Format(time.RFC3339),
		EntityType: entityType,
		EntityID:   entityID,
		Resolution: resolution,
	}

	if len(s.conflicts) >= maxConflictRecords {
		copy(s.conflicts, s.conflicts[1:])
		s.conflicts[len(s.conflicts)-1] = record
	} else {
		s.conflicts = append(s.conflicts, record)
	}

	// Persist to audit log
	if s.auditLogger != nil {
		go s.auditLogger.Write("push", entityType, entityID, "conflict", "conflict_"+resolution, "version mismatch, resolved via "+resolution, 0)
	}
}
