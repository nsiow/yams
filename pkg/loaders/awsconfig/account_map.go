package awsconfig

import "github.com/nsiow/yams/pkg/entities"

// AccountMap contains a mapping from account ID to account metadata
type AccountMap struct {
	mapping map[string]entities.Account
}

// NewAccountMap creates and returns an initialized instance of AccountMap
func NewAccountMap() *AccountMap {
	m := AccountMap{}
	m.mapping = make(map[string]entities.Account)
	return &m
}

// Add creates a new mapping between the provided ID and account
func (m *AccountMap) Add(pType, arn string, account entities.Account) {
	m.mapping[arn] = account
}

// Get retrieves the requested account by ID, if it exists
func (m *AccountMap) Get(pType, arn string) (entities.Account, bool) {
	val, ok := m.mapping[arn]
	return val, ok
}
