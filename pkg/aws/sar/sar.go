package sar

import (
	"fmt"
	"strings"

	"github.com/nsiow/yams/internal/assets"
	"github.com/nsiow/yams/pkg/aws/types"
)

// Type aliases for service authorization semantics
type predicateType = string
type predicateKey = string

// sar is a local alias hiding the asset implementation of SAR
var sar = assets.SAR

// sarIndex is a local alias hiding the asset implementation of SAR
var sarIndex = assets.SARIndex

// Index returns a double-map with structure <service>/<action>/types.Action
func Index() map[string]map[string]types.Action {
	return assets.SARIndex()
}

// Lookup allows for querying a specific api call based on service + action name
func Lookup(service, action string) (types.Action, bool) {
	// SAR index uses lower-case keys
	service = strings.ToLower(service)
	action = strings.ToLower(action)

	idx := sarIndex()
	if actionMap, exists := idx[service]; exists {
		if apicall, exists := actionMap[action]; exists {
			return apicall, true
		}
	}

	return types.Action{}, false
}

// LookupString allows for querying a specific api call based on the "service:action" shorthand
func LookupString(serviceAction string) (types.Action, bool) {
	components := strings.SplitN(serviceAction, ":", 2)
	if len(components) != 2 {
		return types.Action{}, false
	}

	return Lookup(components[0], components[1])
}

// MustLookupString allows for querying a specific api call based on the "service:action" with
// such great confidence that we will panic if we do not find it
func MustLookupString(serviceAction string) types.Action {
	if a, ok := LookupString(serviceAction); ok {
		return a
	}

	panic(fmt.Sprintf("unable to resolve service:action from SAR: '%s'", serviceAction))
}

// query is a struct used to represent the internal state of a SAR query
type Query struct {
	predicates map[predicateType]map[predicateKey]predicate
}

// NewQuery constructs a new query with no base filters
func NewQuery() *Query {
	return &Query{predicates: make(map[predicateType]map[predicateKey]predicate)}
}

// String returns a human-readable representation of the query
func (q *Query) String() string {
	var pkeys []string
	for key := range q.predicates {
		pkeys = append(pkeys, key)
	}

	return fmt.Sprintf("Query[%s]", strings.Join(pkeys, " "))
}

func (q *Query) add(service, key string, pred predicate) *Query {
	_, ok := q.predicates[service]
	if !ok {
		q.predicates[service] = make(map[predicateKey]predicate)
	}

	q.predicates[service][key] = pred

	return q
}

// check filters the provided API call definition through the provided predicates and returns
// whether or not ALL of the predicates are matched
func (q *Query) check(action types.Action) bool {
	if len(q.predicates) == 0 {
		return false
	}

	for _, filters := range q.predicates {
		matchedAny := false
		for _, predicate := range filters {
			match := predicate(action)
			if match {
				matchedAny = true
				break
			}
		}

		if !matchedAny {
			return false
		}
	}

	return true
}

// TODO(nsiow) add wildcard matching to querying

// predicate represents a conditional filter to be applied to an AWS API call definition
type predicate = func(types.Action) bool

// WithService adds a new filter to the query filtering on the "service" field
func (q *Query) WithService(service string) *Query {
	return q.add("service", service, func(a types.Action) bool {
		return strings.EqualFold(a.Service, service)
	})
}

// WithName adds a new filter to the query filtering on the "name" field
func (q *Query) WithName(name string) *Query {
	return q.add("name", name, func(a types.Action) bool {
		return strings.EqualFold(a.Name, name)
	})
}

// Results executes the query and returns all matching API calls
func (q *Query) Results() (results []types.Action) {
	for _, service := range sar() {
		for _, action := range service.Actions {
			if q.check(action) {
				results = append(results, action)
			}
		}
	}

	return results
}
