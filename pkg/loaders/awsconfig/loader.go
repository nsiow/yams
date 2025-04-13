package awsconfig

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
)

// Loader provides the ability to load entity definitions from AWS Config data
type Loader struct {
	universe *entities.Universe
}

// NewLoader provisions and returns a new `Loader` struct, ready to use
func NewLoader() *Loader {
	return &Loader{}
}

// Universe returns an Universe containing the loaded Principals + Resources
func (l *Loader) Universe() *entities.Universe {
	return l.universe
}

// LoadJson loads data from a provided JSON array
// TODO(nsiow) consider having this load from io.Reader instead
func (l *Loader) LoadJson(data []byte) error {
	var items []ConfigItem
	err := json.Unmarshal(data, &items)
	if err != nil {
		return fmt.Errorf("unable to load data as JSON: %w", err)
	}
	return l.loadItems(items)
}

// LoadJson loads data from the provided newline-separate JSONL input
// TODO(nsiow) consider having this load from io.Reader instead
func (l *Loader) LoadJsonl(data []byte) error {
	r := bytes.NewReader(data)
	s := bufio.NewScanner(r)

	// Some buffer customization, since these JSON blobs can get big
	// TODO(nsiow) move these to constants
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 1024*1024)

	var items []ConfigItem
	for s.Scan() {
		// Read the next line; skip empty lines
		b := s.Bytes()
		if len(b) == 0 {
			continue
		}

		// Unmarshal into a single item
		var i ConfigItem
		err := json.Unmarshal(b, &i)
		if err != nil {
			return err
		}

		// Add to running list of items
		items = append(items, i)
	}

	// If we encountered an error scanning, return it
	if err := s.Err(); err != nil {
		return err
	}

	// Proceed to load
	return l.loadItems(items)
}

// loadItems loads data from the provided AWS Config items
func (l *Loader) loadItems(items []ConfigItem) error {
	for _, item := range items {
		err := l.loadItem(item)
		if err != nil {
			return err
		}
	}

	return nil
}

// loadItem converts the provided ConfigItem into an `entities.*` struct and loads it into the
// universe
func (l *Loader) loadItem(item ConfigItem) error {
	switch item.Type {
	case CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT:
		return l.loadAccount(item)
	default:
		return fmt.Errorf("unsure how to handle config item of type: %s", item.Type)
	}
}

// loadAccount parses the custom Yams::Account item and loads it as an [entities.Account] struct
// TODO(nsiow) figure out correct struct linking for godoc
func (l *Loader) loadAccount(item ConfigItem) error {
}
