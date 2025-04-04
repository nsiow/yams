#!/usr/bin/env python3

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

class Action(pydantic.BaseModel):
    Name: str
    Service: str | None = None
    ActionConditionKeys: list[str] = []
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
    service = propagate_service(service)
    service = resolve_resource_pointers(service)
    return service

# aws:RequestTag/${TagKey} => aws:requesttag
def normalize_condition_variables(service: Service) -> Service:
    for action in service.Actions:
        for i in range(len(action.ActionConditionKeys)):
            condition_key = re.sub(r'[/:]\${[a-z0-9]+}$', '', action.ActionConditionKeys[i])
            condition_key = condition_key.lower()
            action.ActionConditionKeys[i] = condition_key
    return service

# add Service.Name to all Service.Actions.Service
def propagate_service(service: Service) -> Service:
    for action in service.Actions:
        action.Service = service.Name
    return service

# resolve Service.Actions[].ResolvedResources
def resolve_resource_pointers(service: Service) -> Service:
    for action in service.Actions:
        for resource in action.Resources:
            try:
                resolved = next(r for r in service.Resources if r.Name == resource.Name)
                action.ResolvedResources.append(resolved)
            except StopIteration:
                raise ValueError('unable to resolve pointer for {}/{}/{}'.format(
                    service.Name, action.Name, resource.Name
                ))
    return service

# ------------------------------------------------------------------------------------------------
# MAIN
# ------------------------------------------------------------------------------------------------

def main():
    service_listing = fetch_service_listing()
    services = [normalize(fetch_service(s)) for s in service_listing]
    services_json = [s.model_dump(exclude_defaults=True) for s in services]

    if len(sys.argv) >= 2:
        out = gzip.open(sys.argv[1], 'wt')
    else:
        out = sys.stdout

    json.dump(services_json, out)
    out.close()

if __name__ == '__main__':
    main()
