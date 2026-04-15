package overlay

import (
	"context"
	"strings"
	"sync"

	"github.com/nsiow/yams/pkg/entities"
)

// MemoryStore is an in-memory implementation of the Store interface.
// It is safe for concurrent use.
type MemoryStore struct {
	mu       sync.RWMutex
	overlays map[string]*entities.Overlay // keyed by ID
	byName   map[string]string            // name -> ID mapping
}

// NewMemoryStore creates a new in-memory overlay store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		overlays: make(map[string]*entities.Overlay),
		byName:   make(map[string]string),
	}
}

// Create stores a new overlay in memory.
func (s *MemoryStore) Create(ctx context.Context, overlay *entities.Overlay) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.overlays[overlay.ID]; exists {
		return ErrAlreadyExists
	}

	// Store a copy to prevent external mutations
	s.overlays[overlay.ID] = s.clone(overlay)
	s.byName[overlay.Name] = overlay.ID

	return nil
}

// Get retrieves an overlay by ID.
func (s *MemoryStore) Get(ctx context.Context, id string) (*entities.Overlay, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	overlay, exists := s.overlays[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to prevent external mutations
	return s.clone(overlay), nil
}

// GetByName retrieves an overlay by name.
func (s *MemoryStore) GetByName(ctx context.Context, name string) (*entities.Overlay, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.byName[name]
	if !exists {
		return nil, ErrNotFound
	}

	overlay, exists := s.overlays[id]
	if !exists {
		return nil, ErrNotFound
	}

	return s.clone(overlay), nil
}

// Update replaces an existing overlay.
func (s *MemoryStore) Update(ctx context.Context, overlay *entities.Overlay) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.overlays[overlay.ID]
	if !exists {
		return ErrNotFound
	}

	// Remove old name mapping if name changed
	if existing.Name != overlay.Name {
		delete(s.byName, existing.Name)
		s.byName[overlay.Name] = overlay.ID
	}

	s.overlays[overlay.ID] = s.clone(overlay)
	return nil
}

// Delete removes an overlay by ID.
func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	overlay, exists := s.overlays[id]
	if !exists {
		return ErrNotFound
	}

	delete(s.byName, overlay.Name)
	delete(s.overlays, id)
	return nil
}

// List returns summaries of all overlays, optionally filtered by query.
func (s *MemoryStore) List(ctx context.Context, query string) ([]entities.OverlaySummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	summaries := make([]entities.OverlaySummary, 0, len(s.overlays))

	for _, overlay := range s.overlays {
		// Filter by query if provided
		if query != "" && !strings.Contains(strings.ToLower(overlay.Name), query) {
			continue
		}
		summaries = append(summaries, overlay.Summary())
	}

	return summaries, nil
}

// Exists checks if an overlay with the given ID exists.
func (s *MemoryStore) Exists(ctx context.Context, id string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.overlays[id]
	return exists, nil
}

// clone creates a deep copy of an overlay using the serialization methods.
func (s *MemoryStore) clone(o *entities.Overlay) *entities.Overlay {
	data := o.ToData()
	return entities.FromData(data)
}

// Size returns the number of overlays in the store.
func (s *MemoryStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.overlays)
}
