package environment

import (
	"github.com/nsiow/yams/pkg/entities"
)

// An Universe corresponds to a set of coexistant Principals and Resources
type Universe struct {
	Principals []entities.Principal
	Resources  []entities.Resource
}
