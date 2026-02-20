package cli

import (
	"fmt"
	"io"
	"os"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/entities"
	v1 "github.com/nsiow/yams/pkg/server/api/v1"
)

// LoadOverlays reads and decodes overlay entity files into an Overlay struct
func LoadOverlays(files []string) (*v1.Overlay, error) {
	type overlayItem struct {
		Type string
	}

	overlay := v1.Overlay{}

	for _, fn := range files {
		file, err := os.Open(fn)
		if err != nil {
			return nil, fmt.Errorf("could not open overlay file '%s': %v", fn, err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("could not read overlay file '%s': %v", fn, err)
		}

		var item overlayItem
		err = json.Unmarshal(content, &item)
		if err != nil {
			return nil, fmt.Errorf("could not decode overlay file '%s': %v", fn, err)
		}
		if item.Type == "" {
			return nil, fmt.Errorf("could not decode overlay file '%s': missing field 'Type'", fn)
		}

		switch item.Type {
		case "AWS::IAM::Role", "AWS::IAM::User":
			var principal entities.Principal
			err = json.Unmarshal(content, &principal)
			if err != nil {
				return nil, fmt.Errorf("could not decode principal from overlay file '%s': %v", fn, err)
			}
			overlay.Principals = append(overlay.Principals, principal)
		case "AWS::IAM::Group":
			var group entities.Group
			err = json.Unmarshal(content, &group)
			if err != nil {
				return nil, fmt.Errorf("could not decode group from overlay file '%s': %v", fn, err)
			}
			overlay.Groups = append(overlay.Groups, group)
		case
			"AWS::IAM::Policy",
			"Yams::Organizations::ServiceControlPolicy",
			"Yams::Organizations::ResourceControlPolicy":
			var policy entities.ManagedPolicy
			err = json.Unmarshal(content, &policy)
			if err != nil {
				return nil, fmt.Errorf("could not decode policy from overlay file '%s': %v", fn, err)
			}
			overlay.Policies = append(overlay.Policies, policy)
		case "Yams::Organizations::Account":
			var account entities.Account
			err = json.Unmarshal(content, &account)
			if err != nil {
				return nil, fmt.Errorf("could not decode account from overlay file '%s': %v", fn, err)
			}
			overlay.Accounts = append(overlay.Accounts, account)
		}

		var resource entities.Resource
		err = json.Unmarshal(content, &resource)
		if err != nil {
			return nil, fmt.Errorf("could not decode resource from overlay file '%s': %v", fn, err)
		}
		overlay.Resources = append(overlay.Resources, resource)
	}

	return &overlay, nil
}
