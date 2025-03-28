package awsconfig

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/entities"
)

// loadResources takes a list of AWS Config items and extracts resources
func loadResources(items []ConfigItem) ([]entities.Resource, error) {
	var rs []entities.Resource

	// Iterate through our AWS Config items
	for _, i := range items {

		// Skip non-AWS resources
		if !strings.HasPrefix(i.Type, "AWS") {
			continue
		}

		// TODO(nsiow) give similar treatment to errors for other entities
		// Load the single r
		r, err := loadResource(i)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error loading resource '%s'", i.Arn), err)
		}

		rs = append(rs, r)
	}

	return rs, nil
}

// loadResource takes a single AWS Config item and returns a parsed resource object
func loadResource(i ConfigItem) (entities.Resource, error) {
	// Construct basic fields
	r := entities.Resource{
		Type:    i.Type,
		Account: i.AccountId,
		Region:  i.Region,
		Arn:     i.Arn,
		Tags:    i.Tags,
	}

	// Add policy where supported
	p, err := extractResourcePolicy(i)
	if err != nil {
		return r, err
	}
	r.Policy = p

	return r, nil
}
