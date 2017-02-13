"use strict";

import { assert } from 'chai';
import * as sinon from 'sinon';

import BusEtaBot from '../../src/bot/BusEtaBot';

suite('BusEtaBot', function () {
  suite('prepare eta message', function () {
    test('with bus stop info', function () {
      const msg = BusEtaBot.format_eta_message({
        updated: '2017-02-09T05:14:44.374Z',
        bus_stop: '96049',
        etas: [{svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22}]
      }, {info: {description: 'Opp Tropicana Condo', road: 'Upp Changi Rd East'}});

      assert.deepEqual(msg.serialise_send(0), {
        chat_id: 0,
        text: '*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n+-----+------+-----+-----+\n| Svc | Next | 2nd | 3rd |\n|-----|------|-----|-----|\n| 2   | 1    | 15  | 18  |\n|-----|------|-----|-----|\n| 24  | -2   | 7   | 22  |\n+-----+------+-----+-----+\n```\n_Last updated: Feb 9, 2017 1:14 PM._',
        parse_mode: 'markdown',
        reply_markup: '{"inline_keyboard":[[{"text":"Refresh","callback_data":"{\\"t\\":\\"eta\\",\\"b\\":\\"96049\\"}"}]]}',
        method: 'sendMessage'
      });
    });

    test('without bus stop info', function () {
      const msg = BusEtaBot.format_eta_message({
        updated: '2017-02-09T05:14:44.374Z',
        bus_stop: '96049',
        etas: [{svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22}]
      });

      assert.deepEqual(msg.serialise_send(0), {
        chat_id: 0,
        text: '*96049*\n```\n+-----+------+-----+-----+\n| Svc | Next | 2nd | 3rd |\n|-----|------|-----|-----|\n| 2   | 1    | 15  | 18  |\n|-----|------|-----|-----|\n| 24  | -2   | 7   | 22  |\n+-----+------+-----+-----+\n```\n_Last updated: Feb 9, 2017 1:14 PM._',
        parse_mode: 'markdown',
        reply_markup: '{"inline_keyboard":[[{"text":"Refresh","callback_data":"{\\"t\\":\\"eta\\",\\"b\\":\\"96049\\"}"}]]}',
        method: 'sendMessage'
      });
    });

    test('filtering on services', function () {
      const msg = BusEtaBot.format_eta_message({
        updated: '2017-02-09T05:14:44.374Z',
        bus_stop: '96049',
        etas: [{svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22}]
      }, {
        services: ['2']
      });

      assert.deepEqual(msg.serialise_send(0), {
        "chat_id": 0,
        "method": "sendMessage",
        "parse_mode": "markdown",
        "reply_markup": "{\"inline_keyboard\":[[{\"text\":\"Refresh\",\"callback_data\":\"{\\\"t\\\":\\\"eta\\\",\\\"b\\\":\\\"96049\\\",\\\"s\\\":[\\\"2\\\"]}\"}]]}",
        "text": "*96049*\n```\n+-----+------+-----+-----+\n| Svc | Next | 2nd | 3rd |\n|-----|------|-----|-----|\n| 2   | 1    | 15  | 18  |\n+-----+------+-----+-----+\n1 more service not shown.```\n_Last updated: Feb 9, 2017 1:14 PM._"
      });
    });
  });

  suite('infer bus stop and service nos', function () {
    test('without services', function () {
      const actual = BusEtaBot.infer_bus_stop_and_service_nos('96049');

      const expected = {
        bus_stop: '96049',
        service_nos: []
      };

      assert.deepEqual(actual, expected);
    });

    test('with services', function () {
      const actual = BusEtaBot.infer_bus_stop_and_service_nos('96049 2 24');

      const expected = {
        bus_stop: '96049',
        service_nos: ['2', '24']
      };

      assert.deepEqual(actual, expected);
    });

  });
});
