package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim"
)

func newTestAPI(t *testing.T) *API {
	t.Helper()
	simulator, err := sim.NewSimulator()
	if err != nil {
		t.Fatalf("Failed to create simulator: %v", err)
	}
	return &API{Simulator: simulator}
}

func newTestAPIWithData(t *testing.T) *API {
	t.Helper()
	simulator, err := sim.NewSimulator()
	if err != nil {
		t.Fatalf("Failed to create simulator: %v", err)
	}

	// Add test data
	account := entities.Account{
		Id:   "123456789012",
		Name: "TestAccount",
	}
	simulator.Universe.PutAccount(account)

	principal := entities.Principal{
		Arn:       "arn:aws:iam::123456789012:user/testuser",
		AccountId: "123456789012",
	}
	simulator.Universe.PutPrincipal(principal)

	resource := entities.Resource{
		Arn:  "arn:aws:s3:::test-bucket",
		Type: "AWS::S3::Bucket",
	}
	simulator.Universe.PutResource(resource)

	group := entities.Group{
		Arn:       "arn:aws:iam::123456789012:group/testgroup",
		AccountId: "123456789012",
	}
	simulator.Universe.PutGroup(group)

	policy := entities.ManagedPolicy{
		Arn:       "arn:aws:iam::123456789012:policy/testpolicy",
		AccountId: "123456789012",
	}
	simulator.Universe.PutPolicy(policy)

	return &API{Simulator: simulator}
}

// -------------------------------------------------------------------------------------------------
// List Tests
// -------------------------------------------------------------------------------------------------

func TestAPI_ListAccounts(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/accounts", nil)

	api.ListAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts []string
	if err := json.Unmarshal(w.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("ListAccounts() invalid JSON: %v", err)
	}
}

func TestAPI_ListGroups(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/groups", nil)

	api.ListGroups(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListGroups() status = %d, want %d", w.Code, http.StatusOK)
	}

	var groups []string
	if err := json.Unmarshal(w.Body.Bytes(), &groups); err != nil {
		t.Fatalf("ListGroups() invalid JSON: %v", err)
	}
}

func TestAPI_ListPolicies(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/policies", nil)

	api.ListPolicies(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListPolicies() status = %d, want %d", w.Code, http.StatusOK)
	}

	var policies []string
	if err := json.Unmarshal(w.Body.Bytes(), &policies); err != nil {
		t.Fatalf("ListPolicies() invalid JSON: %v", err)
	}
}

func TestAPI_ListPrincipals(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/principals", nil)

	api.ListPrincipals(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListPrincipals() status = %d, want %d", w.Code, http.StatusOK)
	}

	var principals []string
	if err := json.Unmarshal(w.Body.Bytes(), &principals); err != nil {
		t.Fatalf("ListPrincipals() invalid JSON: %v", err)
	}
}

func TestAPI_ListResources(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/resources", nil)

	api.ListResources(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListResources() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resources []string
	if err := json.Unmarshal(w.Body.Bytes(), &resources); err != nil {
		t.Fatalf("ListResources() invalid JSON: %v", err)
	}
}

func TestAPI_ListActions(t *testing.T) {
	api := newTestAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/actions", nil)

	api.ListActions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListActions() status = %d, want %d", w.Code, http.StatusOK)
	}

	var actions []string
	if err := json.Unmarshal(w.Body.Bytes(), &actions); err != nil {
		t.Fatalf("ListActions() invalid JSON: %v", err)
	}

	if len(actions) == 0 {
		t.Error("ListActions() returned empty list")
	}
}

// -------------------------------------------------------------------------------------------------
// Get Tests
// -------------------------------------------------------------------------------------------------

