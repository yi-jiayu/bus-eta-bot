"use strict";

import { assert } from 'chai';
import { get_etas } from '../src/lib/datamall';
import Datamall from '../src/bot/Datamall';

suite('datamall', function () {
  test('get etas for existing bus stop', function () {
    return get_etas('96049')
      .then(console.log);
  });

  test('get etas for nonexistent bus stop', function () {
    return get_etas('jiayu')
      .then(parsed_etas => assert.equal(parsed_etas.etas.length, 0));
  });

  test('test', function () {
    const dm = new Datamall();

    return get_etas('96049')
      .then(res => {
        console.log(JSON.stringify(res));
        return res;
      })
      .then(Datamall.parse_etas)
      .then(res => console.log(JSON.stringify(res)));
  });
});
