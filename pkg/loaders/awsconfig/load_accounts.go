package awsconfig

import (
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// loadAccounts takes a list of AWS Config items and extracts Accounts
func loadAccounts(items []ConfigItem) (*AccountMap, error) {
	a := NewAccountMap()

	// Iterate through our AWS Config items, we only look at Yams::Organization::Account
	for _, i := range items {
		if i.Type != CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT {
			continue
		}

		account, err := loadAccount(i)
		if err != nil {
			return nil, fmt.Errorf("error loading account '%s': %w", i.AccountId, err)
		}
		a.Add(i.AccountId, account)
	}

	return a, nil
}

// loadAccount takes a single AWS Config item and returns a parsed account object
func loadAccount(i ConfigItem) (entities.Account, error) {
	af := accountFragment{}
	err := json.Unmarshal(i.Configuration, &af)
	if err != nil {
		return entities.Account{},
			fmt.Errorf("unable to parse account fragment (%+v): %v", af, err)
	}

	return entities.Account{
		Id:       af.AccountId,
		OrgId:    af.OrgId,
		OrgPaths: af.OrgPaths,
		SCPs:     af.SCPs,
	}, nil
}

// accountFragment is a struct describing a configuration fragment for
// Yams::Organizations::Account
type accountFragment struct {
	AccountId string            `json:"accountId"`
	OrgId     string            `json:"orgId"`
	OrgPaths  []string          `json:"orgPaths"`
	SCPs      [][]policy.Policy `json:"serviceControlPolicies"`
}
