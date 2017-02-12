"use strict";

import { assert } from 'chai';
import { get_etas } from '../src/lib/datamall';

suite('datamall', function () {
  test('get etas for existing bus stop', function () {
    return get_etas('96049')
      .then(console.log);
  });

  test('get etas for nonexistent bus stop', function () {
    return get_etas('jiayu')
      .then(parsed_etas => assert.equal(parsed_etas.etas.length, 0));
  });
});
