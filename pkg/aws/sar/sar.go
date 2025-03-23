package sar

import (
	"fmt"
	"iter"
	"strings"

	"github.com/nsiow/yams/internal/assets"
	"github.com/nsiow/yams/internal/log"
	"github.com/nsiow/yams/pkg/entities"
)

// Type aliases for service authorization semantics
type predicateType = string
type predicateKey = string

var LOG = log.Logger("sar")

// data is a local alias hiding the asset implementation of SAR data
var data = assets.SarData

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

func (q *Query) add(service, key string, pred predicate) {
	_, ok := q.predicates[service]
	if !ok {
		q.predicates[service] = make(map[predicateKey]predicate)
	}

	q.predicates[service][key] = pred
}

// check filters the provided API call definition through the provided predicates and returns
// whether or not ALL of the predicates are matched
func (q *Query) check(apicall entities.ApiCall) bool {
	for predicateType, filters := range q.predicates {
		matchedAny := false
		for predicateKey, predicate := range filters {
			match := predicate(apicall)
			LOG.Debug("predicate check",
				"predicate_type", predicateType,
				"predicate_key", predicateKey,
				"predicate_result", match)
			if match {
				matchedAny = true
				break
			}
		}

		LOG.Debug("predicate type check",
			"predicate_type", predicateType,
			"predicate_type_result", matchedAny)

		if !matchedAny {
			return false
		}
	}

	return true
}

// predicate represents a conditional filter to be applied to an AWS API call definition
type predicate = func(entities.ApiCall) bool

// WithService adds a new filter to the query filtering on the "service" field
func (q *Query) WithService(service string) *Query {
	q.add("service", service, func(a entities.ApiCall) bool { return a.Service == service })
	return q
}

// WithName adds a new filter to the query filtering on the "name" field
func (q *Query) WithName(name string) *Query {
	q.add("name", name, func(a entities.ApiCall) bool { return a.Action == name })
	return q
}

// WithAccessLevel adds a new filter to the query filtering on the "access_level" field
func (q *Query) WithAccessLevel(level string) *Query {
	q.add("access_level", level, func(a entities.ApiCall) bool { return a.AccessLevel == level })
	return q
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
