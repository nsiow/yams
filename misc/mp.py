#!/usr/bin/env python3

import gzip
import json
import logging
import os
import sys

import boto3
import joblib


# Set up logging
logging.basicConfig(level=os.environ.get('YAMS_LOG_LEVEL', 'INFO').upper(),
                    stream=sys.stdout)

# Set up cache
os.makedirs('.cache', exist_ok=True)
memory = joblib.Memory('.cache/mp.cache')

# Set up client once (cannot cache/pickle)
iam_client = boto3.client('iam')

@memory.cache
def get_policy(arn: str) -> dict:
    """Retrieve details for the specified policy ARN."""
    resp = iam_client.get_policy(PolicyArn=arn)
    policy_metadata = resp['Policy']
    policy_version = policy_metadata['DefaultVersionId']
    resp2 = iam_client.get_policy_version(PolicyArn=arn, VersionId=policy_version)
    policy = resp2['PolicyVersion']
    return {
        'Arn': policy_metadata['Arn'],
        'Name': policy_metadata['PolicyName'],
        'Document': policy['Document'],
    }

@memory.cache
def list_policy_arns() -> list[str]:
    """Using the AWS API, fetch a list of all AWS-owned managed policies."""
    arns = []
    paginator = iam_client.get_paginator('list_policies')
    for page in paginator.paginate(Scope='AWS'):
        for policy in page['Policies']:
            arns.append(policy['Arn'])
    return arns

def main():
    policy_arns = list_policy_arns()
    policies = {arn: get_policy(arn) for arn in policy_arns}

    if len(sys.argv) >= 2:
        out = gzip.open(sys.argv[1], 'wt')
    else:
        out = sys.stdout

    json.dump(policies, out)
    out.close()

if __name__ == '__main__':
    main()
