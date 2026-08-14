package audit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
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

func (w *groupedCSVWriter) WriteGroup(resource, group string, principals []string) error {
	if len(principals) == 0 {
		return w.w.Write([]string{resource, group, ""})
	}
	for _, principal := range principals {
		if err := w.w.Write([]string{resource, group, principal}); err != nil {
			return err
		}
	}
	return nil
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

func (w *groupedJSONLWriter) WriteGroup(resource, group string, principals []string) error {
	if principals == nil {
		principals = []string{}
	}
	return w.w.Encode(groupedJSONLRecord{
		Resource:   resource,
		Group:      group,
		Principals: principals,
	})
}

func (w *groupedJSONLWriter) Flush() error {
	return nil
}

func newGroupedWriter(format string, w io.Writer) (groupedWriter, error) {
	switch format {
	case formatGroupedCSV:
		return newGroupedCSVWriter(w)
	case formatGroupedJSONL:
		return &groupedJSONLWriter{w: json.NewEncoder(w)}, nil
	default:
		return nil, fmt.Errorf("unsupported grouped format %q", format)
	}
}
