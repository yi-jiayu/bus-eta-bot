#!/usr/bin/env python3

import json
import sqlite3
from collections import OrderedDict

with open('bus_routes.json') as f:
    # in Python 3.7 and above, dictionaries are ordered by default
    # but this will explicitly preserve the key order
    bus_routes = json.load(f, object_pairs_hook=OrderedDict)

conn = sqlite3.connect('datamall.sqlite')
c = conn.cursor()
for bus_route in bus_routes:
    c.execute('insert into bus_routes values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)', tuple(bus_route.values()))
conn.commit()
print('Inserted', len(bus_routes), 'rows into bus_routes')
conn.close()
