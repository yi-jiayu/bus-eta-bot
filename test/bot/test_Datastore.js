"use strict";

import { assert } from 'chai';

import Datastore from "../../src/bot/Datastore";

const datastore = new Datastore();

suite('Datastore', function () {
  suite('get bus stop info', function () {
    test('should return a promise', function () {
      assert.instanceOf(datastore.get_bus_stop_info(), Promise);
    });
  });

  suite('get completions', function () {
    test('should return a promise', function () {
      assert.instanceOf(datastore.get_completions(), Promise);
    });
  });

  suite('get nearby bus stops', function () {
    test('should return a promise', function () {
      assert.instanceOf(datastore.get_nearby_bus_stops(), Promise);
    });
  });
});
