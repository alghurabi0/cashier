package service

import (
	"fmt"
	"time"
)

// ConflictError is returned when an optimistic concurrency check fails.
// The client's expected version (updated_at) does not match the server's current version,
// meaning another write occurred between the client's read and write.
type ConflictError struct {
	EntityType    string    `json:"entity_type"`
	EntityID      string    `json:"entity_id"`
	ServerVersion time.Time `json:"server_version"` // Current updated_at on the server
	ClientVersion time.Time `json:"client_version"` // The version the client expected
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf(
		"conflict: %s/%s was modified (server=%s, client expected=%s)",
		e.EntityType, e.EntityID,
		e.ServerVersion.Format(time.RFC3339),
		e.ClientVersion.Format(time.RFC3339),
	)
}
