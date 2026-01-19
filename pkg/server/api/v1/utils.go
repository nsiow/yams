package v1

import (
	"net/http"
	"strings"

	"github.com/nsiow/yams/pkg/server/httputil"
)

// globalNamespaceTypes is a list of AWS resource types that have global namespaces,
// meaning the account ID cannot be derived from the ARN alone.
var globalNamespaceTypes = []string{
	"AWS::S3::Bucket",
}

// UtilAccountNames returns a mapping of account IDs to account names.
// GET /api/v1/utils/accounts/names
func (api *API) UtilAccountNames(w http.ResponseWriter, req *http.Request) {
	names := make(map[string]string)
	for account := range api.Simulator.Universe.Accounts() {
		names[account.Id] = account.Name
	}
	httputil.WriteJsonResponse(w, req, names)
}

// UtilResourceAccounts returns a mapping of resource ARNs to account IDs for resources
// with global namespaces (e.g., S3 buckets) where the account ID cannot be parsed from the ARN.
// GET /api/v1/utils/resources/accounts
func (api *API) UtilResourceAccounts(w http.ResponseWriter, req *http.Request) {
	mapping := make(map[string]string)
	for resource := range api.Simulator.Universe.Resources() {
		if isGlobalNamespaceType(resource.Type) && resource.AccountId != "" {
			mapping[resource.Key()] = resource.AccountId
		}
	}
	httputil.WriteJsonResponse(w, req, mapping)
}

// isGlobalNamespaceType checks if a resource type has a global namespace.
func isGlobalNamespaceType(resourceType string) bool {
	for _, t := range globalNamespaceTypes {
		if strings.EqualFold(resourceType, t) {
			return true
		}
	}
	return false
}
