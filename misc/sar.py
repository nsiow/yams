#!/usr/bin/env python3

import copy
import json
import logging
import os
import re
import sys
from typing import Union

from bs4 import BeautifulSoup
import joblib
import requests_html as requests


# Set up logging
logging.basicConfig(level=os.environ.get('YAMS_LOG_LEVEL', 'INFO').upper(),
                    stream=sys.stdout)

# Set up cache
memory = joblib.Memory('/tmp/sar.cache')

# Set up a browser session
sess = requests.HTMLSession()

class URLs:
    """Data class to hold URLs we need to use elsewhere."""
    SAR_INDEX = 'https://docs.aws.amazon.com/service-authorization/latest/reference/reference.html'

@memory.cache
def fetch_page(url: str) -> BeautifulSoup:
    """Helper function that fetches the page and parses it as a bs4 object."""
    logging.info('Fetching page: %s', url)
    resp = sess.get(url)
    resp.raise_for_status()
    html = resp.html  # type: ignore
    html.render()
    return BeautifulSoup(html.html, 'html.parser')

def rel_link(url: str, relative: str) -> str:
    """Given the source URL and relative fragment, generate a new URL."""
    cur_dir = os.path.dirname(url)
    return os.path.join(cur_dir, relative)

def extract_sar_links(sar_index: BeautifulSoup) -> list[str]:
    """Extract all SAR page links from the provided index."""
    link_re = re.compile(r'^list_\S+.html$')
    results = [
        rel_link(URLs.SAR_INDEX, a.get('href'))
        for a in sar_index.find_all('a')
        if link_re.match(a.get('href'))
    ]

    skip = [
        'list_awsiot1',
        'list_amazonapigatewaymanagementv2',
    ]

    # remove some invalid SAR pages
    results = [r for r in results if not any(s in r for s in skip)]
    return results

def parse_sar_data(sar_page: BeautifulSoup) -> dict:
    """Iterate SAR pages and parse table contents."""
    return dict(
        service=subparse_service(sar_page),
        actions=subparse_actions(sar_page),
        condition_keys=subparse_condition_keys(sar_page),
    )

def subparse_service(sar_page: BeautifulSoup) -> str:
    """Extract the service name from the provided sar page."""
    match = re.search(r'service prefix: ([a-z0-9-]+)', sar_page.text)
    if not match:
        raise ValueError(f'unable to parse service from text:\n{sar_page.text}')

    service = match.group(1)
    return service


def normalize_scalar(field: str) -> str:
    """Helper function to normalize scalar values found in SAR data tables."""
    field = field.replace('[permission only]', '')
    field = field.strip()
    return field

def normalize_list(field: Union[list, str]) -> list:
    """Helper function to normalize scalar values found in SAR data tables."""
    if isinstance(field, str):
        return field.strip().split()
    elif isinstance(field, list):
        return field

    # type: ignore
    raise TypeError('Not sure how to normalize value: {}'.format(repr(field)))  # type: ignore

def subparse_actions(sar_page: BeautifulSoup) -> list[dict]:
    """Extract the action details from the provided sar page."""
    table = sar_page.find_all(class_='table-container')[0]
    if not table:
        raise ValueError('unable to locate `table-container` on page')

    # Locate and process headers
    exp_headers = ['actions', 'description', 'access_level', 'resource_types', 'condition_keys', 'dependent_actions']
    headers = [re.sub(r'_\(.*\)', '', th.text.strip().lower().replace(' ', '_'))
               for th in table.find('tr').find_all('th')]
    if headers != exp_headers:
        raise ValueError(f'unexpected headers: {headers}')

    actions = []
    for row in table.find_all('tr')[1:]:
        columns = [col.get_text() for col in row.find_all('td')]

        # if we are missing columns, this is a "continuation" row
        if len(columns) < len(headers):
            prev_row_data = actions[0]
            row_headers = exp_headers[-len(columns):]
            row_data = prev_row_data | dict(zip(row_headers, columns))
        else:
            # otherwise it's a normal row
            row_data = dict(zip(exp_headers, columns))

        for a in ['action', 'actions']:
            if a in row_data:
                row_data[a] = normalize_scalar(row_data.pop(a))

        row_data['description'] = normalize_scalar(row_data['description'])
        row_data['access_level'] = normalize_scalar(row_data['access_level'])
        row_data['resource_types'] = normalize_list(row_data['resource_types'])
        row_data['condition_keys'] = normalize_list(row_data['condition_keys'])
        row_data['dependent_actions'] = normalize_list(row_data['dependent_actions'])
        actions.append(row_data)

    return actions

def subparse_condition_keys(sar_page: BeautifulSoup) -> list[str]:
    """Extract the condition key details from the provided sar page."""
    table = sar_page.find_all(class_='table-container')[-1]

    if not table:
        raise ValueError('unable to locate `table-container` on page')

    # Locate and process headers
    headers = [re.sub(r'_\(.*\)', '', th.text.strip().lower().replace(' ', '_'))
               for th in table.find('tr').find_all('th')]
    if headers != ['condition_keys', 'description', 'type']:
        return []

    condition_keys = []
    for row in table.find_all('tr')[1:]:
        row_data = {}
        columns = row.find_all('td')
        for i, col in enumerate(columns):
            row_data[headers[i]] = col.get_text(strip=True).strip()
        condition_keys.append(row_data)

    return condition_keys

def normalize_sar_data(sar_data: list[dict]) -> list[dict]:
    """Helper function to normalize and remove some inconsistencies in the SAR data."""
    # Combine pages for services under the same umbrella
    sar_dict = {}
    for service_data in sar_data:
        try:
            service = service_data['service']
            if service in sar_dict:
                sar_dict[service]['actions'].extend(service_data['actions'])
                sar_dict[service]['condition_keys'].extend(service_data['condition_keys'])
            else:
                sar_dict[service] = service_data
        except Exception:
            print(f'error merging pages for service: {service_data["service"]}', file=sys.stderr)
            raise

    return list(sar_dict.values())

def main():

    sar_index = fetch_page(URLs.SAR_INDEX)
    sar_links = extract_sar_links(sar_index)
    sar_pages = [fetch_page(link) for link in sar_links]
    sar_data = normalize_sar_data([parse_sar_data(page) for page in sar_pages])

    if len(sys.argv) >= 2:
        out = open(sys.argv[1], 'w+')
    else:
        out = sys.stdout

    json.dump(sar_data, out)

if __name__ == '__main__':
    main()
