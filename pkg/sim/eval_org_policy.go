package sim

import (
	"strings"

	"github.com/nsiow/yams/pkg/loaders/awsconfig"
)

func principalIsServiceLinkedRole(s *subject) bool {
	return s.auth.Principal != nil &&
		strings.EqualFold(s.auth.Principal.Type, awsconfig.CONST_TYPE_AWS_IAM_ROLE) &&
		strings.Contains(string(s.auth.Principal.Arn), ":role/aws-service-role/")
}

func principalIsInManagementAccount(s *subject) bool {
	if s.auth.Principal == nil {
		return false
	}
	for _, node := range s.auth.Principal.Account.OrgNodes {
		for _, scp := range node.SCPs {
			if scp.AccountId != "" && scp.AccountId == s.auth.Principal.AccountId {
				return true
			}
		}
	}
	return false
}

func resourceIsInManagementAccount(s *subject) bool {
	if s.auth.Resource == nil {
		return false
	}
	for _, node := range s.auth.Resource.Account.OrgNodes {
		for _, rcp := range node.RCPs {
			if rcp.AccountId != "" && rcp.AccountId == s.auth.Resource.AccountId {
				return true
			}
		}
	}
	return false
}
