"use strict";

const fs = require('fs');

const bus_stops = require('../bus_stops.json');

const bss = [];

for (const bs of bus_stops) {
  bss.push({i: bs.BusStopCode, d: bs.Description});
}

fs.appendFileSync('../bus-stops-array.json', JSON.stringify(bss));
