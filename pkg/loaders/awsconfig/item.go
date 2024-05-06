package awsconfig

import (
	"encoding/json"

	"github.com/nsiow/yams/pkg/entities"
)

// Item defines the structure of a generic CI from AWS Config
type Item struct {
	Type                       string          `json:"resourceType"`
	Account                    string          `json:"accountId"`
	Region                     string          `json:"awsRegion"`
	Arn                        string          `json:"arn"`
	Tags                       []entities.Tag  `json:"tags"`
	Configuration              json.RawMessage `json:"configuration"`
	SupplementaryConfiguration json.RawMessage `json:"supplementaryConfiguration"`
}
