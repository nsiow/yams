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
type StatementBlock struct {
	Values []Statement
}

// UnmarshalJSON instructs how to create StatementBlock fields from raw bytes
func (s *StatementBlock) UnmarshalJSON(data []byte) error {
	// Handle empty string
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	// Handle single statement
	if data[0] == '{' && data[len(data)-1] == '}' {
		stmt := Statement{}
		err := json.Unmarshal(data, &stmt)
		if err != nil {
			return err
		}

		s.Values = []Statement{stmt}
		return nil
	}

	// Handle list of statements
	if data[0] == '[' && data[len(data)-1] == ']' {
		return json.Unmarshal(data, &s.Values)
	}

	return fmt.Errorf("not sure how to handle statement block: %s", string(data))
}

// Statement represents the grammar and structure of an AWS IAM Statement
type Statement struct {
	Sid          string
	Principal    *Principal `json:",omitempty"`
	NotPrincipal *Principal `json:",omitempty"`
	Action       *Action    `json:",omitempty"`
	NotAction    *Action    `json:",omitempty"`
	Resource     *Resource  `json:",omitempty"`
	NotResource  *Resource  `json:",omitempty"`
	Condition    *Condition `json:",omitempty"`
}

// Validate determines whether or not the Statement is valid; returning an error otherwise
func (s *Statement) Validate() error {
	if (s.Principal != nil) && (s.NotPrincipal != nil) {
		return fmt.Errorf("must supply exactly zero or one of (Principal | NotPrincipal)")
	}
	if (s.Action == nil) == (s.NotAction == nil) {
		return fmt.Errorf("must supply exactly one of (Action | NotAction)")
	}
	if (s.Resource == nil) == (s.NotResource == nil) {
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

// TODO(nsiow): Implement deeper validation for Principals

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

	// Handle normal case
	var m map[string]json.RawMessage
	if err := json.Unmarshal(m["AWS"], &p.AWS); err != nil {
		return err
	}
	if err := json.Unmarshal(m["Federated"], &p.Federated); err != nil {
		return err
	}
	if err := json.Unmarshal(m["Service"], &p.Service); err != nil {
		return err
	}
	if err := json.Unmarshal(m["CanonicalUser"], &p.CanonicalUser); err != nil {
		return err
	}
	return nil
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
	return json.Unmarshal(data, &c.Map)
}

// Condition represents the grammar and structure of an AWS IAM Condition
type Condition struct {
	Map map[string]ps.PolyString
}

// UnmarshalJSON instructs how to create Condition fields from raw bytes
func (c *Condition) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.Map)
}
