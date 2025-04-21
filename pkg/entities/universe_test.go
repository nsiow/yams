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
	if !reflect.DeepEqual(slices.Collect(uv.Accounts()), []Account{account}) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			[]Account{account},
			slices.Collect(uv.Accounts()))
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
	if !reflect.DeepEqual(slices.Collect(uv.Groups()), []Group{group}) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			[]Group{group},
			slices.Collect(uv.Groups()))
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
	if !reflect.DeepEqual(slices.Collect(uv.Policies()), []ManagedPolicy{policy}) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			[]ManagedPolicy{policy},
			slices.Collect(uv.Policies()))
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
	if !reflect.DeepEqual(slices.Collect(uv.Principals()), []Principal{principal}) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			[]Principal{principal},
			slices.Collect(uv.Principals()))
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
	if !reflect.DeepEqual(slices.Collect(uv.Resources()), []Resource{resource}) {
		t.Fatalf("wanted collection to be %+v but got %+v",
			[]Resource{resource},
			slices.Collect(uv.Resources()))
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
