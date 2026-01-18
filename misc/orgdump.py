#!/usr/bin/env python3
"""
Dump AWS Organizations data (accounts, SCPs, RCPs) in yams-compatible JSON format.

Usage:
    python orgdump.py [output_file]

If no output file is specified, writes to stdout.
Requires AWS credentials with organizations:* read permissions.
"""

import json
import logging
import sys
from typing import Any

import boto3

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(message)s',
    stream=sys.stderr
)
log = logging.getLogger(__name__)


class OrgDumper:
    """Dumps AWS Organizations data to yams-compatible JSON format."""

    def __init__(self):
        self.client = boto3.client('organizations')
        self.cache: dict[str, Any] = {}

    def dump(self) -> list[dict]:
        """Main entry point - returns list of all org entities."""
        accounts = self._walk_org()
        scps = self._describe_policies('SERVICE_CONTROL_POLICY')
        rcps = self._describe_policies('RESOURCE_CONTROL_POLICY')
        return accounts + scps + rcps

    # -------------------------------------------------------------------------
    # Org tree walking
    # -------------------------------------------------------------------------

    def _walk_org(self) -> list[dict]:
        """Walk the org tree and return all accounts with their org context."""
        org = self._describe_org()
        root = self._describe_root()
        log.info(f"Walking org tree from root {root['Id']}")
        return self._walk(org['Id'], [root['Id']])

    def _walk(self, org_id: str, path: list[str]) -> list[dict]:
        """Recursively walk the org tree, collecting accounts."""
        node = path[-1]

        # If this is an account, return it
        if self._is_account(node):
            log.debug(f"Found account {node}")
            return [self._make_account(org_id, path, node)]

        # Otherwise, walk children (accounts and OUs)
        accounts = []
        for child_type in ['ACCOUNT', 'ORGANIZATIONAL_UNIT']:
            paginator = self.client.get_paginator('list_children')
            for page in paginator.paginate(ParentId=node, ChildType=child_type):
                for child in page['Children']:
                    child_accounts = self._walk(org_id, path + [child['Id']])
                    accounts.extend(child_accounts)

        return accounts

    # -------------------------------------------------------------------------
    # AWS Organizations API calls (with caching)
    # -------------------------------------------------------------------------

    def _describe_org(self) -> dict:
        """Get organization details."""
        key = 'org'
        if key not in self.cache:
            resp = self.client.describe_organization()
            self.cache[key] = resp['Organization']
            log.info(f"Found org {self.cache[key]['Id']}")
        return self.cache[key]

    def _describe_root(self) -> dict:
        """Get the organization root."""
        key = 'root'
        if key not in self.cache:
            resp = self.client.list_roots()
            if len(resp['Roots']) != 1:
                raise ValueError(f"Unexpected number of roots: {len(resp['Roots'])}")
            self.cache[key] = resp['Roots'][0]
            log.info(f"Found root {self.cache[key]['Id']}")
        return self.cache[key]

    def _describe_account(self, account_id: str) -> dict:
        """Get account details."""
        key = f'account/{account_id}'
        if key not in self.cache:
            resp = self.client.describe_account(AccountId=account_id)
            self.cache[key] = resp['Account']
        return self.cache[key]

    def _describe_ou(self, ou_id: str) -> dict:
        """Get organizational unit details."""
        key = f'ou/{ou_id}'
        if key not in self.cache:
            resp = self.client.describe_organizational_unit(OrganizationalUnitId=ou_id)
            self.cache[key] = resp['OrganizationalUnit']
        return self.cache[key]

    def _list_policies_for_target(self, target_id: str, policy_type: str) -> list[str]:
        """List policy ARNs attached to a target (account, OU, or root)."""
        key = f'policies_for/{policy_type}/{target_id}'
        if key not in self.cache:
            arns = []
            paginator = self.client.get_paginator('list_policies_for_target')
            try:
                for page in paginator.paginate(TargetId=target_id, Filter=policy_type):
                    for policy in page['Policies']:
                        arns.append(policy['Arn'])
            except self.client.exceptions.ConstraintViolationException:
                # RCPs might not be enabled - return empty list
                arns = []
            self.cache[key] = arns
        return self.cache[key]

    def _describe_policies(self, policy_type: str) -> list[dict]:
        """List and describe all policies of a given type."""
        type_name = {
            'SERVICE_CONTROL_POLICY': 'Yams::Organizations::ServiceControlPolicy',
            'RESOURCE_CONTROL_POLICY': 'Yams::Organizations::ResourceControlPolicy',
        }[policy_type]

        org = self._describe_org()
        policies = []

        log.info(f"Fetching {policy_type} policies")
        paginator = self.client.get_paginator('list_policies')
        try:
            for page in paginator.paginate(Filter=policy_type):
                for policy_summary in page['Policies']:
                    resp = self.client.describe_policy(PolicyId=policy_summary['Id'])
                    policy = resp['Policy']

                    # Parse the policy document
                    document = json.loads(policy['Content'])

                    policies.append({
                        'resourceType': type_name,
                        'resourceName': policy['PolicySummary']['Name'],
                        'accountId': org['MasterAccountId'],
                        'awsRegion': 'global',
                        'arn': policy['PolicySummary']['Arn'],
                        'configuration': {
                            'document': document,
                        },
                    })
        except self.client.exceptions.AccessDeniedException:
            log.warning(f"Access denied listing {policy_type} - skipping")
        except self.client.exceptions.AWSOrganizationsNotInUseException:
            log.warning(f"{policy_type} not enabled - skipping")

        log.info(f"Found {len(policies)} {policy_type} policies")
        return policies

    # -------------------------------------------------------------------------
    # Helper functions
    # -------------------------------------------------------------------------

    def _is_account(self, node_id: str) -> bool:
        """Check if a node ID is an account (numeric) vs OU/root."""
        return node_id.isdigit()

    def _org_paths(self, org_id: str, path: list[str]) -> list[str]:
        """Build org paths from the node path."""
        paths = []
        segment = f"{org_id}/"
        for p in path:
            if not self._is_account(p):
                segment += f"{p}/"
                paths.append(segment)
        return paths

    def _org_node(self, node_id: str) -> dict:
        """Get org node details (root, OU, or account)."""
        if node_id.startswith('r-'):
            root = self._describe_root()
            node_type = 'ROOT'
            node_id = root['Id']
            arn = root['Arn']
            name = root['Name']
        elif self._is_account(node_id):
            account = self._describe_account(node_id)
            node_type = 'ACCOUNT'
            arn = account['Arn']
            name = account['Name']
        else:
            ou = self._describe_ou(node_id)
            node_type = 'ORGANIZATIONAL_UNIT'
            node_id = ou['Id']
            arn = ou['Arn']
            name = ou['Name']

        scps = self._list_policies_for_target(node_id, 'SERVICE_CONTROL_POLICY')
        rcps = self._list_policies_for_target(node_id, 'RESOURCE_CONTROL_POLICY')

        return {
            'id': node_id,
            'type': node_type,
            'arn': arn,
            'name': name,
            'serviceControlPolicies': scps,
            'resourceControlPolicies': rcps,
        }

    def _make_account(self, org_id: str, path: list[str], account_id: str) -> dict:
        """Build account entity from org context."""
        account = self._describe_account(account_id)
        nodes = [self._org_node(node_id) for node_id in path]

        return {
            'resourceType': 'Yams::Organizations::Account',
            'resourceName': '',
            'accountId': account_id,
            'awsRegion': 'global',
            'arn': account['Arn'],
            'configuration': {
                'name': account['Name'],
                'orgId': org_id,
                'orgPaths': self._org_paths(org_id, path),
                'orgNodes': nodes,
            },
        }


def main():
    dumper = OrgDumper()
    entities = dumper.dump()

    log.info(f"Dumped {len(entities)} entities")

    if len(sys.argv) >= 2:
        with open(sys.argv[1], 'w') as f:
            json.dump(entities, f)
        log.info(f"Wrote output to {sys.argv[1]}")
    else:
        json.dump(entities, sys.stdout)


if __name__ == '__main__':
    main()
