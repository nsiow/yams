#!/usr/bin/env python3

import gzip
import json
import logging
import os
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

class Action(pydantic.BaseModel):
    Name: str
    ActionConditionKeys: list[str]
    Resources: list[ResourcePointer]

class Condition(pydantic.BaseModel):
    Name: str
    Types: list[str]

class Resource(pydantic.BaseModel):
    Name: str
    ARNFormats: list[str]
    ConditionKeys: list[str]

class Service(pydantic.BaseModel):
    Name: str
    Version: str
    Actions: list[Action]
    ConditionKeys: list[Condition]
    Resources: list[Resource]

# ------------------------------------------------------------------------------------------------
# SETUP
# ------------------------------------------------------------------------------------------------

# Set up logging
logging.basicConfig(level=os.environ.get('YAMS_LOG_LEVEL', 'INFO').upper(),
                    stream=sys.stdout)

# Set up cache
os.makedirs('.cache', exist_ok=True)
memory = joblib.Memory('.cache/sar_v2.cache')

# ------------------------------------------------------------------------------------------------
# HELPERS
# ------------------------------------------------------------------------------------------------

def fetch_service_listing() -> list[ServiceListing]:
    resp = requests.get(SERVICE_REFERENCE_URL)
    resp.raise_for_status()
    return [ServiceListing(**s) for s in resp.json()]

def fetch_service(service_listing: ServiceListing) -> Service:
    resp = requests.get(service_listing.url)
    resp.raise_for_status()
    return Service(**resp.json())

def restructure(services: list[Service]) -> dict:
    # TODO(nsiow) 
    return {}

# ------------------------------------------------------------------------------------------------
# MAIN
# ------------------------------------------------------------------------------------------------

def main():
    service_listing = fetch_service_listing()
    services = [fetch_service(s) for s in service_listing]
    sar_data = restructure(services)

    if len(sys.argv) >= 2:
        out = gzip.open(sys.argv[1], 'wt')
    else:
        out = sys.stdout

    json.dump(sar_data, out)
    out.close()

if __name__ == '__main__':
    main()
