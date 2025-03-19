package sar

import (
	"fmt"
	"iter"
	"strings"

	"github.com/nsiow/yams/internal/assets"
	"github.com/nsiow/yams/internal/log"
	"github.com/nsiow/yams/pkg/entities"
)

var LOG = log.Logger("sar")

// data is a local alias hiding the asset implementation of SAR data
var data func() map[string][]entities.ApiCall = assets.SarData

// Map returns a map with format key=service, value=action list for all AWS APIs
func Map() map[string][]entities.ApiCall {
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

// query is a struct used to represent the internal state of a SAR query
type Query struct {
	predicates map[string]predicate
}

// String returns a human-readable representation of the query
func (q *Query) String() string {
	var pkeys []string
	for key := range q.predicates {
		pkeys = append(pkeys, key)
	}

	return fmt.Sprintf("Query[%s]", strings.Join(pkeys, " "))
}

// predicate represents a conditional filter to be applied to an AWS API call definition
type predicate = func(entities.ApiCall) bool

// check filters the provided API call definition through the provided predicates and returns
// whether or not ALL of the predicates are matched
func check(apicall entities.ApiCall, predicates map[string]predicate) bool {
	for pkey, p := range predicates {
		result := p(apicall)
		LOG.Debug("predicate check",
			"predicate_key", pkey,
			"predicate_result", result)
		if !result {
			return false
		}
	}
	return true
}

// NewQuery constructs a new query with no base filters
func NewQuery() *Query {
	return &Query{}
}

// WithService adds a new filter to the query filtering on the "service" field
func (q *Query) WithService(service string) *Query {
	key := "service/" + service
	q.predicates[key] = func(a entities.ApiCall) bool { return a.Service == service }
	return q
}

// WithName adds a new filter to the query filtering on the "name" field
func (q *Query) WithName(name string) *Query {
	key := "name/" + name
	q.predicates[key] = func(a entities.ApiCall) bool { return a.Action == name }
	return q
}

// WithAccessLevel adds a new filter to the query filtering on the "access_level" field
func (q *Query) WithAccessLevel(level string) *Query {
	key := "access_level/" + level
	q.predicates[key] = func(a entities.ApiCall) bool { return a.AccessLevel == level }
	return q
}

// Results executes the query and returns all matching API calls
func (q *Query) Results() (results []entities.ApiCall) {
	for call := range All() {
		if check(call, q.predicates) {
			results = append(results, call)
		}
	}

	return results
}
