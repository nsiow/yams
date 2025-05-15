package policy

import (
	"encoding/json"
	"fmt"
)

// -------------------------------------------------------------------------------------------------
// Helper functions
// -------------------------------------------------------------------------------------------------

func FromJson(data []byte) (Policy, error) {
	var p Policy

	err := json.Unmarshal(data, &p)
	if err != nil {
		return Policy{}, err
	}

	return p, nil
}

func FromJsonString(data string) (Policy, error) {
	return FromJson([]byte(data))
}

// -------------------------------------------------------------------------------------------------
// Custom Marshal/Unmarshal functions
// -------------------------------------------------------------------------------------------------

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

// MarshalJSON instructs how to convert Principal fields to raw bytes
func (p *Principal) MarshalJSON() ([]byte, error) {
	if p.All {
		return []byte(`"*"`), nil
	}

	type alias Principal
	a := alias(*p)
	return json.Marshal(a)
}

// UnmarshalJSON instructs how to create Principal fields from raw bytes
func (p *Principal) UnmarshalJSON(data []byte) error {
	// Handle string case; only valid in this 3-byte sequence
	if len(data) == 3 && string(data) == `"*"` {
		p.All = true
		return nil
	}

	type alias Principal
	a := alias(*p)
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

// IsZero is used for marshaling to indicate when the field should be omitted
func (p *Principal) IsZero() bool {
	return p.Empty()
}
