"use strict";

import Datastore from '../src/lib/Datastore';

const datastore = new Datastore(require('./bus-stop-info.json'), require('./location-data.json'));


suite('datastore', function () {
  suite('info', function () {
    test('info', function () {
      console.log(datastore.get_bus_stop_info('96049'));
    });
  });

  suite('autocomplete', function () {
    test('autocomplete', function () {
      console.log(datastore.get_completions('Diraja'));
    });
  });

  suite('nearby', function () {
    test('nearby', function () {
      console.log(datastore.get_nearby_bus_stops(1.43669132626001, 103.83368337352286));
    });
  });
});
