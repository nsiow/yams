package awsconfig

import (
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
)

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
	AccountId string           `json:"accountId"`
	OrgId     string           `json:"orgId"`
	OrgPaths  []string         `json:"orgPaths"`
	SCPs      [][]entities.Arn `json:"serviceControlPolicies"`
}
