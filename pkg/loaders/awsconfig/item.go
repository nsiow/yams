package awsconfig

import (
	"encoding/json"

	"github.com/nsiow/yams/pkg/entities"
)

// ConfigItem defines the structure of a generic CI from AWS Config
type ConfigItem struct {
	Type                       string          `json:"resourceType"`
	Account                    string          `json:"accountId"`
	Region                     string          `json:"awsRegion"`
	Arn                        string          `json:"arn"`
	Tags                       entities.Tags   `json:"tags"`
	Configuration              json.RawMessage `json:"configuration"`
	SupplementaryConfiguration json.RawMessage `json:"supplementaryConfiguration"`
}
