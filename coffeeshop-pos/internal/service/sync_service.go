package service

import (
	"fmt"
	"log/slog"

	posSync "coffeeshop-pos/internal/sync"

	"github.com/jmoiron/sqlx"
)

// SyncService is a Wails-bound service that exposes sync engine health
// and controls to the frontend dashboard.
type SyncService struct {
	syncWorker *posSync.Worker
	db         *sqlx.DB
}

// NewSyncService creates a new SyncService.
func NewSyncService(syncWorker *posSync.Worker, db *sqlx.DB) *SyncService {
	return &SyncService{
		syncWorker: syncWorker,
		db:         db,
	}
}

// GetSyncStatus returns the full sync health snapshot for the dashboard.
func (s *SyncService) GetSyncStatus() *posSync.SyncStatusSnapshot {
	snap := s.syncWorker.Status.Snapshot()
	return &snap
}

// GetSyncLogs returns recent sync log entries.
func (s *SyncService) GetSyncLogs() []posSync.SyncLogEntry {
	snap := s.syncWorker.Status.Snapshot()
	return snap.RecentLogs
}

// GetPendingOrderCount returns count of unsynced orders (synced=0).
func (s *SyncService) GetPendingOrderCount() int {
	var count int
	s.db.Get(&count, `SELECT COUNT(*) FROM orders WHERE synced = 0`)
	return count
}

// GetFailedOrderCount returns count of dead-lettered orders (synced=2).
func (s *SyncService) GetFailedOrderCount() int {
	var count int
	s.db.Get(&count, `SELECT COUNT(*) FROM orders WHERE synced = 2`)
	return count
}

// GetFailedOrders returns the list of dead-lettered orders with error details.
func (s *SyncService) GetFailedOrders() ([]posSync.FailedOrderInfo, error) {
	return s.syncWorker.GetFailedOrders()
}

// TriggerSync manually triggers a pull + push cycle.
func (s *SyncService) TriggerSync() error {
	slog.Info("sync-service: manual sync triggered")
	if err := s.syncWorker.PullAll(); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}
	if err := s.syncWorker.PushOrders(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}
	s.syncWorker.Status.RecordPullSuccess()
	s.syncWorker.Status.RecordPushSuccess()
	return nil
}

// TriggerFullSync clears all sync timestamps and triggers a full re-sync.
func (s *SyncService) TriggerFullSync() error {
	slog.Info("sync-service: full sync triggered — clearing timestamps")
	if err := s.syncWorker.ResetSyncMeta(); err != nil {
		return fmt.Errorf("reset sync meta failed: %w", err)
	}
	if err := s.syncWorker.PullAll(); err != nil {
		return fmt.Errorf("full pull failed: %w", err)
	}
	if err := s.syncWorker.PushOrders(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}
	s.syncWorker.Status.RecordPullSuccess()
	s.syncWorker.Status.RecordPushSuccess()
	return nil
}

// RetryFailedOrders resets dead-lettered orders (synced=2) back to pending (synced=0).
func (s *SyncService) RetryFailedOrders() (int, error) {
	return s.syncWorker.RetryFailedOrders()
}

// ResetSyncState clears all sync_meta timestamps. The next sync cycle will be a full sync.
func (s *SyncService) ResetSyncState() error {
	return s.syncWorker.ResetSyncMeta()
}

// GetAuditLog returns the most recent N sync audit log entries from persistent storage.
// This is useful for debugging sync issues — entries survive app restarts.
func (s *SyncService) GetAuditLog(limit int) ([]posSync.AuditLogEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	logger := posSync.NewAuditLogger(s.db)
	return logger.QueryRecent(limit)
}

// GetAuditConflicts returns only conflict entries from the audit log.
func (s *SyncService) GetAuditConflicts(limit int) ([]posSync.AuditLogEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	logger := posSync.NewAuditLogger(s.db)
	return logger.QueryConflicts(limit)
}
