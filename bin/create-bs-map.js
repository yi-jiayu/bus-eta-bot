"use strict";

const fs = require('fs');

const bus_stops = require('../bus_stops.json');

let bss = {};

for (const bs of bus_stops) {
  bss[bs.BusStopCode] = {
    r: bs.RoadName,
    d: bs.Description,
    lat: bs.Latitude,
    lon: bs.Longitude
  };
}

fs.appendFileSync('../bus-stop-info.json', JSON.stringify(bss));