func TestAPI_GetAccount(t *testing.T) {
	api := newTestAPIWithData(t)

	// Test valid account - Account uses Id as the key
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/accounts/123456789012", nil)
	req.SetPathValue("key", "123456789012")

	api.GetAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetAccount() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Test missing key
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/accounts/", nil)
	req.SetPathValue("key", "")

	api.GetAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetAccount() with empty key status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Test not found
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/accounts/notfound", nil)
	req.SetPathValue("key", "notfound")

	api.GetAccount(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetAccount() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAPI_GetGroup(t *testing.T) {
	api := newTestAPIWithData(t)

	// Test valid group
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/groups/arn:aws:iam::123456789012:group/testgroup", nil)
	req.SetPathValue("key", "arn:aws:iam::123456789012:group/testgroup")

	api.GetGroup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetGroup() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_GetPolicy(t *testing.T) {
	api := newTestAPIWithData(t)

	// Test valid policy
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/policies/arn:aws:iam::123456789012:policy/testpolicy", nil)
	req.SetPathValue("key", "arn:aws:iam::123456789012:policy/testpolicy")

	api.GetPolicy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetPolicy() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_GetPrincipal(t *testing.T) {
	api := newTestAPIWithData(t)

	// Test valid principal
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/principals/arn:aws:iam::123456789012:user/testuser", nil)
	req.SetPathValue("key", "arn:aws:iam::123456789012:user/testuser")

	api.GetPrincipal(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetPrincipal() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_GetResource(t *testing.T) {
	api := newTestAPIWithData(t)

	// Test valid resource
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/resources/arn:aws:s3:::test-bucket", nil)
	req.SetPathValue("key", "arn:aws:s3:::test-bucket")

	api.GetResource(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetResource() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_GetAction(t *testing.T) {
	api := newTestAPI(t)

	// Test valid action
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/actions/s3:GetObject", nil)
	req.SetPathValue("key", "s3:GetObject")

	api.GetAction(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetAction() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Test missing key
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/actions/", nil)
	req.SetPathValue("key", "")

	api.GetAction(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetAction() with empty key status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Test unknown action
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/actions/unknown:Action", nil)
	req.SetPathValue("key", "unknown:Action")

	api.GetAction(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetAction() unknown action status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// -------------------------------------------------------------------------------------------------
// Search Tests
// -------------------------------------------------------------------------------------------------

func TestAPI_SearchAccounts(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/accounts/search/123", nil)
	req.SetPathValue("search", "123")

	api.SearchAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SearchAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts []string
	if err := json.Unmarshal(w.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("SearchAccounts() invalid JSON: %v", err)
	}
}

func TestAPI_SearchGroups(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/groups/search/test", nil)
	req.SetPathValue("search", "test")

	api.SearchGroups(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SearchGroups() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_SearchPolicies(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/policies/search/test", nil)
	req.SetPathValue("search", "test")

	api.SearchPolicies(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SearchPolicies() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_SearchPrincipals(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/principals/search/test", nil)
	req.SetPathValue("search", "test")

	api.SearchPrincipals(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SearchPrincipals() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_SearchResources(t *testing.T) {
	api := newTestAPIWithData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/resources/search/bucket", nil)
	req.SetPathValue("search", "bucket")

	api.SearchResources(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SearchResources() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_SearchActions(t *testing.T) {
	api := newTestAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/actions/search/s3", nil)
	req.SetPathValue("search", "s3")

	api.SearchActions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SearchActions() status = %d, want %d", w.Code, http.StatusOK)
	}

	var actions []string
	if err := json.Unmarshal(w.Body.Bytes(), &actions); err != nil {
		t.Fatalf("SearchActions() invalid JSON: %v", err)
	}
}

// -------------------------------------------------------------------------------------------------
// Simulation Tests
// -------------------------------------------------------------------------------------------------

func TestAPI_SimRun(t *testing.T) {
	api := newTestAPIWithData(t)

	tests := []struct {
		name       string
		input      SimInput
		wantStatus int
	}{
		{
			name: "valid simulation",
			input: SimInput{
				Principal: "arn:aws:iam::123456789012:user/testuser",
				Action:    "s3:ListBucket",
				Resource:  "arn:aws:s3:::test-bucket",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "missing principal",
			input: SimInput{
				Action:   "s3:ListBucket",
				Resource: "arn:aws:s3:::test-bucket",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing action",
			input: SimInput{
				Principal: "arn:aws:iam::123456789012:user/testuser",
				Resource:  "arn:aws:s3:::test-bucket",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "with explain",
			input: SimInput{
				Principal: "arn:aws:iam::123456789012:user/testuser",
				Action:    "s3:ListBucket",
				Resource:  "arn:aws:s3:::test-bucket",
				Explain:   true,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "with trace",
			input: SimInput{
				Principal: "arn:aws:iam::123456789012:user/testuser",
				Action:    "s3:ListBucket",
				Resource:  "arn:aws:s3:::test-bucket",
				Trace:     true,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "with fuzzy",
			input: SimInput{
				Principal: "arn:aws:iam::123456789012:user/testuser",
				Action:    "s3:ListBucket",
				Resource:  "arn:aws:s3:::test-bucket",
				Fuzzy:     true,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "with context",
			input: SimInput{
				Principal: "arn:aws:iam::123456789012:user/testuser",
				Action:    "s3:ListBucket",
				Resource:  "arn:aws:s3:::test-bucket",
				Context:   map[string]string{"aws:SourceIp": "10.0.0.1"},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/sim", bytes.NewReader(body))

			api.SimRun(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("SimRun() status = %d, want %d, body = %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantStatus == http.StatusOK {
				var out SimOutput
				if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
					t.Fatalf("SimRun() invalid JSON: %v", err)
				}
				if out.Result != "ALLOW" && out.Result != "DENY" {
					t.Errorf("SimRun() result = %s, want ALLOW or DENY", out.Result)
				}
			}
		})
	}
}

func TestAPI_SimRun_InvalidJSON(t *testing.T) {
	api := newTestAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/sim", bytes.NewReader([]byte("invalid json")))

	api.SimRun(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("SimRun() with invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAPI_WhichPrincipals(t *testing.T) {
	api := newTestAPIWithData(t)

	// Valid request
	input := WhichPrincipalsInput{
		Action:   "s3:GetObject",
		Resource: "arn:aws:s3:::test-bucket",
	}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/sim/whichPrincipals", bytes.NewReader(body))

	api.WhichPrincipals(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("WhichPrincipals() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Invalid JSON
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/sim/whichPrincipals", bytes.NewReader([]byte("invalid")))

	api.WhichPrincipals(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("WhichPrincipals() with invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Missing required action
	input2 := WhichPrincipalsInput{
		Resource: "arn:aws:s3:::test-bucket",
	}
	body, _ = json.Marshal(input2)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/sim/whichPrincipals", bytes.NewReader(body))

	api.WhichPrincipals(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("WhichPrincipals() missing action status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAPI_WhichActions(t *testing.T) {
	api := newTestAPIWithData(t)

	// Valid request
	input := WhichActionsInput{
		Principal: "arn:aws:iam::123456789012:user/testuser",
		Resource:  "arn:aws:s3:::test-bucket",
	}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/sim/whichActions", bytes.NewReader(body))

	api.WhichActions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("WhichActions() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Invalid JSON
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/sim/whichActions", bytes.NewReader([]byte("invalid")))

	api.WhichActions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("WhichActions() with invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAPI_WhichResources(t *testing.T) {
	api := newTestAPIWithData(t)

	// Valid request
	input := WhichResourcesInput{
		Principal: "arn:aws:iam::123456789012:user/testuser",
		Action:    "s3:ListBucket",
	}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/sim/whichResources", bytes.NewReader(body))

	api.WhichResources(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("WhichResources() status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Invalid JSON
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/sim/whichResources", bytes.NewReader([]byte("invalid")))

	api.WhichResources(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("WhichResources() with invalid JSON status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Missing required principal
	input2 := WhichResourcesInput{
		Action: "s3:ListBucket",
	}
	body, _ = json.Marshal(input2)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/sim/whichResources", bytes.NewReader(body))

	api.WhichResources(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("WhichResources() missing principal status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// -------------------------------------------------------------------------------------------------
// Overlay Tests
// -------------------------------------------------------------------------------------------------

func TestOverlay_Universe(t *testing.T) {
	overlay := Overlay{
		Accounts: []entities.Account{
			{Id: "123456789012", Name: "TestAccount"},
		},
		Groups: []entities.Group{
			{Arn: "arn:aws:iam::123456789012:group/testgroup"},
		},
		Policies: []entities.ManagedPolicy{
			{Arn: "arn:aws:iam::123456789012:policy/testpolicy"},
		},
		Principals: []entities.Principal{
			{Arn: "arn:aws:iam::123456789012:user/testuser"},
		},
		Resources: []entities.Resource{
			{Arn: "arn:aws:s3:::test-bucket"},
		},
	}

	universe := overlay.Universe()

	if universe == nil {
		t.Fatal("Universe() returned nil")
	}
}

func TestGet_WithFreeze(t *testing.T) {
	api := newTestAPIWithData(t)

	// Test freeze suffix
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/principals/arn:aws:iam::123456789012:user/testuser/freeze", nil)
	req.SetPathValue("key", "arn:aws:iam::123456789012:user/testuser/freeze")

	api.GetPrincipal(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetPrincipal() with freeze status = %d, want %d", w.Code, http.StatusOK)
	}
}


func TestSimRun_Allow(t *testing.T) {
	api := newTestAPIWithData(t)

	// Add a principal with an inline policy that allows s3:ListBucket
	principal := entities.Principal{
		Arn:       "arn:aws:iam::123456789012:user/adminuser",
		AccountId: "123456789012",
		InlinePolicies: []policy.Policy{
			{
				Statement: []policy.Statement{
					{
						Effect:   policy.EFFECT_ALLOW,
						Action:   []string{"s3:ListBucket"},
						Resource: []string{"arn:aws:s3:::allowbucket"},
					},
				},
			},
		},
	}
	api.Simulator.Universe.PutPrincipal(principal)

	// Add a resource that's in the same account
	resource := entities.Resource{
		Arn:       "arn:aws:s3:::allowbucket",
		Type:      "AWS::S3::Bucket",
		AccountId: "123456789012",
	}
	api.Simulator.Universe.PutResource(resource)

	// SimRun should return ALLOW
	input := SimInput{
		Principal: "arn:aws:iam::123456789012:user/adminuser",
		Action:    "s3:ListBucket",
		Resource:  "arn:aws:s3:::allowbucket",
	}
	body, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/sim/run", bytes.NewReader(body))

	api.SimRun(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SimRun() allow status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var out SimOutput
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("SimRun() invalid JSON: %v", err)
	}
	if out.Result != "ALLOW" {
		t.Errorf("SimRun() result = %s, want ALLOW", out.Result)
	}
}

