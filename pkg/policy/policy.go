package policy

import (
	"encoding/json"
	"fmt"

	ps "github.com/nsiow/yams/pkg/polystring"
)

// Policy represents the grammar and structure of an AWS IAM Policy
type Policy struct {
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
		return fmt.Errorf("invalid statement block: %s", string(data))
	}

	// Handle single statement
	if data[0] == '{' && data[len(data)-1] == '}' {
		stmt := Statement{}
		err := json.Unmarshal(data, &stmt)
		if err != nil {
			return fmt.Errorf("error in single-statement clause processing of Statement block: %v", err)
		}

		*s = []Statement{stmt}
		return nil
	}

	// Handle list of statements
	if data[0] == '[' && data[len(data)-1] == ']' {
		var list []Statement
		err := json.Unmarshal(data, &list)
		if err != nil {
			return fmt.Errorf("error in multi-statement clause processing of Statement block: %v", err)
		}
		*s = list
		return nil
	}

	return fmt.Errorf("not sure how to handle statement block: %s", string(data))
}

// Statement represents the grammar and structure of an AWS IAM Statement
type Statement struct {
	Sid          string
	Principal    Principal `json:",omitempty"`
	NotPrincipal Principal `json:",omitempty"`
	Effect       string
	Action       Action    `json:",omitempty"`
	NotAction    Action    `json:",omitempty"`
	Resource     Resource  `json:",omitempty"`
	NotResource  Resource  `json:",omitempty"`
	Condition    Condition `json:",omitempty"`
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

// Principal represents the grammar and structure of an AWS IAM Principal
type Principal struct {
	AWS           ps.PolyString `json:",omitempty"`
	Federated     ps.PolyString `json:",omitempty"`
	Service       ps.PolyString `json:",omitempty"`
	CanonicalUser ps.PolyString `json:",omitempty"`
}

// UnmarshalJSON instructs how to create Principal fields from raw bytes
func (p *Principal) UnmarshalJSON(data []byte) error {
	// Handle empty string
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	// Handle wildcard
	if len(data) == 1 && data[0] == '*' {
		p.AWS = ps.NewPolyString("*")
		p.Federated = ps.NewPolyString("*")
		p.Service = ps.NewPolyString("*")
		p.CanonicalUser = ps.NewPolyString("*")
	}

	// TODO(nsiow) better way to do this?
	// Handle normal case
	var m map[string]json.RawMessage
	if v, ok := m["AWS"]; ok {
		if err := json.Unmarshal(v, &p.AWS); err != nil {
			return fmt.Errorf("error in 'AWS' clause processing of Principal block: %v", err)
		}
	}
	if v, ok := m["Federated"]; ok {
		if err := json.Unmarshal(v, &p.Federated); err != nil {
			return fmt.Errorf("error in 'Federated' clause processing of Principal block: %v", err)
		}
	}
	if v, ok := m["Service"]; ok {
		if err := json.Unmarshal(v, &p.Service); err != nil {
			return fmt.Errorf("error in 'Service' clause processing of Principal block: %v", err)
		}
	}
	if v, ok := m["CanonicalUser"]; ok {
		if err := json.Unmarshal(v, &p.CanonicalUser); err != nil {
			return fmt.Errorf("error in 'CanonicalUser' clause processing of Principal block: %v", err)
		}
	}
	return nil
}

// Empty determines whether or not the specified Principal field is empty
func (p *Principal) Empty() bool {
	return p.AWS.Empty() &&
		p.Service.Empty() &&
		p.Federated.Empty() &&
		p.CanonicalUser.Empty()
}

// Action represents the grammar and structure of an AWS IAM Action
type Action struct {
	ps.PolyString
}

// Resource represents the grammar and structure of an AWS IAM Resource
type Resource struct {
	ps.PolyString
}

// ConditionMap represents the grammar and structure of an AWS IAM Condition map
type ConditionMap struct {
	Map map[string]Condition
}

// UnmarshalJSON instructs how to create ConditionMap fields from raw bytes
func (c *ConditionMap) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &c.Map)
	if err != nil {
		return fmt.Errorf("error in 'ConditionMap' clause processing of Condition block: %v", err)
	}
	return nil
}

// Condition represents the grammar and structure of an AWS IAM Condition
type Condition struct {
	Map map[string]ps.PolyString
}

// UnmarshalJSON instructs how to create Condition fields from raw bytes
func (c *Condition) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &c.Map)
	if err != nil {
		return fmt.Errorf("error in 'Condition' clause processing of Condition block: %v", err)
	}
	return nil
}
