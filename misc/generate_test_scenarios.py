#!/usr/bin/env python3

import json

import itertools
from typing import Any

# --------------------------------------------------------------------------------------------------
# Constants
# --------------------------------------------------------------------------------------------------

MAX_NUM_RESOURCES_PER_TEMPLATE = 500

# --------------------------------------------------------------------------------------------------
# Combinations
# --------------------------------------------------------------------------------------------------

PRINCIPAL_TYPE = [
    'user',
    'role',
]

SAME_OR_CROSS_ACCOUNT = [
    'same_account',
    'x_account',
]

PRINCIPAL_POLICY_TYPE = [
    'inline',
    'managed',
    'group',  # IAM users only
]

AWS_SERVICE = [
    'dynamodb',
    'iam',  # assume-role
    'kms',
    's3',
    'sns',
    'sqs',
]

INCL_RESOURCE_POLICY = [
    'with_resource_policy',
    'sans_resource_policy',
]

DECISION = [
    'should_allow',
    'should_NOT_allow',
]

FEATURE = [
    'basic',
    'principal_tags',
    'resource_tags',
    'permission_boundary',
]

# --------------------------------------------------------------------------------------------------
# Types
# --------------------------------------------------------------------------------------------------

#
# This is really the CF resource, almost CDK-esque, has type, properties
#

class Resource():

    account: int
    logical_name: str
    type: str
    properties: dict[str, Any]

    def __init__(self):
        self.properties = {}

    def as_json(self):
        return json.dumps({
            'Type': self.type,
            'Properties': self.properties,
        })

class TestCase():
    name: str
    policy_type: str
    principal_type: str
    is_x_account: bool
    resource_service: str
    has_resource_policy: bool
    other_feature: str
    should_allow: bool

    def is_valid(self) -> bool:
        return True

    # def gen_principal(self) -> list[Resource]:
    #     pass
    #
    # def gen_resource(self) -> list[Resource]:
    #     pass

    # def gen(self) -> list[Resource]:
    #     return (
    #         self.gen_principal()
    #         + self.gen_resource()
    #     )

#
# Set of templates, we want as few templates per account as possible, so binpack resources into
# templates of 500 resources each until we run out of test cases
#

class Template():

    _resources: list[Resource]

    def __init__(self):
        self._resources = []

# --------------------------------------------------------------------------------------------------
# Calculate product
# --------------------------------------------------------------------------------------------------

def product():
    return itertools.product(
        PRINCIPAL_TYPE,
        SAME_OR_CROSS_ACCOUNT,
        PRINCIPAL_POLICY_TYPE,
        AWS_SERVICE,
        FEATURE,
        DECISION,
    )

def gen_test_cases() -> list[TestCase]:
    test_cases = []
    
    for principal_type in PRINCIPAL_TYPE:
        for same_or_cross_account in SAME_OR_CROSS_ACCOUNT:
            for policy_type in PRINCIPAL_POLICY_TYPE:
                for aws_service in AWS_SERVICE:
                    for feature in FEATURE:
                        for resource_policy in INCL_RESOURCE_POLICY:
                            for decision in DECISION:
                                tc = TestCase()
                                tc.name = ' '.join([
                                    principal_type,
                                    same_or_cross_account,
                                    policy_type,
                                    aws_service,
                                    feature,
                                    resource_policy,
                                    decision,
                                ])
                                tc.principal_type = principal_type
                                tc.is_x_account = same_or_cross_account == 'x_account'
                                tc.policy_type = policy_type
                                tc.resource_service = aws_service
                                tc.other_feature = feature
                                tc.has_resource_policy = resource_policy == 'with_resource_policy'
                                tc.should_allow = decision == 'should_allow'
                                test_cases.append(tc)
    return test_cases

def main():
    test_cases = gen_test_cases()
    print([t.name for t in test_cases])

if __name__ == '__main__':
    main()
