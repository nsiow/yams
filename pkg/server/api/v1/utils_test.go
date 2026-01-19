package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/sim"
)

func newTestAPIWithUtilsData(t *testing.T) *API {
	t.Helper()
	simulator, err := sim.NewSimulator()
	if err != nil {
		t.Fatalf("Failed to create simulator: %v", err)
	}

	// Add test accounts
	simulator.Universe.PutAccount(entities.Account{
		Id:   "111111111111",
		Name: "Account One",
	})
	simulator.Universe.PutAccount(entities.Account{
		Id:   "222222222222",
		Name: "Account Two",
	})

	// Add S3 bucket (global namespace resource - no account in ARN)
	simulator.Universe.PutResource(entities.Resource{
		Arn:       "arn:aws:s3:::my-test-bucket",
		Type:      "AWS::S3::Bucket",
		AccountId: "111111111111",
	})

	// Add another S3 bucket with different account
	simulator.Universe.PutResource(entities.Resource{
		Arn:       "arn:aws:s3:::another-bucket",
		Type:      "AWS::S3::Bucket",
		AccountId: "222222222222",
	})

	// Add EC2 instance (regional resource - has account in ARN)
	simulator.Universe.PutResource(entities.Resource{
		Arn:       "arn:aws:ec2:us-east-1:111111111111:instance/i-1234567890abcdef0",
		Type:      "AWS::EC2::Instance",
		AccountId: "111111111111",
	})

	return &API{Simulator: simulator}
}

func TestAPI_UtilAccountNames(t *testing.T) {
	api := newTestAPIWithUtilsData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/utils/accounts/names", nil)

	api.UtilAccountNames(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UtilAccountNames() status = %d, want %d", w.Code, http.StatusOK)
	}

	var names map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &names); err != nil {
		t.Fatalf("UtilAccountNames() invalid JSON: %v", err)
	}

	// Verify accounts are present
	if names["111111111111"] != "Account One" {
		t.Errorf("UtilAccountNames() missing or incorrect account 111111111111, got %q", names["111111111111"])
	}
	if names["222222222222"] != "Account Two" {
		t.Errorf("UtilAccountNames() missing or incorrect account 222222222222, got %q", names["222222222222"])
	}
}

func TestAPI_UtilAccountNames_Empty(t *testing.T) {
	api := newTestAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/utils/accounts/names", nil)

	api.UtilAccountNames(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UtilAccountNames() status = %d, want %d", w.Code, http.StatusOK)
	}

	var names map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &names); err != nil {
		t.Fatalf("UtilAccountNames() invalid JSON: %v", err)
	}

	if len(names) != 0 {
		t.Errorf("UtilAccountNames() expected empty map, got %d entries", len(names))
	}
}

func TestAPI_UtilResourceAccounts(t *testing.T) {
	api := newTestAPIWithUtilsData(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/utils/resources/accounts", nil)

	api.UtilResourceAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UtilResourceAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("UtilResourceAccounts() invalid JSON: %v", err)
	}

	// Verify S3 buckets are in the mapping (global namespace resources)
	if accounts["arn:aws:s3:::my-test-bucket"] != "111111111111" {
		t.Errorf("UtilResourceAccounts() missing or incorrect S3 bucket, got %q", accounts["arn:aws:s3:::my-test-bucket"])
	}
	if accounts["arn:aws:s3:::another-bucket"] != "222222222222" {
		t.Errorf("UtilResourceAccounts() missing or incorrect S3 bucket, got %q", accounts["arn:aws:s3:::another-bucket"])
	}

	// Verify EC2 instance is NOT in the mapping (account is in ARN)
	if _, ok := accounts["arn:aws:ec2:us-east-1:111111111111:instance/i-1234567890abcdef0"]; ok {
		t.Error("UtilResourceAccounts() should not include EC2 instance (account is in ARN)")
	}
}

func TestAPI_UtilResourceAccounts_Empty(t *testing.T) {
	api := newTestAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/utils/resources/accounts", nil)

	api.UtilResourceAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UtilResourceAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("UtilResourceAccounts() invalid JSON: %v", err)
	}

	if len(accounts) != 0 {
		t.Errorf("UtilResourceAccounts() expected empty map, got %d entries", len(accounts))
	}
}

func TestAPI_UtilResourceAccounts_NoAccountId(t *testing.T) {
	simulator, err := sim.NewSimulator()
	if err != nil {
		t.Fatalf("Failed to create simulator: %v", err)
	}

	// Add S3 bucket without account ID - should not appear in mapping
	simulator.Universe.PutResource(entities.Resource{
		Arn:       "arn:aws:s3:::bucket-no-account",
		Type:      "AWS::S3::Bucket",
		AccountId: "", // No account ID
	})

	api := &API{Simulator: simulator}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/utils/resources/accounts", nil)

	api.UtilResourceAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UtilResourceAccounts() status = %d, want %d", w.Code, http.StatusOK)
	}

	var accounts map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &accounts); err != nil {
		t.Fatalf("UtilResourceAccounts() invalid JSON: %v", err)
	}

	// Bucket without account ID should not be in mapping
	if _, ok := accounts["arn:aws:s3:::bucket-no-account"]; ok {
		t.Error("UtilResourceAccounts() should not include bucket without account ID")
	}
}

func TestIsGlobalNamespaceType(t *testing.T) {
	tests := []struct {
		resourceType string
		want         bool
	}{
		{"AWS::S3::Bucket", true},
		{"aws::s3::bucket", true}, // case insensitive
		{"AWS::EC2::Instance", false},
		{"AWS::IAM::Role", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			if got := isGlobalNamespaceType(tt.resourceType); got != tt.want {
				t.Errorf("isGlobalNamespaceType(%q) = %v, want %v", tt.resourceType, got, tt.want)
			}
		})
	}
}

func TestAPI_UtilResourcelessActions(t *testing.T) {
	api := newTestAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/utils/actions/resourceless", nil)

	api.UtilResourcelessActions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UtilResourcelessActions() status = %d, want %d", w.Code, http.StatusOK)
	}

	var actions []string
	if err := json.Unmarshal(w.Body.Bytes(), &actions); err != nil {
		t.Fatalf("UtilResourcelessActions() invalid JSON: %v", err)
	}

	// Should return a non-empty list of resourceless actions
	if len(actions) == 0 {
		t.Error("UtilResourcelessActions() returned empty list, expected resourceless actions")
	}

	// Verify some known resourceless actions are present
	knownResourceless := map[string]bool{
		"s3:ListAllMyBuckets": false, // Note: s3:ListBuckets doesn't exist in SAR
		"ec2:DescribeRegions": false,
		"iam:ListUsers":       false,
	}

	for _, action := range actions {
		if _, ok := knownResourceless[action]; ok {
			knownResourceless[action] = true
		}
	}

	for action, found := range knownResourceless {
		if !found {
			t.Errorf("UtilResourcelessActions() missing expected action: %s", action)
		}
	}

	// Verify the list is sorted
	for i := 1; i < len(actions); i++ {
		if actions[i] < actions[i-1] {
			t.Errorf("UtilResourcelessActions() list is not sorted: %s comes after %s", actions[i], actions[i-1])
			break
		}
	}
}
