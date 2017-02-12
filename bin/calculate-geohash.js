"use strict";

const Geo = require('geo-nearby');

// const fs = require('fs');
// const ngeohash = require('ngeohash');
//
// const bus_stops = require('../bus_stops.json');
//
// for (const bs of bus_stops) {
//   bs.GeoHash = ngeohash.encode_int(bs.Latitude, bs.Longitude);
// }
//
// fs.appendFile('../bus_stops_with_geohash.json', JSON.stringify(bus_stops));

const data = require('../bus_stops.json');

const dataSet = Geo.createCompactSet(data, {
  id: 'BusStopCode',
  lat: 'Latitude',
  lon: 'Longitude',
  file: '../location-data.json'
});
