#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "joblib",
#     "pydantic",
#     "requests",
# ]
# ///

import gzip
import json
import logging
import os
import re
import sys

import joblib
import pydantic
import requests

# ------------------------------------------------------------------------------------------------
# CONSTANTS
# ------------------------------------------------------------------------------------------------

SERVICE_REFERENCE_URL = 'https://servicereference.us-east-1.amazonaws.com/'

# ------------------------------------------------------------------------------------------------
# MODELS
# ------------------------------------------------------------------------------------------------

class ServiceListing(pydantic.BaseModel):
    service: str
    url: str

class ResourcePointer(pydantic.BaseModel):
    Name: str

class Resource(pydantic.BaseModel):
    Name: str
    ARNFormats: list[str] = []
    ConditionKeys: list[str] = []
    CustomHandling: list[str] = []

class ActionProperties(pydantic.BaseModel):
    IsList: bool = False
    IsPermissionManagement: bool = False
    IsTaggingOnly: bool = False
    IsWrite: bool = False

class ActionAnnotations(pydantic.BaseModel):
    Properties: ActionProperties = ActionProperties()

class Action(pydantic.BaseModel):
    Name: str
    Service: str | None = None
    AccessLevel: str | None = None
    ActionConditionKeys: list[str] = []
    Annotations: ActionAnnotations = ActionAnnotations()
    Resources: list[ResourcePointer] = []
    ResolvedResources: list[Resource] = []

class Condition(pydantic.BaseModel):
    Name: str
    Types: list[str]

class Service(pydantic.BaseModel):
    Name: str
    Version: str
    Actions: list[Action]
    ConditionKeys: list[Condition] = []
    Resources: list[Resource] = []

# ------------------------------------------------------------------------------------------------
# SETUP
# ------------------------------------------------------------------------------------------------

# Set up logging
logging.basicConfig(level=os.environ.get('YAMS_LOG_LEVEL', 'INFO').upper(),
                    stream=sys.stdout)

# Set up cache
os.makedirs('.cache', exist_ok=True)
mem = joblib.Memory('.cache/sar_v2.cache')

# ------------------------------------------------------------------------------------------------
# HELPERS
# ------------------------------------------------------------------------------------------------

@mem.cache
def fetch_service_listing() -> list[ServiceListing]:
    resp = requests.get(SERVICE_REFERENCE_URL)
    resp.raise_for_status()
    return [ServiceListing(**s) for s in resp.json()]

@mem.cache
def fetch_service(service_listing: ServiceListing) -> Service:
    resp = requests.get(service_listing.url)
    resp.raise_for_status()
    return Service(**resp.json())

def normalize(service: Service) -> Service:
    service = normalize_condition_variables(service)
    service = normalize_resource_arn_formats(service)
    service = propagate_service(service)
    service = propagate_access_level(service)
    service = resolve_resource_pointers(service)
    service = apply_custom_handling(service)
    return service

# aws:RequestTag/${TagKey} => aws:requesttag
def normalize_condition_variables(service: Service) -> Service:
    condkey_regex = r'[/:]\${[a-zA-Z0-9]+}$'
    for action in service.Actions:
        for i in range(len(action.ActionConditionKeys)):
            condition_key = re.sub(condkey_regex, '', action.ActionConditionKeys[i])
            condition_key = condition_key.lower()
            action.ActionConditionKeys[i] = condition_key
    for resource in service.Resources:
        for i in range(len(resource.ConditionKeys)):
            condition_key = re.sub(condkey_regex, '', resource.ConditionKeys[i])
            condition_key = condition_key.lower()
            resource.ConditionKeys[i] = condition_key
    return service

# arn:${Partition}:dynamodb:${Region}:${Account}:table/${TableName} => arn:*:dynamodb:*:*:table/*
def normalize_resource_arn_formats(service: Service) -> Service:
    arn_format_regex = r'\${[a-zA-Z0-9]+?}'
    for resource in service.Resources:
        for i in range(len(resource.ARNFormats)):
            format = re.sub(arn_format_regex, '*', resource.ARNFormats[i])
            resource.ARNFormats[i] = format
    return service

# add Service.Name to all Service.Actions.Service
def propagate_service(service: Service) -> Service:
    for action in service.Actions:
        action.Service = service.Name
    return service

# derive AccessLevel from Annotations.Properties
def propagate_access_level(service: Service) -> Service:
    for action in service.Actions:
        props = action.Annotations.Properties
        if props.IsPermissionManagement:
            action.AccessLevel = 'Permissions management'
        elif props.IsTaggingOnly:
            action.AccessLevel = 'Tagging'
        elif props.IsList:
            action.AccessLevel = 'List'
        elif props.IsWrite:
            action.AccessLevel = 'Write'
        else:
            action.AccessLevel = 'Read'
    return service

# resolve Service.Actions[].ResolvedResources
def resolve_resource_pointers(service: Service) -> Service:
    for action in service.Actions:
        for resource in action.Resources:
            try:
                resolved = next(r for r in service.Resources if r.Name == resource.Name)
                action.ResolvedResources.append(resolved)
            except StopIteration:
                # Some actions reference resources that don't exist in the service
                # This is a data quality issue from AWS; skip these
                logging.warning('unable to resolve pointer for %s/%s/%s',
                    service.Name, action.Name, resource.Name)
    return service

# apply custom handling rules to resolved resources
def apply_custom_handling(service: Service) -> Service:
    for action in service.Actions:
        for resource in action.ResolvedResources:
            # S3 bucket resources with arn:*:s3:::* should disallow slashes
            # This prevents bucket-level actions from matching object ARNs
            if service.Name == 's3' and resource.Name == 'bucket':
                if 'arn:*:s3:::*' in resource.ARNFormats:
                    if 'DisallowSlashes' not in resource.CustomHandling:
                        resource.CustomHandling.append('DisallowSlashes')
    return service

# ------------------------------------------------------------------------------------------------
# MAIN
# ------------------------------------------------------------------------------------------------

def main():
    service_listing = fetch_service_listing()
    services = [normalize(fetch_service(s)) for s in service_listing]
    # Exclude Annotations since we've derived AccessLevel from it
    services_json = [s.model_dump(exclude_defaults=True, exclude={'Actions': {'__all__': {'Annotations'}}}) for s in services]

    if len(sys.argv) >= 2:
        out = gzip.open(sys.argv[1], 'wt')
    else:
        out = sys.stdout

    json.dump(services_json, out)
    out.close()

if __name__ == '__main__':
    main()
