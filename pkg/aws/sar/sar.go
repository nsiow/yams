package sar

import (
	"fmt"
	"iter"
	"strings"

	"github.com/nsiow/yams/internal/assets"
	"github.com/nsiow/yams/pkg/entities"
)

// Constants for access levels
const (
	ACCESS_LEVEL_WRITE       = "Write"
	ACCESS_LEVEL_READ        = "Read"
	ACCESS_LEVEL_LIST        = "List"
	ACCESS_LEVEL_TAGGING     = "Tagging"
	ACCESS_LEVEL_PERMISSIONS = "Permissions management"
)

// Type aliases for service authorization semantics
type predicateType = string
type predicateKey = string

// data is a local alias hiding the asset implementation of SAR data
var data = assets.SarData

// Map returns a map with format key=service, value=action list for all AWS APIs
func Map() map[string]map[string]entities.ApiCall {
	return data()
}

// All returns a slice containing all known AWS API calls
func All() iter.Seq[entities.ApiCall] {
	return func(yield func(entities.ApiCall) bool) {
		for _, callList := range data() {
			for _, call := range callList {
				if !yield(call) {
					return
				}
			}
		}
	}
}

// Lookup allows for querying a specific api call based on service + action name
func Lookup(service, action string) (entities.ApiCall, bool) {
	data := data()
	if actionMap, exists := data[service]; exists {
		if apicall, exists := actionMap[action]; exists {
			return apicall, true
		}
	}

	return entities.ApiCall{}, false
}

// LookupString allows for querying a specific api call based on the "service:action" shorthand
func LookupString(serviceAction string) (entities.ApiCall, bool) {
	components := strings.SplitN(serviceAction, ":", 1)
	if len(components) != 2 {
		return entities.ApiCall{}, false
	}

	return Lookup(components[0], components[1])
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
func (q *Query) check(apicall entities.ApiCall) bool {
	if len(q.predicates) == 0 {
		return false
	}

	for _, filters := range q.predicates {
		matchedAny := false
		for _, predicate := range filters {
			match := predicate(apicall)
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

// predicate represents a conditional filter to be applied to an AWS API call definition
type predicate = func(entities.ApiCall) bool

// WithService adds a new filter to the query filtering on the "service" field
// TODO(nsiow) add wildcard matching to querying
func (q *Query) WithService(service string) *Query {
	return q.add("service", service, func(a entities.ApiCall) bool {
		return strings.EqualFold(a.Service, service)
	})
}

// WithName adds a new filter to the query filtering on the "name" field
func (q *Query) WithName(name string) *Query {
	return q.add("name", name, func(a entities.ApiCall) bool {
		return strings.EqualFold(a.Action, name)
	})
}

// WithAccessLevel adds a new filter to the query filtering on the "access_level" field
func (q *Query) WithAccessLevel(level string) *Query {
	return q.add("access_level", level, func(a entities.ApiCall) bool {
		return strings.EqualFold(a.AccessLevel, level)
	})
}

// Results executes the query and returns all matching API calls
func (q *Query) Results() (results []entities.ApiCall) {
	for call := range All() {
		if q.check(call) {
			results = append(results, call)
		}
	}

	return results
}
