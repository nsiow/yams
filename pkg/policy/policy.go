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
			return fmt.Errorf("error in single-statement clause of Statement:\ndata=%s\nerror=%v", string(data), err)
		}

		*s = []Statement{stmt}
		return nil
	}

	// Handle list of statements
	if data[0] == '[' && data[len(data)-1] == ']' {
		var list []Statement
		err := json.Unmarshal(data, &list)
		if err != nil {
			return fmt.Errorf("error in multi-statement clause of Statement:\ndata=%s\nerror=%v", string(data), err)
		}
		*s = list
		return nil
	}

	return fmt.Errorf("not sure how to handle statement block: %s", string(data))
}

// Statement represents the grammar and structure of an AWS IAM Statement
type Statement struct {
	Sid          string
	Principal    PrincipalBlock `json:",omitempty"`
	NotPrincipal PrincipalBlock `json:",omitempty"`
	Effect       string
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

// PrincipalBlock represents a set of Principals, provided in string or map form
type PrincipalBlock = Principal

// Principal represents the grammar and structure of an AWS IAM Principal
type Principal struct {
	AWS           Value `json:",omitempty"`
	Federated     Value `json:",omitempty"`
	Service       Value `json:",omitempty"`
	CanonicalUser Value `json:",omitempty"`
}

// UnmarshalJSON instructs how to create Principal fields from raw bytes
func (p *PrincipalBlock) UnmarshalJSON(data []byte) error {
	// Handle string case; only valid in this 3-byte sequence
	if len(data) == 3 && string(data) == `"*"` {
		p.AWS = []string{"*"}
		p.Federated = []string{"*"}
		p.Service = []string{"*"}
		p.CanonicalUser = []string{"*"}
		return nil
	}

	var principal Principal
	err := json.Unmarshal(data, &principal)
	if err != nil {
		return fmt.Errorf("unable to parse:\nprincipal = %s\nerror = %v", string(data), err)
	}

	*p = principal
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
	Value
}

// Resource represents the grammar and structure of an AWS IAM Resource
type Resource struct {
	Value
}

// ConditionBlock represents the grammar and structure of an AWS IAM Condition block
type ConditionBlock map[ConditionOperation]Condition

// ConditionOperation represents the operation portion of a condition block
type ConditionOperation string

// Condition represents the grammar and structure of an AWS IAM Condition
type Condition map[ConditionKey]ConditionValue

// ConditionKey represents the key portion of a condition
type ConditionKey string

// ConditionValue represents the value portion of a condition
type ConditionValue struct {
	bools   []bool
	numbers []int
	strings []string
}

// nTrue counts the booleans input and returns the number of them that are true
func nTrue(b ...bool) int {
	n := 0
	for _, v := range b {
		if v {
			n++
		}
	}
	return n
}

// Validate confirms that we have one and only one type of value
func (c *ConditionValue) Validate() error {
	nTrue := nTrue(len(c.bools) > 0, len(c.numbers) > 0, len(c.strings) > 0)
	if nTrue > 1 {
		return fmt.Errorf("multiple (%d) types observed in condition value: %+v", nTrue, c)
	}
	return nil
}

// Bools returns the contained values if we have bools; otherwise errors
func (c *ConditionValue) Bools() ([]bool, error) {
	err := c.Validate()
	if err != nil {
		return nil, err
	}
	return c.bools, nil
}

// Numbers returns the contained values if we have numbers; otherwise errors
func (c *ConditionValue) Numbers() ([]int, error) {
	err := c.Validate()
	if err != nil {
		return nil, err
	}
	return c.numbers, nil
}

// Strings returns the contained values if we have strings; otherwise errors
func (c *ConditionValue) Strings() ([]string, error) {
	err := c.Validate()
	if err != nil {
		return nil, err
	}
	return c.strings, nil
}

// MarshalJSON instructs how to create raw bytes from ConditionValue fields
func (c *ConditionValue) MarshalJSON() ([]byte, error) {
	var items []any

	for _, x := range c.bools {
		items = append(items, x)
	}
	for _, x := range c.numbers {
		items = append(items, x)
	}
	for _, x := range c.strings {
		items = append(items, x)
	}

	return json.Marshal(items)
}

// UnmarshalJSON instructs how to create ConditionValue fields from raw bytes
func (c *ConditionValue) UnmarshalJSON(data []byte) error {
	// First make sure the data can be marshalled at all
	var raw any
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return fmt.Errorf("unable to parse:\nconditionValue = %s\nerror = %v", string(data), err)
	}

	// Handle the different cases between both types
	switch cast := raw.(type) {
	case bool:
		c.bools = []bool{cast}
	case int:
		c.numbers = []int{cast}
	case string:
		c.strings = []string{cast}
	case []any:
		// Otherwise iterate through and fill out arrays; we'll check for homogeneity later
		for _, a := range cast {
			switch item := a.(type) {
			case bool:
				c.bools = append(c.bools, item)
			case int:
				c.numbers = append(c.numbers, item)
			case string:
				c.strings = append(c.strings, item)
			default:
				return fmt.Errorf("unsure how to handle type '%T' for condition value array: %v", a, a)
			}
		}
	case nil:
		break
	default:
		return fmt.Errorf("unsure how to handle type '%T' for condition value: %v", cast, cast)
	}

	return nil
}
