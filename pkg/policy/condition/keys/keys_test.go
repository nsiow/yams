package keys

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestIsGlobalConditionKey(t *testing.T) {
	tests := []testlib.TestCase[string, bool]{
		// Principal keys
		{Name: "principal_arn", Input: PrincipalArn, Want: true},
		{Name: "principal_account", Input: PrincipalAccount, Want: true},
		{Name: "principal_org_paths", Input: PrincipalOrgPaths, Want: true},
		{Name: "principal_org_id", Input: PrincipalOrgId, Want: true},
		{Name: "principal_is_aws_service", Input: PrincipalIsAwsService, Want: true},
		{Name: "principal_service_name", Input: PrincipalServiceName, Want: true},
		{Name: "principal_service_names_list", Input: PrincipalServiceNamesList, Want: true},
		{Name: "principal_type", Input: PrincipalType, Want: true},
		{Name: "user_id", Input: UserId, Want: true},
		{Name: "username", Input: Username, Want: true},

		// Auth keys
		{Name: "federated_provider", Input: FederatedProvider, Want: true},
		{Name: "token_issue_time", Input: TokenIssueTime, Want: true},
		{Name: "mfa_age", Input: MultiFactorAuthAge, Want: true},
		{Name: "mfa_present", Input: MultiFactorAuthPresent, Want: true},
		{Name: "ec2_source_vpc", Input: Ec2InstanceSourceVpc, Want: true},
		{Name: "ec2_source_ipv4", Input: Ec2InstanceSourcePrivateIPv4, Want: true},
		{Name: "source_identity", Input: SourceIdentity, Want: true},
		{Name: "role_delivery", Input: RoleDelivery, Want: true},
		{Name: "source_instance_arn", Input: SourceInstanceArn, Want: true},

		// Network keys
		{Name: "source_ip", Input: SourceIp, Want: true},
		{Name: "source_vpc", Input: SourceVpc, Want: true},
		{Name: "source_vpce", Input: SourceVpce, Want: true},
		{Name: "vpc_source_ip", Input: VpcSourceIp, Want: true},

		// Resource keys
		{Name: "resource_account", Input: ResourceAccount, Want: true},
		{Name: "resource_org_paths", Input: ResourceOrgPaths, Want: true},
		{Name: "resource_org_id", Input: ResourceOrgId, Want: true},

		// Service chain keys
		{Name: "called_via", Input: CalledVia, Want: true},
		{Name: "called_via_first", Input: CalledViaFirst, Want: true},
		{Name: "called_via_last", Input: CalledViaLast, Want: true},
		{Name: "via_aws_service", Input: ViaAwsService, Want: true},

		// Time and misc keys
		{Name: "current_time", Input: CurrentTime, Want: true},
		{Name: "epoch_time", Input: EpochTime, Want: true},
		{Name: "referer", Input: Referer, Want: true},
		{Name: "requested_region", Input: RequestedRegion, Want: true},
		{Name: "tag_keys", Input: TagKeys, Want: true},
		{Name: "secure_transport", Input: SecureTransport, Want: true},
		{Name: "source_arn", Input: SourceArn, Want: true},
		{Name: "source_account", Input: SourceAccount, Want: true},
		{Name: "source_org_paths", Input: SourceOrgPaths, Want: true},
		{Name: "source_org_id", Input: SourceOrgId, Want: true},
		{Name: "user_agent", Input: UserAgent, Want: true},

		// Case insensitivity tests
		{Name: "uppercase", Input: "AWS:PRINCIPALARN", Want: true},
		{Name: "mixed_case", Input: "AWS:PrincipalArn", Want: true},

		// Non-global keys
		{Name: "not_global", Input: "s3:prefix", Want: false},
		{Name: "empty", Input: "", Want: false},
		{Name: "random_string", Input: "some-random-key", Want: false},

		// Tag prefixes are NOT global keys by themselves
		{Name: "principal_tag_prefix", Input: PrincipalTagPrefix, Want: false},
		{Name: "resource_tag_prefix", Input: ResourceTagPrefix, Want: false},
		{Name: "request_tag_prefix", Input: RequestTagPrefix, Want: false},
	}

	testlib.RunTestSuite(t, tests, func(key string) (bool, error) {
		return IsGlobalConditionKey(key), nil
	})
}

func TestGlobalConditionKeyConstants(t *testing.T) {
	// Verify expected values for key constants
	if PrincipalArn != "aws:principalarn" {
		t.Errorf("PrincipalArn expected 'aws:principalarn', got '%s'", PrincipalArn)
	}
	if PrincipalAccount != "aws:principalaccount" {
		t.Errorf("PrincipalAccount expected 'aws:principalaccount', got '%s'", PrincipalAccount)
	}
	if SourceIp != "aws:sourceip" {
		t.Errorf("SourceIp expected 'aws:sourceip', got '%s'", SourceIp)
	}
	if CurrentTime != "aws:currenttime" {
		t.Errorf("CurrentTime expected 'aws:currenttime', got '%s'", CurrentTime)
	}
}

func TestTagPrefixConstants(t *testing.T) {
	if PrincipalTagPrefix != "aws:principaltag" {
		t.Errorf("PrincipalTagPrefix expected 'aws:principaltag', got '%s'", PrincipalTagPrefix)
	}
	if ResourceTagPrefix != "aws:resourcetag" {
		t.Errorf("ResourceTagPrefix expected 'aws:resourcetag', got '%s'", ResourceTagPrefix)
	}
	if RequestTagPrefix != "aws:requesttag" {
		t.Errorf("RequestTagPrefix expected 'aws:requesttag', got '%s'", RequestTagPrefix)
	}
}
