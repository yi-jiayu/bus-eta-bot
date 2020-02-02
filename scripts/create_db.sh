#/bin/sh -

set -e

if [ -z "$DATAMALL_ACCOUNT_KEY" ]; then
	echo '$DATAMALL_ACCOUNT_KEY not set' >&2
	exit 1
fi

./scrape_bus_stops.py
./scrape_bus_routes.py

if [ -f datamall.sqlite ]; then
	echo datamall.sqlite already exists >&2
	exit 1
fi

sqlite3 datamall.sqlite < schema.sql
echo 'Created sqlite database datamall.sqlite'

./load_bus_stops.py
./load_bus_routes.py
