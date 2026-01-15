package entities

import (
	"reflect"
	"slices"
	"testing"
)

// -------------------------------------------------------------------------------------------------
// Universe
// -------------------------------------------------------------------------------------------------

func TestUniverse(t *testing.T) {
	uv := NewUniverse()
	if uv == nil {
		t.Fatalf("universe was unwantedly nil")
	}
}

// -------------------------------------------------------------------------------------------------
// Accounts
// -------------------------------------------------------------------------------------------------

func TestUniverse_Accounts(t *testing.T) {
	uv := NewUniverse()

	// define account
	account := Account{uv: uv, Id: "55555"}

	// check before adding account
	if uv.HasAccount(account.Id) {
		t.Fatalf("universe had account unwantedly")
	}
	if _, ok := uv.Account(account.Id); ok {
		t.Fatalf("universe found account unwantedly")
	}

	// add account
	uv.PutAccount(account)

	// check presence
	if !uv.HasAccount(account.Id) {
		t.Fatalf("universe missing account: %s", account.Id)
	}
	a, ok := uv.Account(account.Id)
	if !ok {
		t.Fatalf("universe missing account: %s", account.Id)
	}

	// check value
	if a.Id != account.Id {
		t.Fatalf("wanted account ID %s but saw %s", account.Id, a.Id)
	}

	// check collection
	collection := slices.Collect(uv.Accounts())
	if len(collection) != 1 ||
		!reflect.DeepEqual(*collection[0], account) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			account,
			collection)
	}

	// remove account
	uv.RemoveAccount(account.Id)

	// check absence
	if uv.HasAccount(account.Id) {
		t.Fatalf("universe had account unwantedly (after removal)")
	}
	if _, ok := uv.Account(account.Id); ok {
		t.Fatalf("universe found account unwantedly (after removal)")
	}
}

// -------------------------------------------------------------------------------------------------
// Groups
// -------------------------------------------------------------------------------------------------

func TestUniverse_Groups(t *testing.T) {
	uv := NewUniverse()

	// define group
	group := Group{uv: uv, Arn: "arn:aws:iam::55555:group/group-1"}

	// check before adding group
	if uv.HasGroup(group.Arn) {
		t.Fatalf("universe had group unwantedly")
	}
	if _, ok := uv.Group(group.Arn); ok {
		t.Fatalf("universe found group unwantedly")
	}

	// add group
	uv.PutGroup(group)

	// check presence
	if !uv.HasGroup(group.Arn) {
		t.Fatalf("universe missing group: %s", group.Arn)
	}
	a, ok := uv.Group(group.Arn)
	if !ok {
		t.Fatalf("universe missing group: %s", group.Arn)
	}

	// check value
	if a.Arn != group.Arn {
		t.Fatalf("wanted group ID %s but saw %s", group.Arn, a.Arn)
	}

	// check collection
	collection := slices.Collect(uv.Groups())
	if len(collection) != 1 ||
		!reflect.DeepEqual(*collection[0], group) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			group,
			collection)
	}

	// remove group
	uv.RemoveGroup(group.Arn)

	// check absence
	if uv.HasGroup(group.Arn) {
		t.Fatalf("universe had group unwantedly (after removal)")
	}
	if _, ok := uv.Group(group.Arn); ok {
		t.Fatalf("universe found group unwantedly (after removal)")
	}
}

// -------------------------------------------------------------------------------------------------
// Policies
// -------------------------------------------------------------------------------------------------

func TestUniverse_Policies(t *testing.T) {
	uv := NewUniverse()

	// define policy
	policy := ManagedPolicy{Arn: "arn:aws:iam::55555:policy/policy-1"}

	// check before adding policy
	if uv.HasPolicy(policy.Arn) {
		t.Fatalf("universe had policy unwantedly")
	}
	if _, ok := uv.Policy(policy.Arn); ok {
		t.Fatalf("universe found policy unwantedly")
	}

	// add policy
	uv.PutPolicy(policy)

	// check presence
	if !uv.HasPolicy(policy.Arn) {
		t.Fatalf("universe missing policy: %s", policy.Arn)
	}
	a, ok := uv.Policy(policy.Arn)
	if !ok {
		t.Fatalf("universe missing policy: %s", policy.Arn)
	}

	// check value
	if a.Arn != policy.Arn {
		t.Fatalf("wanted policy ID %s but saw %s", policy.Arn, a.Arn)
	}

	// check collection
	collection := slices.Collect(uv.Policies())
	if len(collection) != 1 ||
		!reflect.DeepEqual(*collection[0], policy) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			policy,
			collection)
	}

	// remove policy
	uv.RemovePolicy(policy.Arn)

	// check absence
	if uv.HasPolicy(policy.Arn) {
		t.Fatalf("universe had policy unwantedly (after removal)")
	}
	if _, ok := uv.Policy(policy.Arn); ok {
		t.Fatalf("universe found policy unwantedly (after removal)")
	}
}

// -------------------------------------------------------------------------------------------------
// Principals
// -------------------------------------------------------------------------------------------------

