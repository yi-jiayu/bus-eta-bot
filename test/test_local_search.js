"use strict";

import Geo from 'geo-nearby';

let bus_stops;

let bs_map;
let bs_array;
let bs_locs;

suite('local search', function () {
  suite('load', function () {
    test('load map', function () {
      bs_map = require('./bus-stop-info.json');
    });

    test('load array', function () {
      bs_array = require('./bus-stops-array.json');
    });

    test('load locations', function () {
      bs_locs = require('./location-data.json');
    });
  });

  suite('search', function () {
    test('search map', function () {
      for (const key in bs_map) {
        //noinspection JSUnfilteredForInLoop
        const local = bs_map[key].d.indexOf('East Coast Pk Svc Rd');
      }
    });

    test('search array and lookup map', function () {
      for (const bs of bs_array) {
        const local = bs.d.indexOf('East Coast Pk Svc Rd')
      }
    });
  });

  suite('nearby', function () {
    bs_locs = require('./location-data.json');

    test('search nearby bsearch', function () {
      // set arbitrary limit of 5 nearest bus stops
      const geo = new Geo(bs_locs || require('./location-data.json'), {sorted: true, limit: 5});

      // try arbitrary 100m radius for now
      const nearby = geo.nearBy(1.43755803458153, 103.83159143616113, 100);

      console.log(nearby);
    });

    test('search nearby no bsearch', function () {
      // set arbitrary limit of 5 nearest bus stops
      const geo = new Geo(bs_locs || require('./location-data.json'), {limit: 5});

      // try arbitrary 100m radius for now
      const nearby = geo.nearBy(1.43755803458153, 103.83159143616113, 100);

      console.log(nearby);
    })
  });
});
