package sim

import "github.com/nsiow/yams/pkg/entities"

// Event represents the smallest simulatable occurrence
type Event struct {
	Action      string
	Principal   *entities.Principal
	Resource    *entities.Resource
	AuthContext *AuthContext
}
