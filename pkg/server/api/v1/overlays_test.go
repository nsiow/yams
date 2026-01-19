package v1

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/overlay"
)

func newTestOverlayAPI(t *testing.T) *OverlayAPI {
	t.Helper()
	return &OverlayAPI{Store: overlay.NewMemoryStore()}
}

func TestOverlayAPI_CreateOverlay(t *testing.T) {
	api := newTestOverlayAPI(t)

	input := CreateOverlayInput{
		Name: "test-overlay",
		Principals: []entities.Principal{
			{Arn: "arn:aws:iam::123456789012:role/test"},
		},
	}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/overlays", bytes.NewReader(body))

	api.CreateOverlay(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("CreateOverlay() status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var data entities.OverlayData
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("CreateOverlay() invalid JSON: %v", err)
	}

	if data.Name != "test-overlay" {
		t.Errorf("CreateOverlay() name = %q, want %q", data.Name, "test-overlay")
	}
	if len(data.Principals) != 1 {
		t.Errorf("CreateOverlay() principals = %d, want 1", len(data.Principals))
	}
}

func TestOverlayAPI_CreateOverlay_MissingName(t *testing.T) {
	api := newTestOverlayAPI(t)

	input := CreateOverlayInput{}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/overlays", bytes.NewReader(body))

	api.CreateOverlay(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateOverlay() missing name status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestOverlayAPI_CreateOverlay_InvalidJSON(t *testing.T) {
	api := newTestOverlayAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/overlays", bytes.NewReader([]byte("invalid")))

	api.CreateOverlay(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateOverlay() invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestOverlayAPI_CreateOverlay_Duplicate(t *testing.T) {
	api := newTestOverlayAPI(t)

	// Create first overlay
	input := CreateOverlayInput{Name: "test-overlay"}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/overlays", bytes.NewReader(body))
	api.CreateOverlay(w, req)

	// Try to create duplicate
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/overlays", bytes.NewReader(body))
	api.CreateOverlay(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("CreateOverlay() duplicate status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestOverlayAPI_GetOverlay(t *testing.T) {
	api := newTestOverlayAPI(t)

	// Create overlay first
	o := entities.NewOverlay("test-overlay")
	o.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	_ = api.Store.Create(context.Background(), o)

	// Get overlay
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/overlays/"+o.ID, nil)
	req.SetPathValue("id", o.ID)

	api.GetOverlay(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetOverlay() status = %d, want %d", w.Code, http.StatusOK)
	}

	var data entities.OverlayData
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("GetOverlay() invalid JSON: %v", err)
	}

	if data.Name != "test-overlay" {
		t.Errorf("GetOverlay() name = %q, want %q", data.Name, "test-overlay")
	}
}

func TestOverlayAPI_GetOverlay_NotFound(t *testing.T) {
	api := newTestOverlayAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/overlays/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")

	api.GetOverlay(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetOverlay() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestOverlayAPI_GetOverlay_MissingID(t *testing.T) {
	api := newTestOverlayAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/overlays/", nil)
	req.SetPathValue("id", "")

	api.GetOverlay(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetOverlay() missing ID status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestOverlayAPI_UpdateOverlay(t *testing.T) {
	api := newTestOverlayAPI(t)

	// Create overlay first
	o := entities.NewOverlay("test-overlay")
	_ = api.Store.Create(context.Background(), o)

	// Update overlay
	input := UpdateOverlayInput{
		Name: "updated-overlay",
		Principals: []entities.Principal{
			{Arn: "arn:aws:iam::123456789012:role/updated"},
		},
	}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/overlays/"+o.ID, bytes.NewReader(body))
	req.SetPathValue("id", o.ID)

	api.UpdateOverlay(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateOverlay() status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var data entities.OverlayData
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("UpdateOverlay() invalid JSON: %v", err)
	}

	if data.Name != "updated-overlay" {
		t.Errorf("UpdateOverlay() name = %q, want %q", data.Name, "updated-overlay")
	}
	if len(data.Principals) != 1 {
		t.Errorf("UpdateOverlay() principals = %d, want 1", len(data.Principals))
	}
}

func TestOverlayAPI_UpdateOverlay_NotFound(t *testing.T) {
	api := newTestOverlayAPI(t)

	input := UpdateOverlayInput{Name: "updated"}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/overlays/nonexistent", bytes.NewReader(body))
	req.SetPathValue("id", "nonexistent")

	api.UpdateOverlay(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("UpdateOverlay() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestOverlayAPI_DeleteOverlay(t *testing.T) {
	api := newTestOverlayAPI(t)

	// Create overlay first
	o := entities.NewOverlay("test-overlay")
	_ = api.Store.Create(context.Background(), o)

	// Delete overlay
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/overlays/"+o.ID, nil)
	req.SetPathValue("id", o.ID)

	api.DeleteOverlay(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("DeleteOverlay() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify deletion
	exists, _ := api.Store.Exists(context.Background(), o.ID)
	if exists {
		t.Error("DeleteOverlay() overlay still exists")
	}
}

func TestOverlayAPI_DeleteOverlay_NotFound(t *testing.T) {
	api := newTestOverlayAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/overlays/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")

	api.DeleteOverlay(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("DeleteOverlay() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestOverlayAPI_ListOverlays(t *testing.T) {
	api := newTestOverlayAPI(t)

	// Create some overlays
	o1 := entities.NewOverlay("alpha-overlay")
	o1.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	_ = api.Store.Create(context.Background(), o1)

	o2 := entities.NewOverlay("beta-overlay")
	o2.Universe.PutResource(entities.Resource{Arn: "arn:aws:s3:::bucket"})
	_ = api.Store.Create(context.Background(), o2)

	// List all
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/overlays", nil)

	api.ListOverlays(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListOverlays() status = %d, want %d", w.Code, http.StatusOK)
	}

	var summaries []entities.OverlaySummary
	if err := json.Unmarshal(w.Body.Bytes(), &summaries); err != nil {
		t.Fatalf("ListOverlays() invalid JSON: %v", err)
	}

	if len(summaries) != 2 {
		t.Errorf("ListOverlays() count = %d, want 2", len(summaries))
	}
}

func TestOverlayAPI_ListOverlays_WithQuery(t *testing.T) {
	api := newTestOverlayAPI(t)

	// Create overlays
	_ = api.Store.Create(context.Background(), entities.NewOverlay("alpha-overlay"))
	_ = api.Store.Create(context.Background(), entities.NewOverlay("beta-test"))

	// Search
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/overlays?q=alpha", nil)

	api.ListOverlays(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListOverlays() status = %d, want %d", w.Code, http.StatusOK)
	}

	var summaries []entities.OverlaySummary
	if err := json.Unmarshal(w.Body.Bytes(), &summaries); err != nil {
		t.Fatalf("ListOverlays() invalid JSON: %v", err)
	}

	if len(summaries) != 1 {
		t.Errorf("ListOverlays() with query count = %d, want 1", len(summaries))
	}
}