func TestUniverse_Principals(t *testing.T) {
	uv := NewUniverse()

	// define principal
	principal := Principal{uv: uv, Arn: "arn:aws:iam::55555:user/user-1"}

	// check before adding principal
	if uv.HasPrincipal(principal.Arn) {
		t.Fatalf("universe had principal unwantedly")
	}
	if _, ok := uv.Principal(principal.Arn); ok {
		t.Fatalf("universe found principal unwantedly")
	}

	// add principal
	uv.PutPrincipal(principal)

	// check presence
	if !uv.HasPrincipal(principal.Arn) {
		t.Fatalf("universe missing principal: %s", principal.Arn)
	}
	a, ok := uv.Principal(principal.Arn)
	if !ok {
		t.Fatalf("universe missing principal: %s", principal.Arn)
	}

	// check value
	if a.Arn != principal.Arn {
		t.Fatalf("wanted principal ID %s but saw %s", principal.Arn, a.Arn)
	}

	// check collection
	collection := slices.Collect(uv.Principals())
	if len(collection) != 1 ||
		!reflect.DeepEqual(*collection[0], principal) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			principal,
			collection)
	}

	// remove principal
	uv.RemovePrincipal(principal.Arn)

	// check absence
	if uv.HasPrincipal(principal.Arn) {
		t.Fatalf("universe had principal unwantedly (after removal)")
	}
	if _, ok := uv.Principal(principal.Arn); ok {
		t.Fatalf("universe found principal unwantedly (after removal)")
	}
}

// -------------------------------------------------------------------------------------------------
// Resources
// -------------------------------------------------------------------------------------------------

func TestUniverse_Resources(t *testing.T) {
	uv := NewUniverse()

	// define resource
	resource := Resource{uv: uv, Arn: "arn:aws:iam::55555:user/user-1"}

	// check before adding resource
	if uv.HasResource(resource.Arn) {
		t.Fatalf("universe had resource unwantedly")
	}
	if _, ok := uv.Resource(resource.Arn); ok {
		t.Fatalf("universe found resource unwantedly")
	}

	// add resource
	uv.PutResource(resource)

	// check presence
	if !uv.HasResource(resource.Arn) {
		t.Fatalf("universe missing resource: %s", resource.Arn)
	}
	a, ok := uv.Resource(resource.Arn)
	if !ok {
		t.Fatalf("universe missing resource: %s", resource.Arn)
	}

	// check value
	if a.Arn != resource.Arn {
		t.Fatalf("wanted resource ID %s but saw %s", resource.Arn, a.Arn)
	}

	// check collection
	collection := slices.Collect(uv.Resources())
	if len(collection) != 1 ||
		!reflect.DeepEqual(*collection[0], resource) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			resource,
			collection)
	}

	// remove resource
	uv.RemoveResource(resource.Arn)

	// check absence
	if uv.HasResource(resource.Arn) {
		t.Fatalf("universe had resource unwantedly (after removal)")
	}
	if _, ok := uv.Resource(resource.Arn); ok {
		t.Fatalf("universe found resource unwantedly (after removal)")
	}
}

func TestUniverse_Subresources(t *testing.T) {
	uv := NewUniverse()

	// define resource
	resource := Resource{uv: uv, Type: "AWS::S3::Bucket", Arn: "arn:aws:s3:::mybucket"}
	subresourceArn := resource.Arn + "/object.txt"

	// add resource
	uv.PutResource(resource)

	// check subresource
	a, ok := uv.Resource(subresourceArn)
	if !ok {
		t.Fatalf("universe missing subresource: %s", subresourceArn)
	}

	expected := &Resource{
		uv:   resource.uv,
		Type: "AWS::S3::Bucket::Object",
		Arn:  "arn:aws:s3:::mybucket/object.txt",
	}

	// check result
	if !reflect.DeepEqual(a, expected) {
		t.Fatalf("wanted subresource to be %#v but got %#v", expected, a)
	}
}

// -------------------------------------------------------------------------------------------------
// Merge
// -------------------------------------------------------------------------------------------------

func TestUniverse_Merge(t *testing.T) {
	uv1 := NewUniverse()
	uv1.PutAccount(Account{Id: "111111111111"})
	uv1.PutGroup(Group{Arn: "arn:aws:iam::111111111111:group/group1"})
	uv1.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::111111111111:policy/pol1"})
	uv1.PutPrincipal(Principal{Arn: "arn:aws:iam::111111111111:role/role1"})
	uv1.PutResource(Resource{Arn: "arn:aws:s3:::bucket1"})

	uv2 := NewUniverse()
	uv2.PutAccount(Account{Id: "222222222222"})
	uv2.PutGroup(Group{Arn: "arn:aws:iam::222222222222:group/group2"})
	uv2.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::222222222222:policy/pol2"})
	uv2.PutPrincipal(Principal{Arn: "arn:aws:iam::222222222222:role/role2"})
	uv2.PutResource(Resource{Arn: "arn:aws:s3:::bucket2"})

	uv1.Merge(uv2)

	// Verify merged counts
	if uv1.NumAccounts() != 2 {
		t.Fatalf("expected 2 accounts, got %d", uv1.NumAccounts())
	}
	if uv1.NumGroups() != 2 {
		t.Fatalf("expected 2 groups, got %d", uv1.NumGroups())
	}
	if uv1.NumPolicies() != 2 {
		t.Fatalf("expected 2 policies, got %d", uv1.NumPolicies())
	}
	if uv1.NumPrincipals() != 2 {
		t.Fatalf("expected 2 principals, got %d", uv1.NumPrincipals())
	}
	if uv1.NumResources() != 2 {
		t.Fatalf("expected 2 resources, got %d", uv1.NumResources())
	}

	// Verify all entities are accessible
	if !uv1.HasAccount("111111111111") || !uv1.HasAccount("222222222222") {
		t.Fatal("missing accounts after merge")
	}
}

