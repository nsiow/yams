package sim

import (
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// AuthContext defines the tertiary context of a request that can be used for authz decisions
type AuthContext struct {
	Action     string
	Principal  *entities.Principal
	Resource   *entities.Resource
	Properties map[string]policy.Value
}

func (a *AuthContext) Key(key string) string {
	return ""
}
