#!/usr/bin/env python3

import json
import logging
import os
import sys

import boto3
import joblib

# Set up logging
logging.basicConfig(level=os.environ.get('YAMS_LOG_LEVEL', 'INFO').upper(),
                    stream=sys.stdout)
for name in ['boto', 'botocore', 'urllib3', 's3transfer']:
    logging.getLogger(name).setLevel(logging.WARNING)

# Set up cache
mem = joblib.Memory('/tmp/orgcrawl.cache')

# Set up client
orgclient = boto3.client('organizations')

@mem.cache
def get_org_root() -> str:
    """Returns the ID of the organization root."""
    resp = orgclient.list_roots()
    roots = resp['Roots']

    if len(roots) != 1:
        raise ValueError(f'unexpected organizations.listRoots return: {resp}')

    return roots[0]['Id']

@mem.cache
def get_org_structure(org_root_id: str) -> dict:
    """Walk through the org structure and return a representation of the root/ou/acct hierarchy.

    At the end of evaluation, `structure` contains a map of accounts to their position within the
    organization, e.g.

    {
      "111111111111": [
        "<root_id>",
        "<ou_1>",
        "<ou_2>",
        ...
        "111111111111",
      ],
      "222222222222": [
        "<root_id>",
        "<ou_1>",
        "<ou_2>",
        "<ou_3>",
        ...
        "222222222222",
      ],
      ...
    }
    """
    q = [org_root_id]
    structure = {}

    while q:
        node = q.pop(0)
        path = node.split('/')
        base = path[-1]

        # Base case - we reached an account
        if base.isnumeric():
            structure[base] = path
            continue

        # Otherwise list children and add them to the queue
        for type_ in ['ORGANIZATIONAL_UNIT', 'ACCOUNT']:
            resp = orgclient.list_children(ParentId=base, ChildType=type_)
            children = resp['Children']
            for child in children:
                id_ = child['Id']
                path = '/'.join([node, id_])
                q.append(path)

    return structure

@mem.cache
def get_policy_structure(org_structure: dict) -> dict:
    """Traverse the org structure and for each node, look up the attached policies.

    At the end of evaluation, `structure` contains a map of org entities to the policies attached
    to that entity, e.g.

    {
      "r-123": [
        "p-12345",
      ],
      "111111111111": [
        "p-11111",
      ],
      "ou-123": [
        "p-22222",
        "p-33333",
      ],
      ...
    }
    """
    structure = {}

    for target_list in org_structure.values():
        for target_id in target_list:
            if target_id not in structure:
                resp = orgclient.list_policies_for_target(TargetId=target_id,
                                                          Filter='SERVICE_CONTROL_POLICY')
                structure[target_id] = [p['Id'] for p in resp['Policies']]

    return structure

@mem.cache
def get_policies(policy_structure: dict) -> dict:
    """Traverse the attached policies and describe each of them.

    At the end of evaluation, `policies` contains a map of policy IDs to definitions
    of that entity, e.g.

    {
      "p-12345": {
        "PolicySummary": {
            "Id": "p-12345",
            "Arn": "arn:aws:organizations::111111111111:policy/o-aaaaaaaaaa/service_control_policy/p-aaaaaaaa",
            "Name": "Example",
            "Description": "An example",
            "Type": "SERVICE_CONTROL_POLICY",
            "AwsManaged": false
        },
        "Content": "<json string of policy contents>"
      }
    }
    """
    policies = {}

    for policy_list in policy_structure.values():
        for policy_id in policy_list:
            if policy_id not in policies:
                resp = orgclient.describe_policy(PolicyId=policy_id)
                policies[policy_id] = resp['Policy']

    return policies

def as_config_blobs(org_id: str,
                    org_structure: dict,
                    policy_structure: dict, policy_data: dict) -> list[dict]:
    """Combine the org + policy structure into a parsable format."""
    data = []

    for account, ou_path in org_structure.items():
        policies = []
        for node in ou_path:
            node_policies = []
            for policy_id in policy_structure[node]:
                policy = policy_data[policy_id]
                node_policies.append(policy)
            policies.append(node_policies)

        parsed_policies = []
        for policy_level in policies:
            parsed_policy_level = []
            for policy_obj in policy_level:
                policy_summary = policy_obj['PolicySummary']
                policy_content = policy_obj['Content']
                policy = json.loads(policy_content)
                policy['Id'] = get_policy_id(policy_summary)
                parsed_policy_level.append(policy)
            parsed_policies.append(parsed_policy_level)

        account_config_blob = {
            'arn': f'arn:yams:::account/{account}',
            'resourceType': 'Yams::Organizations::Account',
            'accountId': account,
            'configuration': {
              'orgId': org_id,
              'orgPaths': ou_path,
              'serviceControlPolicies': parsed_policies,
            },
        }
        data.append(account_config_blob)

    return data

def get_policy_id(policy_summary: dict) -> str:
    """Return a useful, human-readable identifier for the provided policy."""
    return f'{policy_summary["Arn"]}/{policy_summary["Name"]}'

def main():
    org_root_id = get_org_root()
    print(f'[✓] Discovered root: {org_root_id}')
    org_structure = get_org_structure(org_root_id)
    logging.debug('org structure: %s', org_structure)
    print(f'[✓] Discovered org structure for {len(org_structure)} entities')
    policy_structure = get_policy_structure(org_structure)
    logging.debug('policy structure: %s', policy_structure)
    print(f'[✓] Discovered policy structure for {len(policy_structure)} entities')
    policy_data = get_policies(policy_structure)
    logging.debug('policy data: %s', policy_data)
    print(f'[✓] Discovered details for {len(policy_data)} policies')
    data = as_config_blobs(org_root_id, org_structure, policy_structure, policy_data)
    with open('orgdump.json', 'w+') as f:
        json.dump(data, f, indent=2, sort_keys=True)
    print(f'[✓] Wrote org crawl results to: orgdump.json')

if __name__ == '__main__':
    main()