// -------------------------------------------------------------------------------------------------
// Overlay
// -------------------------------------------------------------------------------------------------

func TestUniverse_Overlay(t *testing.T) {
	base := NewUniverse()
	overlay := NewUniverse()

	// Test with overlay
	result := base.Overlay(overlay)
	if len(result) != 2 {
		t.Fatalf("expected 2 universes, got %d", len(result))
	}
	if result[0] != overlay {
		t.Fatal("overlay should be first (higher priority)")
	}
	if result[1] != base {
		t.Fatal("base should be second")
	}

	// Test with nil overlay
	result = base.Overlay(nil)
	if len(result) != 1 {
		t.Fatalf("expected 1 universe, got %d", len(result))
	}
	if result[0] != base {
		t.Fatal("base should be the only element")
	}
}

// -------------------------------------------------------------------------------------------------
// Size
// -------------------------------------------------------------------------------------------------

func TestUniverse_Size(t *testing.T) {
	uv := NewUniverse()

	if uv.Size() != 0 {
		t.Fatalf("expected size 0, got %d", uv.Size())
	}

	uv.PutAccount(Account{Id: "123456789012"})
	uv.PutGroup(Group{Arn: "arn:aws:iam::123456789012:group/g"})
	uv.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::123456789012:policy/p"})
	uv.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/r"})
	uv.PutResource(Resource{Arn: "arn:aws:s3:::bucket"})

	if uv.Size() != 5 {
		t.Fatalf("expected size 5, got %d", uv.Size())
	}
}

// -------------------------------------------------------------------------------------------------
// Arns methods
// -------------------------------------------------------------------------------------------------

func TestUniverse_GroupArns(t *testing.T) {
	uv := NewUniverse()
	uv.PutGroup(Group{Arn: "arn:aws:iam::123456789012:group/g1"})
	uv.PutGroup(Group{Arn: "arn:aws:iam::123456789012:group/g2"})

	arns := uv.GroupArns()
	if len(arns) != 2 {
		t.Fatalf("expected 2 arns, got %d", len(arns))
	}
}

func TestUniverse_PolicyArns(t *testing.T) {
	uv := NewUniverse()
	uv.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::123456789012:policy/p1"})
	uv.PutPolicy(ManagedPolicy{Arn: "arn:aws:iam::123456789012:policy/p2"})

	arns := uv.PolicyArns()
	if len(arns) != 2 {
		t.Fatalf("expected 2 arns, got %d", len(arns))
	}
}

func TestUniverse_PrincipalArns(t *testing.T) {
	uv := NewUniverse()
	uv.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/r1"})
	uv.PutPrincipal(Principal{Arn: "arn:aws:iam::123456789012:role/r2"})

	arns := uv.PrincipalArns()
	if len(arns) != 2 {
		t.Fatalf("expected 2 arns, got %d", len(arns))
	}
}

func TestUniverse_ResourceArns(t *testing.T) {
	uv := NewUniverse()
	uv.PutResource(Resource{Arn: "arn:aws:s3:::bucket1"})
	uv.PutResource(Resource{Arn: "arn:aws:s3:::bucket2"})

	arns := uv.ResourceArns()
	if len(arns) != 2 {
		t.Fatalf("expected 2 arns, got %d", len(arns))
	}
}

// -------------------------------------------------------------------------------------------------
// LoadBasePolicies
// -------------------------------------------------------------------------------------------------

func TestUniverse_LoadBasePolicies(t *testing.T) {
	uv := NewUniverse()

	// Initially no policies
	if uv.NumPolicies() != 0 {
		t.Fatalf("expected 0 policies initially, got %d", uv.NumPolicies())
	}

	// Load base policies
	uv.LoadBasePolicies()

	// Should have loaded AWS managed policies
	if uv.NumPolicies() == 0 {
		t.Fatal("expected policies to be loaded")
	}

	initialCount := uv.NumPolicies()

	// Call again should be idempotent
	uv.LoadBasePolicies()

	if uv.NumPolicies() != initialCount {
		t.Fatal("LoadBasePolicies should be idempotent")
	}
}
