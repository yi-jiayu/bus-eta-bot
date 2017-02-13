"use strict";

import { assert } from 'chai';

import LocalDatastore from '../../src/bot/LocalDatastore';

const info = require('../../src/data/bus-stop-info.json');
const locations = require('../../src/data/location-data.json');

const datastore = new LocalDatastore(info, locations);

suite('LocalDatastore', function () {
  suite('get bus stop info', function () {
    test('existing bs stop id', function () {
      return datastore.get_bus_stop_info(96049)
        .then(bs => assert.equal(bs.id, '96049'));
    });

    test('get nonexistent bus stop id', function () {
      return datastore.get_bus_stop_info('yi-jiayu')
        .then(bs => assert.strictEqual(bs, null));
    });
  });

  suite('get completions', function () {
    test('match bus stop ids by prefix', function () {
      const query = '9604';
      return datastore.get_completions(query)
        .then(cs => {
          // this bus stop exists
          assert.isTrue(cs.length > 0);

          for (const c of cs) {
            assert.isTrue(c.id.startsWith(query));
          }
        });
    });

    test('match bus stops descriptions', function () {
      const query = 'casa ';
      return datastore.get_completions(query)
        .then(cs => {
          assert.isTrue(cs.length > 0, 'this query should return results');

          for (const c of cs) {
            assert.isTrue(c.description.toUpperCase().indexOf(query.toUpperCase().trim()) !== -1);
          }
        });
    });

    test('match bus stop roads', function () {
      const query = 'kurau rd';
      return datastore.get_completions(query)
        .then(cs => {
          assert.isTrue(cs.length > 0, 'this query should return results');

          for (const c of cs) {
            assert.isTrue(c.road.toUpperCase().indexOf(query.toUpperCase().trim()) !== -1);
          }
        })
    });
  });

  suite('get nearby bus stops', function () {
    test('reasonable location', function () {
      return datastore.get_nearby_bus_stops(1.320299, 103.904771)
        .then(nearby => {
          assert.isTrue(nearby.length > 0, 'there are bus stops near this location');
        });
    });
  });
});
