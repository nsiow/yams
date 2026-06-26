package policy

import (
	stdjson "encoding/json"
	"fmt"
	"strings"

	json "github.com/bytedance/sonic"
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

// UnmarshalJSON instructs how to create Statement fields from raw bytes
func (s *Statement) UnmarshalJSON(data []byte) error {
	if err := validateStatementValues(data); err != nil {
		return err
	}

	type alias Statement
	var a alias
	err := json.Unmarshal(data, &a)
	if err != nil {
		return fmt.Errorf("unable to parse:\nstatement = %s\nerror = %v", data, err)
	}

	*s = Statement(a)
	return nil
}

func validateStatementValues(data []byte) error {
	var raw map[string]stdjson.RawMessage
	if err := stdjson.Unmarshal(data, &raw); err != nil {
		return nil
	}

	for _, field := range []string{"Action", "NotAction", "Resource", "NotResource"} {
		value, ok := raw[field]
		if ok && !valueIsString(value) {
			return fmt.Errorf("statement value for %s must be string or []string", field)
		}
	}
	return nil
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

	if err := validatePrincipalValues(data); err != nil {
		return err
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

func validatePrincipalValues(data []byte) error {
	var raw map[string]stdjson.RawMessage
	if err := stdjson.Unmarshal(data, &raw); err != nil {
		return nil
	}

	for key, value := range raw {
		if !valueIsString(value) {
			return fmt.Errorf("principal value for %s must be string or []string", key)
		}
	}
	return nil
}

func valueIsString(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, `"`) {
		return true
	}
	if !strings.HasPrefix(trimmed, "[") {
		return false
	}

	var raw []stdjson.RawMessage
	if err := stdjson.Unmarshal(data, &raw); err != nil {
		return false
	}
	for _, item := range raw {
		if !strings.HasPrefix(strings.TrimSpace(string(item)), `"`) {
			return false
		}
	}
	return true
}

// IsZero is used for marshaling to indicate when the field should be omitted
func (p *Principal) IsZero() bool {
	return p.Empty()
}
