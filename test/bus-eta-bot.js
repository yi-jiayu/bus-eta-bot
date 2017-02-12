"use strict";

import { assert } from 'chai';
import Bot from '../src/lib/Bot';
import BusEtaBot from '../src/bot/BusEtaBot';
import bus_eta_bot from '../src/bot/bus-eta-bot';

suite('bot', function () {
  suite('bot static functions', function () {
    test('correctly identify bot commands', function () {
      const update = {
        "update_id": 100000000, "message": {
          "message_id": 5,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1", "type": "private"},
          "date": 1476436973,
          "text": "/version",
          "entities": [{"type": "bot_command", "offset": 0, "length": 8}]
        }
      };

      const is_bot_command = Bot._is_command(update);

      assert.isTrue(is_bot_command);
    });

    test('no false positive for bot command', function () {
      const update = {
        "update_id": 100000000, "message": {
          "message_id": 5,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1", "type": "private"},
          "date": 1476436973,
          "text": "hello"
        }
      };

      const is_bot_command = Bot._is_command(update);

      assert.isFalse(is_bot_command);
    });

    test('entities without command', function () {
      const update = {
        "update_id": 100000000,
        "message": {
          "message_id": 25,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu", "type": "private"},
          "date": 1486870410,
          "text": "10099 (Opp Bt Merah Town Ctr)\nJln Bt Merah",
          "entities": [{"type": "bold", "offset": 0, "length": 29}]
        }
      };

      console.log(update.message.entities.find(e => e.type === 'bot_command'));

      console.log(Bot._is_command(update));
    });
  });

  suite('bot command handling', function () {
    test('unregistered command', function () {
      const bot = new Bot();

      const update = {
        "update_id": 100000000, "message": {
          "message_id": 5,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1", "type": "private"},
          "date": 1476436973,
          "text": "/version",
          "entities": [{"type": "bot_command", "offset": 0, "length": 8}]
        }
      };

      return bot.handle(update)
        .catch(err => assert.equal(err, 'err: no command handler registered for version'));
    });

    test('single command handler', function () {
      const bot = new Bot();

      bot.command('version', (b, msg) => {
        assert.strictEqual(b, bot);
      });

      const update = {
        "update_id": 100000000, "message": {
          "message_id": 5,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1", "type": "private"},
          "date": 1476436973,
          "text": "/version",
          "entities": [{"type": "bot_command", "offset": 0, "length": 8}]
        }
      };

      return bot.handle(update);
    });

    test('more than one command handler', function () {
      const bot = new Bot();

      bot.command('version',
        (b, msg) => {
          msg.hello = 'bye';
        },
        (b, msg) => {
          assert.equal(msg.hello, 'bye');
        }
      );

      const update = {
        "update_id": 100000000, "message": {
          "message_id": 5,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1", "type": "private"},
          "date": 1476436973,
          "text": "/version",
          "entities": [{"type": "bot_command", "offset": 0, "length": 8}]
        }
      };

      return bot.handle(update);
    });

    test('not a command', function () {
      const bot = new Bot();

      return bot.handle({
        "update_id": 100000000,
        "message": {
          "message_id": 25,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu", "type": "private"},
          "date": 1486870410,
          "text": "10099 (Opp Bt Merah Town Ctr)\nJln Bt Merah",
          "entities": [{"type": "bold", "offset": 0, "length": 29}]
        }
      });
    });
  });

  suite('bot message handling', function () {
    test('handle text message', function () {
      const bot = new Bot();

      bot.message('text', (bot, text_msg) => 'received a text message');

      const update = {
        "update_id": 100000000, "message": {
          "message_id": 5,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1", "type": "private"},
          "date": 1476436973,
          "text": "version"
        }
      };

      return bot.handle(update)
        .then(val => assert.equal(val, 'received a text message'));
    });
  });

  suite('bot callback query handling', function () {
    test('handle callback query', function () {
      const bot = new Bot();

      bot.callback_query('eta', (bot, cbq) => {
        return 'ok'
      });

      const callback_query = {
        "update_id": 100000000, "callback_query": {
          "id": "100000000000000000",
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1"},
          "message": {
            "message_id": 6,
            "from": {"id": 263767667, "first_name": "CalculusGame", "username": "CalculusGameBot"},
            "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu1", "type": "private"},
            "date": 1476437912,
            "text": "Etas for bus stop 96049\nSvc    Inc. Buses\n2         3    14    22\n24       15    27    35\n5        12    23    30",
            "entities": [{"type": "pre", "offset": 24, "length": 89}]
          },
          "chat_instance": "-755686065470890988",
          "data": "{\"t\": \"eta\", \"d\":false, \"a\": \"81111 155\"}"
        }
      };

      return bot.handle(callback_query)
        .then(val => assert.equal(val, 'ok'));
    });
  });

  suite('bot inline query handling', function () {
    test('inline query', function () {
      return bus_eta_bot.handle({
        "update_id": 493995276,
        "inline_query": {
          "id": "100000000000000000",
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu"},
          "location": {"latitude": 1.342488, "longitude": 103.842018},
          "query": "opp",
          "offset": ""
        }
      })
        .then(console.log)
        .catch(console.error);
    });

    test('chosen inline result', function () {
      const cir_update = {
        "update_id": 100000000,
        "chosen_inline_result": {
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu"},
          "location": {"latitude": 1.309022, "longitude": 103.911827},
          "query": "",
          "result_id": "10009"
        }
      }
    });
  });

  suite('busetabot static functions', function () {
    test('bot prepare eta message', function () {
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
  });

});

