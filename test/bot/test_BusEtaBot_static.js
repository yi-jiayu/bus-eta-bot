"use strict";

import { assert } from 'chai';

import BusEtaBot from '../../src/bot/BusEtaBot';

suite('BusEtaBot static functions', function () {
  suite('format eta message', function () {
    test('with bus stop info', function () {
      const msg = BusEtaBot.format_eta_message({
        updated: '2017-02-09T05:14:44.374Z',
        bus_stop_id: '96049',
        etas: [{svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22}]
      }, {info: {description: 'Opp Tropicana Condo', road: 'Upp Changi Rd East'}});

      assert.deepEqual(msg.serialise_send(0), {
        chat_id: 0,
        text: '*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n+-----+------+-----+-----+\n| Svc | Next | 2nd | 3rd |\n|-----|------|-----|-----|\n| 2   | 1    | 15  | 18  |\n|-----|------|-----|-----|\n| 24  | -2   | 7   | 22  |\n+-----+------+-----+-----+\n```\n_Last updated: Feb 9, 2017 1:14 PM._',
        parse_mode: 'markdown',
        reply_markup: '{"inline_keyboard":[[{"text":"Refresh","callback_data":"{\\"t\\":\\"refresh\\",\\"b\\":\\"96049\\",\\"s\\":[]}"}]]}',
        method: 'sendMessage'
      });
    });

    test('without bus stop info', function () {
      const msg = BusEtaBot.format_eta_message({
        updated: '2017-02-09T05:14:44.374Z',
        bus_stop_id: '96049',
        etas: [{svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22}]
      });


      assert.deepEqual(msg.serialise_send(0), {
        chat_id: 0,
        text: '*96049*\n```\n+-----+------+-----+-----+\n| Svc | Next | 2nd | 3rd |\n|-----|------|-----|-----|\n| 2   | 1    | 15  | 18  |\n|-----|------|-----|-----|\n| 24  | -2   | 7   | 22  |\n+-----+------+-----+-----+\n```\n_Last updated: Feb 9, 2017 1:14 PM._',
        parse_mode: 'markdown',
        reply_markup: '{"inline_keyboard":[[{"text":"Refresh","callback_data":"{\\"t\\":\\"refresh\\",\\"b\\":\\"96049\\",\\"s\\":[]}"}]]}',
        method: 'sendMessage'
      });
    });

    test('filtering on services', function () {
      const msg = BusEtaBot.format_eta_message({
        updated: '2017-02-09T05:14:44.374Z',
        bus_stop_id: '96049',
        etas: [{svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22}]
      }, {
        services: ['2']
      });

      assert.deepEqual(msg.serialise_send(0), {
        chat_id: 0,
        text: '*96049*\n```\n+-----+------+-----+-----+\n| Svc | Next | 2nd | 3rd |\n|-----|------|-----|-----|\n| 2   | 1    | 15  | 18  |\n+-----+------+-----+-----+\n1 more service not shown.```\n_Last updated: Feb 9, 2017 1:14 PM._',
        parse_mode: 'markdown',
        reply_markup: '{"inline_keyboard":[[{"text":"Refresh","callback_data":"{\\"t\\":\\"refresh\\",\\"b\\":\\"96049\\",\\"s\\":[\\"2\\"]}"}]]}',
        method: 'sendMessage'
      });
    });

    test('compact table when showing many services', function () {
      const msg = BusEtaBot.format_eta_message({
        updated: '2017-02-09T05:14:44.374Z',
        bus_stop_id: '96049',
        etas: [{svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22},
          {svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22},
          {svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22},
          {svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22},
          {svc_no: '2', next: 1, subsequent: 15, third: 18},
          {svc_no: '24', next: -2, subsequent: 7, third: 22}]
      }, {info: {description: 'Opp Tropicana Condo', road: 'Upp Changi Rd East'}});

      assert.deepEqual(msg.serialise_send(0), {
        chat_id: 0,
        text: '*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\nSvc  Next  2nd  3rd  \n2    1     15   18   \n24   -2    7    22   \n2    1     15   18   \n24   -2    7    22   \n2    1     15   18   \n24   -2    7    22   \n2    1     15   18   \n24   -2    7    22   \n2    1     15   18   \n24   -2    7    22   \n```\n_Last updated: Feb 9, 2017 1:14 PM._',
        parse_mode: 'markdown',
        reply_markup: '{"inline_keyboard":[[{"text":"Refresh","callback_data":"{\\"t\\":\\"refresh\\",\\"b\\":\\"96049\\",\\"s\\":[]}"}]]}',
        method: 'sendMessage'
      });
    });
  });

  suite('infer bus stop and service nos', function () {
    test('without services', function () {
      const actual = BusEtaBot.infer_bus_stop_and_service_nos('96049');

      const expected = {
        bus_stop_id: '96049',
        service_nos: []
      };

      assert.deepEqual(actual, expected);
    });

    test('with services', function () {
      const actual = BusEtaBot.infer_bus_stop_and_service_nos('96049 2 24');

      const expected = {
        bus_stop_id: '96049',
        service_nos: ['2', '24']
      };

      assert.deepEqual(actual, expected);
    });

    test('with multiple services', function () {
      const actual = BusEtaBot.infer_bus_stop_and_service_nos('96049 2 24 5');

      const expected = {
        bus_stop_id: '96049',
        service_nos: ['2', '24', '5']
      };

      assert.deepEqual(actual, expected);
    });
  });

  suite('prepare inline query answer', function () {
    test('inline query without results', function () {
      const completions = [];

      assert.isNull(BusEtaBot.prepare_inline_query_answer(completions));
    });

    test('one result', function () {
      const completions = [
        {
          id: '96049',
          description: 'Opp Tropicana Condo',
          road: 'Upp Changi Rd East',
          latitude: 1.2,
          longitude: 100.2
        }
      ];

      const expected = {
        "results": [{
          "type": "location",
          "id": "96049",
          "title": "Opp Tropicana Condo (96049)",
          "latitude": 1.2,
          "longitude": 100.2,
          "reply_markup": {
            "inline_keyboard": [[{
              "text": "Refresh",
              "callback_data": "{\"t\":\"eta\",\"b\":\"96049\"}"
            }]]
          },
          "input_message_content": {
            "message_text": "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n\n`Fetching etas...`",
            "parse_mode": "markdown"
          }
        }], "params": {"cache_time": 86400, "next_offset": ""}
      };

      const actual = BusEtaBot.prepare_inline_query_answer(completions);


      assert.deepEqual(actual, expected);
    });

    test('more than one result', function () {
      const completions = [
        {
          id: '96049',
          description: 'Opp Tropicana Condo',
          road: 'Upp Changi Rd East',
          latitude: 1.2,
          longitude: 100.2
        },
        {
          id: '96041',
          description: 'Bef Tropicana Condo',
          road: 'Upp Changi Rd East',
          latitude: 1.33995375346513,
          longitude: 103.96079768187379
        }
      ];

      const expected = {
        "results": [{
          "type": "location",
          "id": "96049",
          "title": "Opp Tropicana Condo (96049)",
          "latitude": 1.2,
          "longitude": 100.2,
          "reply_markup": {
            "inline_keyboard": [[{
              "text": "Refresh",
              "callback_data": "{\"t\":\"eta\",\"b\":\"96049\"}"
            }]]
          },
          "input_message_content": {
            "message_text": "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n\n`Fetching etas...`",
            "parse_mode": "markdown"
          }
        }, {
          "type": "location",
          "id": "96041",
          "title": "Bef Tropicana Condo (96041)",
          "latitude": 1.33995375346513,
          "longitude": 103.96079768187379,
          "reply_markup": {
            "inline_keyboard": [[{
              "text": "Refresh",
              "callback_data": "{\"t\":\"eta\",\"b\":\"96041\"}"
            }]]
          },
          "input_message_content": {
            "message_text": "*Bef Tropicana Condo (96041)*\nUpp Changi Rd East\n\n`Fetching etas...`",
            "parse_mode": "markdown"
          }
        }], "params": {"cache_time": 86400, "next_offset": ""}
      };

      const actual = BusEtaBot.prepare_inline_query_answer(completions);

      assert.deepEqual(actual, expected);
    });
  });

  suite('prepare welcome message', function () {
    test('', function () {
      const reply = BusEtaBot.prepare_welcome_message('Jiayu');

      const expected = {
        "text": "*Hello Jiayu,*\n\nBus Eta Bot is a Telegram bot which can tell you how long you have to wait for your bus to arrive. \n\nTo get started, try sending me a bus stop code to get etas for, such as `96049`.\n\nAlternatively, you can also search for bus stops by sending me an inline query. To try this out, use the button below, or type my username (@BusEtaBot) followed by a bus stop code, description or road name in any chat.\n\nThanks for trying out Bus Eta Bot! If you find Bus Eta Bot useful, do help to spread the word or use /feedback and leave some feedback to help make Bus Eta Bot even better!\n\nIf you're stuck, you can get help with using Bus Eta Bot by sending /help.\n\nRegards,  \n*J*  \nBus Eta Bot Developer",
        "params": {
          "parse_mode": "markdown",
          "reply_markup": "{\"inline_keyboard\":[[{\"text\":\"Get etas for bus stop 96049\",\"callback_data\":\"{\\\"t\\\":\\\"eta_demo\\\"}\"}],[{\"text\":\"Try an inline query\",\"switch_inline_query_current_chat\":\"Changi\"}]]}"
        }
      };

      assert.deepEqual(reply, expected);
    });
  });

  suite('prepare help message', function () {
    test('', function () {
      const reply = BusEtaBot.prepare_help_message();

      const expected = {
        "text": "Here's some help on how to use Bus Eta Bot:\n\nhttp://telegra.ph/Bus-Eta-Bot-Help-02-23",
        "params": {}
      };

      assert.deepEqual(reply, expected);
    });
  });
});
