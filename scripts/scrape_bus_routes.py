#!/usr/bin/env python3

import json
import os
import sys
import urllib.request

endpoint = 'http://datamall2.mytransport.sg/ltaodataservice/BusRoutes?$skip='


def get_bus_routes(account_key: str, offset: int):
    url = '{}{}'.format(endpoint, offset)
    req = urllib.request.Request(url, headers={'AccountKey': account_key})
    with urllib.request.urlopen(req) as f:
        bus_routes = json.load(f)
        return bus_routes


def main():

    account_key = os.environ.get('DATAMALL_ACCOUNT_KEY')
    if account_key is None:
        print('Error: DATAMALL_ACCOUNT_KEY environment variable not set.')
        sys.exit(1)

    if len(sys.argv) > 1:
        out_file = sys.argv[1]
    else:
        out_file = './bus_routes.json'

    print('[START] Fetch bus routes')
    with open(out_file, 'w') as f:
        offset = 0
        count = 0
        while True:
            bus_routes = get_bus_routes(account_key, offset)
            if len(bus_routes['value']) == 0:
                break
            count += len(bus_routes['value'])

            f.write('{:04d}: {}\n'.format(offset, json.dumps(bus_routes['value'])))
            sys.stdout.write('\rOffset: {}'.format(offset))
            offset += len(bus_routes['value'])
    print('\rFetched {} bus routes.'.format(count))
    print('[END]   Fetch bus routes')

    print('[START] Collect bus routes')
    all_bus_routes = []
    with open(out_file) as f:
        for line in f:
            bus_routes = json.loads(line[6:])
            all_bus_routes.extend(bus_routes)
    print('[END]   Collect bus routes')

    print('[START] Write bus routes')
    with open(out_file, 'w') as f:
        json.dump(all_bus_routes, f)
    print('[END]   Write bus routes')


if __name__ == "__main__":
    main()
