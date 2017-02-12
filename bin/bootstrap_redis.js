"use strict";

const fs = require('fs');
const redis = require('redis');

const bus_stops = require('../bus_stops.json');

const bus_stop_nos = [];
const bus_stop_descs = [];

for (const bus_stop of bus_stops) {
  bus_stop_nos.push(0, `${bus_stop.BusStopCode}|${bus_stop.Description}`);
  bus_stop_descs.push(0, `${bus_stop.Description.toUpperCase()}|${bus_stop.BusStopCode}|${bus_stop.Description}`);
}

const client = redis.createClient();

client.zadd('bus_stop_nos', bus_stop_nos, (err, res) => {
  if (err) {
    console.error(err);
  } else {
    console.log(res);
  }
});

client.zadd('bus_stop_descs', bus_stop_descs, (err, res) => {
  if (err) {
    console.error(err);
  } else {
    console.log(res);
  }
});
