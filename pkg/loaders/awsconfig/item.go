package awsconfig

import (
	"slices"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/entities"
)

// ConfigItem defines the structure of a generic CI from AWS Config
type ConfigItem struct {
	Type      string         `json:"resourceType"`
	Name      string         `json:"resourceName"`
	AccountId string         `json:"accountId"`
	Region    string         `json:"awsRegion"`
	Arn       entities.Arn   `json:"arn"`
	Tags      []entities.Tag `json:"tags,omitzero"`
}

// configBlob is an internal-only struct used for multi-stage JSON unmarshalling
//
// When unmarshalling from JSON, it allows us to peek at the type of the config item before
// delegating to a more specialized handler
type configBlob struct {
	Type string `json:"resourceType"`
	raw  []byte `json:"-"`
}

func (c *configBlob) UnmarshalJSON(data []byte) error {
	type alias configBlob
	var a alias
	err := json.Unmarshal(data, &a)
	if err != nil {
		return err
	}

	c.Type = a.Type
	c.raw = slices.Clone(data)
	return nil
}
