package policy

import (
	"encoding/json"
	"fmt"
)

// Policy represents the grammar and structure of an AWS IAM Policy
type Policy struct {
	Version   string
	Id        string
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

// UnmarshalJSON instructs how to create Effect fields from raw bytes
func (e *Effect) UnmarshalJSON(data []byte) error {
	var effect string
	err := json.Unmarshal(data, &effect)
	if err != nil {
		return fmt.Errorf("unable to parse:\neffect = %s\nerror = %v", data, err)
	}

	switch effect {
	case EFFECT_ALLOW:
		*e = EFFECT_ALLOW
		return nil
	case EFFECT_DENY:
		*e = EFFECT_DENY
		return nil
	default:
		return fmt.Errorf("invalid value for 'Effect' field: %s", effect)
	}
}

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

// IsZero is used for marshaling to indicate when the field should be omitted
func (p *Principal) IsZero() bool {
	return p.Empty()
}

// MarshalJSON instructs how to convert Principal fields to raw bytes
func (p *Principal) MarshalJSON() ([]byte, error) {
	if p.All {
		return []byte(`"*"`), nil
	}

	type alias *Principal
	a := alias(p)
	return json.Marshal(a)
}

// UnmarshalJSON instructs how to create Principal fields from raw bytes
func (p *Principal) UnmarshalJSON(data []byte) error {
	// Handle string case; only valid in this 3-byte sequence
	if len(data) == 3 && string(data) == `"*"` {
		p.All = true
		return nil
	}

	type alias *Principal
	a := alias(p)
	err := json.Unmarshal(data, &a)
	if err != nil {
		return fmt.Errorf("unable to parse:\nprincipal = %s\nerror = %v", data, err)
	}

	p.AWS = a.AWS
	p.Federated = a.Federated
	p.Service = a.Service
	p.CanonicalUser = a.CanonicalUser
	return nil
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
