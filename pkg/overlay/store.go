package overlay

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/entities"
)

// Common errors returned by OverlayStore implementations.
var (
	ErrNotFound      = errors.New("overlay not found")
	ErrAlreadyExists = errors.New("overlay already exists")
)

// Store defines the interface for overlay storage backends.
// All operations are context-aware to support timeouts and cancellation.
type Store interface {
	// Create stores a new overlay. Returns ErrAlreadyExists if an overlay
	// with the same ID already exists.
	Create(ctx context.Context, overlay *entities.Overlay) error

	// Get retrieves an overlay by ID. Returns ErrNotFound if not found.
	Get(ctx context.Context, id string) (*entities.Overlay, error)

	// GetByName retrieves an overlay by name. Returns ErrNotFound if not found.
	GetByName(ctx context.Context, name string) (*entities.Overlay, error)

	// Update replaces an existing overlay. Returns ErrNotFound if not found.
	Update(ctx context.Context, overlay *entities.Overlay) error

	// Delete removes an overlay by ID. Returns ErrNotFound if not found.
	Delete(ctx context.Context, id string) error

	// List returns summaries of all overlays, optionally filtered by a search query.
	// If query is empty, all overlays are returned.
	List(ctx context.Context, query string) ([]entities.OverlaySummary, error)

	// Exists checks if an overlay with the given ID exists.
	Exists(ctx context.Context, id string) (bool, error)
}

// NewStore creates a Store based on the provided spec string.
// Supported formats:
//   - "" or "memory": in-memory store
//   - "ddb://<table-name>": DynamoDB store using the specified table
func NewStore(spec string) (Store, error) {
	if spec == "" || spec == "memory" {
		return NewMemoryStore(), nil
	}

	if strings.HasPrefix(spec, "ddb://") {
		table := strings.TrimPrefix(spec, "ddb://")
		if table == "" {
			return nil, fmt.Errorf("invalid overlay spec: table name required for ddb://")
		}
		return NewDynamoDBStore(table)
	}

	return nil, fmt.Errorf("unknown overlay store spec: %s", spec)
}
