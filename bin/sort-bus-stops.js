"use strict";

const fs = require('fs');

let bus_stops = require('./bus-stops.json');

bus_stops = bus_stops.sort((a, b) => {
  if (a.g === b.g) return 0;
  else if (a.g < b.g) return -1;
  else return 1;
});

fs.appendFileSync('../bus-stops-sorted.json', JSON.stringify(bus_stops));
