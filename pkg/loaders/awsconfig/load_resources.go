package awsconfig

import (
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/entities"
)

// loadResources takes a list of AWS Config items and extracts resources
func loadResources(items []ConfigItem, accounts *AccountMap) ([]entities.Resource, error) {
	var resources []entities.Resource

	// Iterate through our AWS Config items
	for _, i := range items {

		// Skip non-AWS resources
		if !strings.HasPrefix(i.Type, "AWS") {
			continue
		}

		r, err := loadResource(i)
		if err != nil {
			return nil, fmt.Errorf("error loading resource '%s': %w", i.Arn, err)
		}

		resources = append(resources, r)
	}

	return resources, nil
}

// loadResource takes a single AWS Config item and returns a parsed resource object
func loadResource(i ConfigItem) (entities.Resource, error) {
	// Construct basic fields
	r := entities.Resource{
		Type:      i.Type,
		AccountId: i.AccountId,
		Region:    i.Region,
		Arn:       i.Arn,
		Tags:      i.Tags,
	}

	// Add policy where supported
	p, err := extractResourcePolicy(i)
	if err != nil {
		return r, err
	}
	r.Policy = p

	return r, nil
}
