package overlay

import (
	"context"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
)

func TestMemoryStore_Create(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	overlay.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})

	// Create should succeed
	err := store.Create(ctx, overlay)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Create same overlay again should fail
	err = store.Create(ctx, overlay)
	if err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}

	// Verify store size
	if store.Size() != 1 {
		t.Errorf("expected size 1, got %d", store.Size())
	}
}

func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	overlay.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})

	_ = store.Create(ctx, overlay)

	// Get should succeed
	retrieved, err := store.Get(ctx, overlay.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != overlay.Name {
		t.Errorf("expected name %q, got %q", overlay.Name, retrieved.Name)
	}
	if retrieved.ID != overlay.ID {
		t.Errorf("expected ID %q, got %q", overlay.ID, retrieved.ID)
	}
	if retrieved.NumPrincipals() != 1 {
		t.Errorf("expected 1 principal, got %d", retrieved.NumPrincipals())
	}

	// Get non-existent should fail
	_, err = store.Get(ctx, "non-existent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_GetByName(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	_ = store.Create(ctx, overlay)

	// GetByName should succeed
	retrieved, err := store.GetByName(ctx, "test-overlay")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if retrieved.ID != overlay.ID {
		t.Errorf("expected ID %q, got %q", overlay.ID, retrieved.ID)
	}

	// GetByName non-existent should fail
	_, err = store.GetByName(ctx, "non-existent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_Update(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	_ = store.Create(ctx, overlay)

	// Add entities and update
	overlay.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	err := store.Update(ctx, overlay)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, _ := store.Get(ctx, overlay.ID)
	if retrieved.NumPrincipals() != 1 {
		t.Errorf("expected 1 principal after update, got %d", retrieved.NumPrincipals())
	}

	// Update non-existent should fail
	nonExistent := entities.NewOverlay("non-existent")
	err = store.Update(ctx, nonExistent)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_Update_NameChange(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	overlay := entities.NewOverlay("original-name")
	_ = store.Create(ctx, overlay)

	// Change name (need to create new overlay with same ID)
	updated := &entities.Overlay{
		Name:      "new-name",
		ID:        overlay.ID,
		CreatedAt: overlay.CreatedAt,
		Universe:  entities.NewUniverse(),
	}

	err := store.Update(ctx, updated)
	if err != nil {
		t.Fatalf("Update with name change failed: %v", err)
	}

	// Old name should not work
	_, err = store.GetByName(ctx, "original-name")
	if err != ErrNotFound {
		t.Error("old name should not be found")
	}

	// New name should work
	retrieved, err := store.GetByName(ctx, "new-name")
	if err != nil {
		t.Fatalf("GetByName with new name failed: %v", err)
	}
	if retrieved.ID != overlay.ID {
		t.Error("ID should remain the same")
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	_ = store.Create(ctx, overlay)

	// Delete should succeed
	err := store.Delete(ctx, overlay.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = store.Get(ctx, overlay.ID)
	if err != ErrNotFound {
		t.Error("overlay should be deleted")
	}

	// Name mapping should also be removed
	_, err = store.GetByName(ctx, "test-overlay")
	if err != ErrNotFound {
		t.Error("name mapping should be removed")
	}

	// Delete non-existent should fail
	err = store.Delete(ctx, "non-existent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_List(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create multiple overlays
	overlay1 := entities.NewOverlay("alpha-overlay")
	overlay1.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	_ = store.Create(ctx, overlay1)

	overlay2 := entities.NewOverlay("beta-overlay")
	overlay2.Universe.PutResource(entities.Resource{Arn: "arn:aws:s3:::test-bucket"})
	_ = store.Create(ctx, overlay2)

	overlay3 := entities.NewOverlay("gamma-test")
	_ = store.Create(ctx, overlay3)

	// List all
	summaries, err := store.List(ctx, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(summaries) != 3 {
		t.Errorf("expected 3 summaries, got %d", len(summaries))
	}

	// List with query
	summaries, err = store.List(ctx, "overlay")
	if err != nil {
		t.Fatalf("List with query failed: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("expected 2 summaries matching 'overlay', got %d", len(summaries))
	}

	// Case-insensitive search
	summaries, err = store.List(ctx, "ALPHA")
	if err != nil {
		t.Fatalf("List with uppercase query failed: %v", err)
	}
	if len(summaries) != 1 {
		t.Errorf("expected 1 summary matching 'ALPHA', got %d", len(summaries))
	}
}

func TestMemoryStore_Exists(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	_ = store.Create(ctx, overlay)

	// Exists should return true
	exists, err := store.Exists(ctx, overlay.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected overlay to exist")
	}

	// Non-existent should return false
	exists, err = store.Exists(ctx, "non-existent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected overlay to not exist")
	}
}

func TestMemoryStore_Isolation(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create overlay
	overlay := entities.NewOverlay("test-overlay")
	overlay.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	_ = store.Create(ctx, overlay)

	// Modify original after create
	overlay.Universe.PutResource(entities.Resource{Arn: "arn:aws:s3:::bucket"})

	// Retrieved overlay should not have the resource
	retrieved, _ := store.Get(ctx, overlay.ID)
	if retrieved.NumResources() != 0 {
		t.Error("store should be isolated from external mutations after Create")
	}

	// Modify retrieved overlay
	retrieved.Universe.PutResource(entities.Resource{Arn: "arn:aws:s3:::bucket2"})

	// Get again, should not have the resource
	retrieved2, _ := store.Get(ctx, overlay.ID)
	if retrieved2.NumResources() != 0 {
		t.Error("store should be isolated from external mutations after Get")
	}
}
