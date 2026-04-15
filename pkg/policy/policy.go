package policy

import (
	"fmt"
)

// Policy represents the grammar and structure of an AWS IAM Policy
type Policy struct {
	Name      string `json:"_Name,omitempty"` // Inline policy name (custom field, not part of AWS policy)
	Version   string
	Id        string `json:",omitempty"`
	Statement StatementBlock
}

// Empty returns true if the policy is empty of any effective statements
func (p *Policy) Empty() bool {
	return len(p.Statement) == 0
}

// Validate determines whether or not the Policy is valid; returning an error otherwise
func (p *Policy) Validate() error {
	for _, stmt := range p.Statement {
		if err := stmt.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// StatementBlock represents one or more statements, provided in array or map form
type StatementBlock []Statement

// Statement represents the grammar and structure of an AWS IAM Statement
type Statement struct {
	Sid          string `json:",omitempty"`
	Effect       Effect
	Principal    Principal      `json:",omitzero"`
	NotPrincipal Principal      `json:",omitzero"`
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
	if (s.Principal.All || !s.Principal.Empty()) && (s.NotPrincipal.All || !s.NotPrincipal.Empty()) {
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

// Effect corresponds to the Allow/Deny directive of the Policy
type Effect string

// EFFECT_Allow corresponds to Effect=Allow in an IAM policy
const EFFECT_ALLOW = "Allow"

// EFFECT_DENY corresponds to Effect=Deny in an IAM policy
const EFFECT_DENY = "Deny"

// Principal represents the grammar and structure of an AWS IAM Principal represented in map form
type Principal struct {
	All           bool  `json:"-"` // case for Principal=*
	AWS           Value `json:",omitempty"`
	Federated     Value `json:",omitempty"`
	Service       Value `json:",omitempty"`
	CanonicalUser Value `json:",omitempty"`
}

// Empty determines whether or not the specified Principal field is empty
func (p *Principal) Empty() bool {
	return !p.All &&
		p.AWS.Empty() &&
		p.Service.Empty() &&
		p.Federated.Empty() &&
		p.CanonicalUser.Empty()
}

// Action represents the grammar and structure of an AWS IAM Action
type Action = Value

// Resource represents the grammar and structure of an AWS IAM Resource
type Resource = Value

// ConditionBlock represents the grammar and structure of an AWS IAM Condition block
type ConditionBlock = map[ConditionOperator]ConditionValues

// ConditionOperator represents the operation portion of a condition block
type ConditionOperator = string

// ConditionValues represents the key/value portion of a condition block
type ConditionValues = map[string]Value
