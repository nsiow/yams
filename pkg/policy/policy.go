package policy

import (
	"encoding/json"
	"fmt"
)

// Policy represents the grammar and structure of an AWS IAM Policy
type Policy struct {
	// Context   string `json:"-"` // TODO(nsiow) implement context from Policy -> Statement for tracking
	Version   string
	Id        string
	Statement StatementBlock
}

// StatementBlock represents one or more statements, provided in array or map form
type StatementBlock []Statement

// UnmarshalJSON instructs how to create StatementBlock fields from raw bytes
func (s *StatementBlock) UnmarshalJSON(data []byte) error {
	// Handle empty/too-small string
	if len(data) < 2 {
		return fmt.Errorf("invalid statement block: %s", data)
	}

	// Check for null case
	if len(data) == 4 && string(data) == "null" {
		*s = []Statement{}
		return nil
	}

	// Handle single statement
	if data[0] == '{' && data[len(data)-1] == '}' {
		stmt := Statement{}
		err := json.Unmarshal(data, &stmt)
		if err != nil {
			return fmt.Errorf("error in single-statement clause of Statement:\ndata=%s\nerror=%v", data, err)
		}

		*s = []Statement{stmt}
		return nil
	}

	// Handle list of statements
	if data[0] == '[' && data[len(data)-1] == ']' {
		var list []Statement
		err := json.Unmarshal(data, &list)
		if err != nil {
			return fmt.Errorf("error in multi-statement clause of Statement:\ndata=%s\nerror=%v", data, err)
		}
		*s = list
		return nil
	}

	return fmt.Errorf("not sure how to handle statement block: %s", data)
}

// Statement represents the grammar and structure of an AWS IAM Statement
type Statement struct {
	Sid          string
	Effect       string
	Principal    Principal      `json:",omitempty"`
	NotPrincipal Principal      `json:",omitempty"`
	Action       Action         `json:",omitempty"`
	NotAction    Action         `json:",omitempty"`
	Resource     Resource       `json:",omitempty"`
	NotResource  Resource       `json:",omitempty"`
	Condition    ConditionBlock `json:",omitempty"`
}

// Validate determines whether or not the Statement is valid; returning an error otherwise
//
// Validity here is strictly in terms of the IAM grammar, and makes no guarantees around policy values
func (s *Statement) Validate() error {
	if !s.Principal.Empty() && !s.NotPrincipal.Empty() {
		return fmt.Errorf("must supply exactly zero or one of (Principal | NotPrincipal)")
	}
	if s.Action.Empty() == s.NotAction.Empty() {
		return fmt.Errorf("must supply exactly one of (Action | NotAction)")
	}
	if s.Resource.Empty() == s.NotResource.Empty() {
		return fmt.Errorf("must supply exactly one of (Resource | NotResource)")
	}

	return nil
}

// Principal represents a set of Principals, provided in string or map form
type Principal PrincipalMap

// Empty determines whether or not the specified Principal field is empty
func (p *Principal) Empty() bool {
	return p.AWS.Empty() &&
		p.Service.Empty() &&
		p.Federated.Empty() &&
		p.CanonicalUser.Empty()
}

// PrincipalMap represents the grammar and structure of an AWS IAM Principal represented in map form
type PrincipalMap struct {
	AWS           Value `json:",omitempty"`
	Federated     Value `json:",omitempty"`
	Service       Value `json:",omitempty"`
	CanonicalUser Value `json:",omitempty"`
}

// UnmarshalJSON instructs how to create Principal fields from raw bytes
func (p *Principal) UnmarshalJSON(data []byte) error {
	// Handle string case; only valid in this 3-byte sequence
	if len(data) == 3 && string(data) == `"*"` {
		p.AWS = []string{"*"}
		p.Federated = []string{"*"}
		p.Service = []string{"*"}
		p.CanonicalUser = []string{"*"}
		return nil
	}

	var principal PrincipalMap
	err := json.Unmarshal(data, &principal)
	if err != nil {
		return fmt.Errorf("unable to parse:\nprincipal = %s\nerror = %v", data, err)
	}

	p.AWS = principal.AWS
	p.Federated = principal.Federated
	p.Service = principal.Service
	p.CanonicalUser = principal.CanonicalUser
	return nil
}

// Action represents the grammar and structure of an AWS IAM Action
type Action = Value

// Resource represents the grammar and structure of an AWS IAM Resource
type Resource = Value

// ConditionBlock represents the grammar and structure of an AWS IAM Condition block
type ConditionBlock = map[ConditionOperation]Condition

// ConditionOperation represents the operation portion of a condition block
type ConditionOperation = string

// Condition represents the grammar and structure of an AWS IAM Condition
type Condition = map[ConditionKey]ConditionValue

// ConditionKey represents the key portion of a condition
type ConditionKey = string

// ConditionValue represents the value portion of a condition
type ConditionValue any
