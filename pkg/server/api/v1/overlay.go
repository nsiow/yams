package v1

import "github.com/nsiow/yams/pkg/entities"

type Overlay struct {
	Accounts   []entities.Account       `json:"accounts"`
	Groups     []entities.Group         `json:"groups"`
	Policies   []entities.ManagedPolicy `json:"policies"`
	Principals []entities.Principal     `json:"principals"`
	Resources  []entities.Resource      `json:"resources"`
}

func (s *Overlay) Universe() *entities.Universe {
	return entities.NewBuilder().
		WithAccounts(s.Accounts...).
		WithGroups(s.Groups...).
		WithPolicies(s.Policies...).
		WithPrincipals(s.Principals...).
		WithResources(s.Resources...).
		Build()
}
