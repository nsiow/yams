package entities

import (
	"testing"
	"time"
)

func TestNewOverlay(t *testing.T) {
	o := NewOverlay("test-overlay")

	if o.Name != "test-overlay" {
		t.Errorf("expected name 'test-overlay', got %q", o.Name)
	}

	// ID should be sha1 hash of name
	expectedID := GenerateOverlayID("test-overlay")
	if o.ID != expectedID {
		t.Errorf("expected ID %q, got %q", expectedID, o.ID)
	}

	// CreatedAt should be recent
	if time.Since(o.CreatedAt) > time.Second {
		t.Error("CreatedAt should be recent")
	}

	// Universe should be initialized
	if o.Universe == nil {
		t.Error("Universe should not be nil")
	}
}

func TestNewOverlayWithTime(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	o := NewOverlayWithTime("test-overlay", createdAt)

	if o.Name != "test-overlay" {
		t.Errorf("expected name 'test-overlay', got %q", o.Name)
	}

	if !o.CreatedAt.Equal(createdAt) {
		t.Errorf("expected CreatedAt %v, got %v", createdAt, o.CreatedAt)
	}
}

func TestGenerateOverlayID(t *testing.T) {
	// Same name should produce same ID
	id1 := GenerateOverlayID("test-overlay")
	id2 := GenerateOverlayID("test-overlay")
	if id1 != id2 {
		t.Error("same name should produce same ID")
	}

	// Different names should produce different IDs
	id3 := GenerateOverlayID("other-overlay")
	if id1 == id3 {
		t.Error("different names should produce different IDs")
	}

	// ID should be 40 characters (sha1 hex)
	if len(id1) != 40 {
		t.Errorf("expected ID length 40, got %d", len(id1))
	}
}

func TestOverlay_Size(t *testing.T) {
	o := NewOverlay("test")

	if o.Size() != 0 {
		t.Errorf("expected size 0, got %d", o.Size())
	}

	o.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	o.Universe.PutResource(Resource{Arn: "arn:aws:s3:::test-bucket"})

	if o.Size() != 2 {
		t.Errorf("expected size 2, got %d", o.Size())
	}
}

func TestOverlay_IsEmpty(t *testing.T) {
	o := NewOverlay("test")

	if !o.IsEmpty() {
		t.Error("new overlay should be empty")
	}

	o.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/test"})

	if o.IsEmpty() {
		t.Error("overlay with entities should not be empty")
	}
}

func TestOverlay_NumMethods(t *testing.T) {
	o := NewOverlay("test")

	// All counts should be zero initially
	if o.NumPrincipals() != 0 {
		t.Error("expected 0 principals")
	}
	if o.NumResources() != 0 {
		t.Error("expected 0 resources")
	}
	if o.NumPolicies() != 0 {
		t.Error("expected 0 policies")
	}
	if o.NumAccounts() != 0 {
		t.Error("expected 0 accounts")
	}
	if o.NumGroups() != 0 {
		t.Error("expected 0 groups")
	}

	// Add one of each
	o.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	o.Universe.PutResource(Resource{Arn: "arn:aws:s3:::test-bucket"})
	o.Universe.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::123456789012:policy/test"})
	o.Universe.PutAccount(Account{Id: "123456789012"})
	o.Universe.PutGroup(Group{Arn: "arn:aws:iam::123456789012:group/test"})

	if o.NumPrincipals() != 1 {
		t.Error("expected 1 principal")
	}
	if o.NumResources() != 1 {
		t.Error("expected 1 resource")
	}
	if o.NumPolicies() != 1 {
		t.Error("expected 1 policy")
	}
	if o.NumAccounts() != 1 {
		t.Error("expected 1 account")
	}
	if o.NumGroups() != 1 {
		t.Error("expected 1 group")
	}
}

func TestOverlay_NilUniverse(t *testing.T) {
	// Test methods handle nil Universe gracefully
	o := &Overlay{Name: "test"}

	if o.Size() != 0 {
		t.Error("Size should return 0 for nil universe")
	}
	if !o.IsEmpty() {
		t.Error("IsEmpty should return true for nil universe")
	}
	if o.NumPrincipals() != 0 {
		t.Error("NumPrincipals should return 0 for nil universe")
	}
	if o.NumResources() != 0 {
		t.Error("NumResources should return 0 for nil universe")
	}
	if o.NumPolicies() != 0 {
		t.Error("NumPolicies should return 0 for nil universe")
	}
	if o.NumAccounts() != 0 {
		t.Error("NumAccounts should return 0 for nil universe")
	}
	if o.NumGroups() != 0 {
		t.Error("NumGroups should return 0 for nil universe")
	}
}

