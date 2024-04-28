package environment

import (
	"github.com/nsiow/yams/pkg/principal"
	"github.com/nsiow/yams/pkg/resource"
)

// TODO(nsiow) rename this to universe
// An Environment corresponds to a set of coexistant Principals and Resources
type Environment struct {
	Principals []principal.Principal
	Resources  []resource.Resource
}
