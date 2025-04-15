package awsconfig

import (
	"encoding/json"

	"github.com/nsiow/yams/pkg/entities"
)

// ConfigItem defines the structure of a generic CI from AWS Config
type ConfigItem struct {
	Type      string          `json:"resourceType"`
	AccountId string          `json:"accountId"`
	Region    string          `json:"awsRegion"`
	Arn       entities.Arn    `json:"arn"`
	Tags      []entities.Tag  `json:"tags"`
	raw       json.RawMessage `json:"-"`
}

func (c *ConfigItem) UnmarshalJSON(data []byte) error {
	type alias *ConfigItem
	a := alias(c)
	err := json.Unmarshal(data, &a)
	if err != nil {
		return err
	}

	*c = ConfigItem(*a)
	c.raw = data
	return nil
}