func TestOverlay_ToDataAndFromData(t *testing.T) {
	o := NewOverlay("test-overlay")

	// Add entities
	o.Universe.PutAccount(Account{Id: "123456789012", Name: "Test Account"})
	o.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	o.Universe.PutResource(Resource{Arn: "arn:aws:s3:::test-bucket"})
	o.Universe.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::123456789012:policy/test"})
	o.Universe.PutGroup(Group{Arn: "arn:aws:iam::123456789012:group/test"})

	// Convert to data
	data := o.ToData()

	if data.Name != "test-overlay" {
		t.Errorf("expected name 'test-overlay', got %q", data.Name)
	}
	if data.ID != o.ID {
		t.Errorf("expected ID %q, got %q", o.ID, data.ID)
	}
	if len(data.Accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(data.Accounts))
	}
	if len(data.Principals) != 1 {
		t.Errorf("expected 1 principal, got %d", len(data.Principals))
	}
	if len(data.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(data.Resources))
	}
	if len(data.Policies) != 1 {
		t.Errorf("expected 1 policy, got %d", len(data.Policies))
	}
	if len(data.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(data.Groups))
	}

	// Convert back from data
	restored := FromData(data)

	if restored.Name != o.Name {
		t.Errorf("expected name %q, got %q", o.Name, restored.Name)
	}
	if restored.ID != o.ID {
		t.Errorf("expected ID %q, got %q", o.ID, restored.ID)
	}
	if restored.NumAccounts() != 1 {
		t.Errorf("expected 1 account, got %d", restored.NumAccounts())
	}
	if restored.NumPrincipals() != 1 {
		t.Errorf("expected 1 principal, got %d", restored.NumPrincipals())
	}
	if restored.NumResources() != 1 {
		t.Errorf("expected 1 resource, got %d", restored.NumResources())
	}
	if restored.NumPolicies() != 1 {
		t.Errorf("expected 1 policy, got %d", restored.NumPolicies())
	}
	if restored.NumGroups() != 1 {
		t.Errorf("expected 1 group, got %d", restored.NumGroups())
	}
}

func TestOverlay_Summary(t *testing.T) {
	o := NewOverlay("test-overlay")
	o.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	o.Universe.PutResource(Resource{Arn: "arn:aws:s3:::test-bucket"})

	summary := o.Summary()

	if summary.Name != "test-overlay" {
		t.Errorf("expected name 'test-overlay', got %q", summary.Name)
	}
	if summary.NumPrincipals != 1 {
		t.Errorf("expected 1 principal, got %d", summary.NumPrincipals)
	}
	if summary.NumResources != 1 {
		t.Errorf("expected 1 resource, got %d", summary.NumResources)
	}
}

func TestOverlay_Clone(t *testing.T) {
	o := NewOverlay("original")
	o.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	o.Universe.PutResource(Resource{Arn: "arn:aws:s3:::test-bucket"})

	clone := o.Clone("cloned")

	// Verify new name and ID
	if clone.Name != "cloned" {
		t.Errorf("expected name 'cloned', got %q", clone.Name)
	}
	if clone.ID == o.ID {
		t.Error("clone should have different ID")
	}

	// Verify entities were copied
	if clone.NumPrincipals() != 1 {
		t.Errorf("expected 1 principal, got %d", clone.NumPrincipals())
	}
	if clone.NumResources() != 1 {
		t.Errorf("expected 1 resource, got %d", clone.NumResources())
	}

	// Verify independence (modifying clone doesn't affect original)
	clone.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/another"})
	if o.NumPrincipals() != 1 {
		t.Error("modifying clone affected original")
	}
}

func TestOverlay_ArnMethods(t *testing.T) {
	o := NewOverlay("test")
	o.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/b"})
	o.Universe.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/a"})
	o.Universe.PutResource(Resource{Arn: "arn:aws:s3:::bucket-b"})
	o.Universe.PutResource(Resource{Arn: "arn:aws:s3:::bucket-a"})
	o.Universe.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::123456789012:policy/b"})
	o.Universe.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::123456789012:policy/a"})
	o.Universe.PutGroup(Group{Arn: "arn:aws:iam::123456789012:group/b"})
	o.Universe.PutGroup(Group{Arn: "arn:aws:iam::123456789012:group/a"})
	o.Universe.PutAccount(Account{Id: "222222222222"})
	o.Universe.PutAccount(Account{Id: "111111111111"})

	// Verify all are sorted
	principals := o.PrincipalArns()
	if len(principals) != 2 || principals[0] != "arn:aws:iam::123456789012:role/a" {
		t.Errorf("PrincipalArns not sorted correctly: %v", principals)
	}

	resources := o.ResourceArns()
	if len(resources) != 2 || resources[0] != "arn:aws:s3:::bucket-a" {
		t.Errorf("ResourceArns not sorted correctly: %v", resources)
	}

	policies := o.PolicyArns()
	if len(policies) != 2 || policies[0] != "arn:aws:iam::123456789012:policy/a" {
		t.Errorf("PolicyArns not sorted correctly: %v", policies)
	}

	groups := o.GroupArns()
	if len(groups) != 2 || groups[0] != "arn:aws:iam::123456789012:group/a" {
		t.Errorf("GroupArns not sorted correctly: %v", groups)
	}

	accounts := o.AccountIds()
	if len(accounts) != 2 || accounts[0] != "111111111111" {
		t.Errorf("AccountIds not sorted correctly: %v", accounts)
	}
}

func TestOverlay_ArnMethods_NilUniverse(t *testing.T) {
	o := &Overlay{Name: "test"}

	if o.PrincipalArns() != nil {
		t.Error("PrincipalArns should return nil for nil universe")
	}
	if o.ResourceArns() != nil {
		t.Error("ResourceArns should return nil for nil universe")
	}
	if o.PolicyArns() != nil {
		t.Error("PolicyArns should return nil for nil universe")
	}
	if o.GroupArns() != nil {
		t.Error("GroupArns should return nil for nil universe")
	}
	if o.AccountIds() != nil {
		t.Error("AccountIds should return nil for nil universe")
	}
}
