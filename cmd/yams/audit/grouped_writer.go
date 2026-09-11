package audit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/bits"

	"github.com/nsiow/yams/pkg/entities"
)

type groupedCSVWriter struct {
	w *csv.Writer
}

func newGroupedCSVWriter(w io.Writer) (*groupedCSVWriter, error) {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"resource", "group", "principal"}); err != nil {
		return nil, err
	}
	return &groupedCSVWriter{w: cw}, nil
}

func (w *groupedCSVWriter) WriteResource(
	resource string,
	groups []groupMembership,
	principals []*entities.FrozenPrincipal,
) (int, error) {
	relationships := 0
	for _, group := range groups {
		groupPrincipals := principalsFromBits(group.bits, principals)
		relationships += len(groupPrincipals)
		if len(groupPrincipals) == 0 {
			if err := w.w.Write([]string{resource, group.name, ""}); err != nil {
				return 0, err
			}
			continue
		}
		for _, principal := range groupPrincipals {
			if err := w.w.Write([]string{resource, group.name, principal}); err != nil {
				return 0, err
			}
		}
	}
	return relationships, nil
}

func (w *groupedCSVWriter) Flush() error {
	w.w.Flush()
	return w.w.Error()
}

type groupedJSONLRecord struct {
	Resource   string   `json:"resource"`
	Group      string   `json:"group"`
	Principals []string `json:"principals"`
}

type groupedJSONLWriter struct {
	w *json.Encoder
}

func (w *groupedJSONLWriter) WriteResource(
	resource string,
	groups []groupMembership,
	principals []*entities.FrozenPrincipal,
) (int, error) {
	relationships := 0
	for _, group := range groups {
		groupPrincipals := principalsFromBits(group.bits, principals)
		relationships += len(groupPrincipals)
		if groupPrincipals == nil {
			groupPrincipals = []string{}
		}
		if err := w.w.Encode(groupedJSONLRecord{
			Resource:   resource,
			Group:      group.name,
			Principals: groupPrincipals,
		}); err != nil {
			return 0, err
		}
	}
	return relationships, nil
}

func (w *groupedJSONLWriter) Flush() error {
	return nil
}

type resourceCountJSONLRecord struct {
	Resource       string `json:"resource"`
	PrincipalCount int    `json:"principal_count"`
}

type resourceCountsJSONLWriter struct {
	w     *json.Encoder
	union []uint64
}

func (w *resourceCountsJSONLWriter) WriteResource(
	resource string,
	groups []groupMembership,
	_ []*entities.FrozenPrincipal,
) (int, error) {
	wordCount := 0
	if len(groups) > 0 {
		wordCount = len(groups[0].bits)
	}
	if cap(w.union) < wordCount {
		w.union = make([]uint64, wordCount)
	} else {
		w.union = w.union[:wordCount]
		clear(w.union)
	}

	relationships := 0
	for _, group := range groups {
		for index, word := range group.bits {
			relationships += bits.OnesCount64(word)
			w.union[index] |= word
		}
	}
	principalCount := 0
	for _, word := range w.union {
		principalCount += bits.OnesCount64(word)
	}
	return relationships, w.w.Encode(resourceCountJSONLRecord{
		Resource:       resource,
		PrincipalCount: principalCount,
	})
}

func (w *resourceCountsJSONLWriter) Flush() error {
	return nil
}

func newGroupedWriter(format string, w io.Writer) (groupedWriter, error) {
	switch format {
	case formatGroupedCSV:
		return newGroupedCSVWriter(w)
	case formatGroupedJSONL:
		return &groupedJSONLWriter{w: json.NewEncoder(w)}, nil
	case formatResourceCountsJSONL:
		return &resourceCountsJSONLWriter{w: json.NewEncoder(w)}, nil
	default:
		return nil, fmt.Errorf("unsupported grouped format %q", format)
	}
}
