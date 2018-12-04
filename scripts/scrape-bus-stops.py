#!/usr/bin/env python3

import json
import os
import sys
import urllib.request

endpoint = 'http://datamall2.mytransport.sg/ltaodataservice/BusStops?$skip='


def get_bus_stops(account_key: str, offset: int):
    url = '{}{}'.format(endpoint, offset)
    req = urllib.request.Request(url, headers={'AccountKey': account_key})
    with urllib.request.urlopen(req) as f:
        bus_stops = json.load(f)
        return bus_stops


def main():

    account_key = os.environ.get('DATAMALL_ACCOUNT_KEY')
    if account_key is None:
        print('Error: DATAMALL_ACCOUNT_KEY environment variable not set.')
        sys.exit(1)

    if len(sys.argv) > 1:
        out_file = sys.argv[1]
    else:
        out_file = './bus_stops.json'

    print('[START] Fetch bus stops')
    with open(out_file, 'w') as f:
        offset = 0
        count = 0
        while True:
            bus_stops = get_bus_stops(account_key, offset)
            if len(bus_stops['value']) == 0:
                break
            count += len(bus_stops['value'])

            f.write('{:04d}: {}\n'.format(offset, json.dumps(bus_stops['value'])))
            sys.stdout.write('\rOffset: {}'.format(offset))
            offset += len(bus_stops['value'])
    print('\rFetched {} bus stops.'.format(count))
    print('[END]   Fetch bus stops')

    print('[START] Collect bus stops')
    all_bus_stops = []
    with open(out_file) as f:
        for line in f:
            bus_stops = json.loads(line[6:])
            all_bus_stops.extend(bus_stops)
    print('[END]   Collect bus stops')

    print('[START] Write bus stops')
    with open(out_file, 'w') as f:
        json.dump(all_bus_stops, f)
    print('[END]   Write bus stops')


if __name__ == "__main__":
    main()
