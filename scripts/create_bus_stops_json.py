#!/usr/bin/env python3

import json
import sqlite3
import sys

bus_stops_map = {}
conn = sqlite3.connect('datamall.sqlite')
conn.row_factory = sqlite3.Row
c = conn.cursor()
for row in c.execute('''select bus_stop_code code, road_name, description, latitude, longitude from bus_stops'''):
    stop = dict(row)
    bus_stops_map[stop['code']] = stop

conn.row_factory = None
c = conn.cursor()
for key in bus_stops_map.keys():
    c.execute('''select distinct service_no
from bus_routes
where bus_stop_code = ?''', (key,))
    rows = c.fetchall()
    services = [r[0] for r in rows]
    bus_stops_map[key]['services'] = services

bus_stops = sorted(bus_stops_map.values(), key=lambda s: s['code'])
json.dump(bus_stops, sys.stdout)
