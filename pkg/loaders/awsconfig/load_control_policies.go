package awsconfig

import (
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

type ControlPolicyStruct struct {
	SCPs *ControlPolicyMap
}

// loadControlPolicies takes a list of AWS Config items and extracts Service Control Policies
func loadControlPolicies(items []ConfigItem) (*ControlPolicyStruct, error) {
	cp := ControlPolicyStruct{
		SCPs: NewControlPolicyMap(),
	}

	// Iterate through our AWS Config items, we only look at AWS::Yams::Account
	for _, i := range items {
		if i.Type != CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT {
			continue
		}

		// Save the decoded policy into our map
		fragment, err := loadControlPolicy(i)
		if err != nil {
			return nil, err
		}
		cp.SCPs.Add(i.Account, fragment.ServiceControlPolicies)
	}

	return &cp, nil
}

// --------------------------------------------------------------------------------
// IAM Policies
// --------------------------------------------------------------------------------

// loadControlPolicy takes a single AWS Config item and returns a parsed control policy object
func loadControlPolicy(i ConfigItem) (*controlPolicyFragment, error) {
	// Parse policy configuration
	pf := controlPolicyFragment{}
	err := json.Unmarshal(i.Configuration, &pf)
	if err != nil {
		return nil, fmt.Errorf("unable to parse control policy fragment (%+v): %v", pf, err)
	}

	return &pf, nil
}

// controlPolicyFragment is a struct describing a configuration fragment for
// Yams::Organizations::Account
type controlPolicyFragment struct {
	ServiceControlPolicies [][]policy.Policy `json:"serviceControlPolicies"`
}
