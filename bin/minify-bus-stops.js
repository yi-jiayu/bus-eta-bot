"use strict";

const fs = require('fs');

const bus_stops = require('./bus-stops-sorted.json');
const bss = [];

for (const bs of bus_stops) {
  bss.push({
    i: bs.BusStopCode,
    g: bs.GeoHash
  });
}

fs.appendFileSync('bus-stop-locations.json', JSON.stringify(bss));
