#!/usr/bin/env python3

import logging
import os
import re
import sys

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
    return [
        rel_link(URLs.SAR_INDEX, a.get('href'))
        for a in sar_index.find_all('a')
        if link_re.match(a.get('href'))
    ]

def parse_sar_data(sar_page: BeautifulSoup) -> dict:
    """Iterate SAR pages and parse table contents."""
    return dict(
        service=subparse_service(sar_page),
        actions=subparse_actions(sar_page),
    )

def subparse_service(sar_page: BeautifulSoup) -> str:
    """Extract the service name from the provided sar page."""
    match = re.search(r'service prefix: ([a-z0-9]+)', sar_page.text)
    if not match:
        raise ValueError(f'unable to parse service from text:\n{sar_page.text}')

    return match.group(1)

def subparse_actions(sar_page: BeautifulSoup) -> list[dict]:
    """Extract the action details from the provided sar page."""

def main():
    sar_index = fetch_page(URLs.SAR_INDEX)
    sar_links = extract_sar_links(sar_index)
    sar_pages = [fetch_page(link) for link in sar_links[:1]] 
    sar_data = [parse_sar_data(page) for page in sar_pages[:1]]
    print(sar_data)

if __name__ == '__main__':
    main()
