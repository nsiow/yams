package sim

import (
	"strings"

	"github.com/nsiow/yams/pkg/arn"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// resourceMayAllowCrossAccount reports whether a resource policy has an allow statement whose
// principal selector could match a principal outside the resource account. Conditions and other
// statement fields are intentionally ignored so this remains a conservative prefilter.
func resourceMayAllowCrossAccount(resource *entities.FrozenResource) bool {
	if resource == nil ||
		resource.AccountId == "" ||
		resource.AccountId == placeholderWildcardAccount {
		return true
	}
	for statementIndex := range resource.Policy.Statement {
		statement := &resource.Policy.Statement[statementIndex]
		if statement.Effect != policy.EFFECT_ALLOW {
			continue
		}
		switch {
		case statement.Principal.All:
			return true
		case statement.NotPrincipal.All:
			continue
		case statement.Principal.Empty():
			return true
		}
		for _, pattern := range statement.Principal.AWS {
			if principalPatternMayMatchOtherAccount(pattern, resource.AccountId) {
				return true
			}
		}
	}
	return false
}

func principalPatternMayMatchOtherAccount(pattern, resourceAccount string) bool {
	if isAccountRootMatch(pattern, resourceAccount) {
		return false
	}
	if strings.ContainsAny(pattern, "*?") {
		return true
	}
	principalAccount := arn.Account(pattern)
	if principalAccount == "" {
		if looksLikeAccountID(pattern) {
			principalAccount = pattern
		} else {
			return true
		}
	}
	return principalAccount != resourceAccount
}

func looksLikeAccountID(value string) bool {
	if len(value) != 12 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
