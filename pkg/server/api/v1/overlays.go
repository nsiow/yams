package v1

import (
	"fmt"
	"net/http"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/overlay"
	"github.com/nsiow/yams/pkg/server/httputil"
)

// OverlayAPI handles overlay storage operations.
type OverlayAPI struct {
	Store overlay.Store
}

// CreateOverlayInput is the request body for creating an overlay.
type CreateOverlayInput struct {
	Name       string                   `json:"name"`
	Accounts   []entities.Account       `json:"accounts,omitempty"`
	Groups     []entities.Group         `json:"groups,omitempty"`
	Policies   []entities.ManagedPolicy `json:"policies,omitempty"`
	Principals []entities.Principal     `json:"principals,omitempty"`
	Resources  []entities.Resource      `json:"resources,omitempty"`
}

// UpdateOverlayInput is the request body for updating an overlay.
type UpdateOverlayInput struct {
	Name       string                   `json:"name,omitempty"`
	Accounts   []entities.Account       `json:"accounts,omitempty"`
	Groups     []entities.Group         `json:"groups,omitempty"`
	Policies   []entities.ManagedPolicy `json:"policies,omitempty"`
	Principals []entities.Principal     `json:"principals,omitempty"`
	Resources  []entities.Resource      `json:"resources,omitempty"`
}

// ListOverlays returns summaries of all overlays, optionally filtered by query.
// GET /api/v1/overlays?q=search
func (api *OverlayAPI) ListOverlays(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query().Get("q")

	summaries, err := api.Store.List(req.Context(), query)
	if err != nil {
		httputil.ServerError(w, req, fmt.Errorf("failed to list overlays: %v", err))
		return
	}

	httputil.WriteJsonResponse(w, req, summaries)
}

// GetOverlay retrieves an overlay by ID.
// GET /api/v1/overlays/{id}
func (api *OverlayAPI) GetOverlay(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		httputil.ClientError(w, req, fmt.Errorf("missing overlay ID"))
		return
	}

	o, err := api.Store.Get(req.Context(), id)
	if err == overlay.ErrNotFound {
		httputil.Error(w, req, http.StatusNotFound, fmt.Errorf("overlay not found: %s", id))
		return
	}
	if err != nil {
		httputil.ServerError(w, req, fmt.Errorf("failed to get overlay: %v", err))
		return
	}

	httputil.WriteJsonResponse(w, req, o.ToData())
}

// CreateOverlay creates a new overlay.
// POST /api/v1/overlays
func (api *OverlayAPI) CreateOverlay(w http.ResponseWriter, req *http.Request) {
	var input CreateOverlayInput
	decoder := json.ConfigDefault.NewDecoder(req.Body)
	if err := decoder.Decode(&input); err != nil {
		httputil.ClientError(w, req, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	if input.Name == "" {
		httputil.ClientError(w, req, fmt.Errorf("missing required field 'name'"))
		return
	}

	o := entities.NewOverlay(input.Name)

	// Populate entities
	for _, a := range input.Accounts {
		o.Universe.PutAccount(a)
	}
	for _, g := range input.Groups {
		o.Universe.PutGroup(g)
	}
	for _, p := range input.Policies {
		o.Universe.PutPolicy(p)
	}
	for _, p := range input.Principals {
		o.Universe.PutPrincipal(p)
	}
	for _, r := range input.Resources {
		o.Universe.PutResource(r)
	}

	if err := api.Store.Create(req.Context(), o); err != nil {
		if err == overlay.ErrAlreadyExists {
			httputil.Error(w, req, http.StatusConflict, fmt.Errorf("overlay with this name already exists"))
			return
		}
		httputil.ServerError(w, req, fmt.Errorf("failed to create overlay: %v", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	httputil.WriteJsonResponse(w, req, o.ToData())
}

// UpdateOverlay updates an existing overlay.
// PUT /api/v1/overlays/{id}
func (api *OverlayAPI) UpdateOverlay(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		httputil.ClientError(w, req, fmt.Errorf("missing overlay ID"))
		return
	}

	// Get existing overlay
	existing, err := api.Store.Get(req.Context(), id)
	if err == overlay.ErrNotFound {
		httputil.Error(w, req, http.StatusNotFound, fmt.Errorf("overlay not found: %s", id))
		return
	}
	if err != nil {
		httputil.ServerError(w, req, fmt.Errorf("failed to get overlay: %v", err))
		return
	}

	var input UpdateOverlayInput
	decoder := json.ConfigDefault.NewDecoder(req.Body)
	if err := decoder.Decode(&input); err != nil {
		httputil.ClientError(w, req, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	// Update name if provided
	if input.Name != "" {
		existing.Name = input.Name
	}

	// Replace universe with new entities
	existing.Universe = entities.NewUniverse()
	for _, a := range input.Accounts {
		existing.Universe.PutAccount(a)
	}
	for _, g := range input.Groups {
		existing.Universe.PutGroup(g)
	}
	for _, p := range input.Policies {
		existing.Universe.PutPolicy(p)
	}
	for _, p := range input.Principals {
		existing.Universe.PutPrincipal(p)
	}
	for _, r := range input.Resources {
		existing.Universe.PutResource(r)
	}

	if err := api.Store.Update(req.Context(), existing); err != nil {
		httputil.ServerError(w, req, fmt.Errorf("failed to update overlay: %v", err))
		return
	}

	httputil.WriteJsonResponse(w, req, existing.ToData())
}

// DeleteOverlay deletes an overlay by ID.
// DELETE /api/v1/overlays/{id}
func (api *OverlayAPI) DeleteOverlay(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		httputil.ClientError(w, req, fmt.Errorf("missing overlay ID"))
		return
	}

	err := api.Store.Delete(req.Context(), id)
	if err == overlay.ErrNotFound {
		httputil.Error(w, req, http.StatusNotFound, fmt.Errorf("overlay not found: %s", id))
		return
	}
	if err != nil {
		httputil.ServerError(w, req, fmt.Errorf("failed to delete overlay: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
