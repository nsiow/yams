package environment

import (
	"github.com/nsiow/yams/pkg/entities"
)

// TODO(nsiow) rename this to universe
// An Environment corresponds to a set of coexistant Principals and Resources
type Environment struct {
	Principals []entities.Principal
	Resources  []entities.Resource
}
