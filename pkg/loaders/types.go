package loaders

import "github.com/nsiow/yams/pkg/entities"

// Loader defines the interface for a struct that loads entities from some data source
type Loader interface {
	Principals() []entities.Principal
	Resources() []entities.Resource
}
