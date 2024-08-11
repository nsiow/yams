#!/usr/bin/env python3

from typing import Dict

import boto3
from botocore.client import BaseClient

def get_org_root(client: BaseClient) -> str:
    """Returns the ID of the organization root."""
    resp = client.list_roots()
    roots = resp['Roots']

    if len(roots) != 1:
        raise ValueError(f'unexpected organizations.ListRoots return: {resp}')

    return roots[0]['Id']

def get_org_structure(client: BaseClient, org_root_id: str) -> Dict:
    """Recurse through the org structure and return a representation of the root/ou/acct hierarchy.

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
            resp = client.list_children(ParentId=base, ChildType=type_)
            children = resp['Children']
            for child in children:
                id_ = child['Id']
                path = '/'.join([node, id_])
                q.append(path)

    return structure

def get_policy_structure(client: BaseClient, org_structure: Dict) -> Dict:
    """Traverse the org structure and for each node, look up the attached policies."""
    structure = {}

    for target_list in org_structure.values():
        for target_id in target_list:
            if target_id not in structure:
                resp = client.list_policies_for_target(TargetId=target_id,
                                                       Filter='SERVICE_CONTROL_POLICY')
                structure[target_id] = [p['Id'] for p in resp['Policies']]

    return structure

def get_policies(client: BaseClient, policy_structure: Dict) -> Dict:
    """Traverse the attached policies and describe each of them."""
    policies = {}

    for policy_list in policy_structure.values():
        for policy_id in policy_list:
            if policy_id not in policies:
                resp = client.describe_policy(PolicyId=policy_id)
                policies[policy_id] = resp['Policy']

    return policies

def main():
    client = boto3.client('organizations')
    org_root_id = get_org_root(client)
    print(f'[✓] Discovered root: {org_root_id}')
    org_structure = get_org_structure(client, org_root_id)
    print(f'[✓] Discovered org structure for {len(org_structure)} entities')
    policy_structure = get_policy_structure(client, org_structure)
    print(f'[✓] Discovered policy structure for {len(policy_structure)} entities')

if __name__ == '__main__':
    main()
