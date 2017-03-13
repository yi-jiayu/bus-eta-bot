/**
 * Created by fn on 2/20/2017.
 */

import { assert } from 'chai';
import sinon from 'sinon';

import BusEtaBot from '../../src/bot/BusEtaBot';

suite('BusEtaBot', function () {
  suite('eta callback query handling', function () {
    test('old style eta callback', function () {
      const bot = new BusEtaBot();

      const answer_cbq_spy = sinon.spy();

      bot._callback_query_handlers['eta'].unshift((bot, cbq) => {
        cbq.answer = answer_cbq_spy;
      });

      const cbq = {
        "update_id": 100000000,
        "callback_query": {
          "id": "100000000000000000",
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "message": {
            "message_id": 85,
            "from": {"id": 100000000, "first_name": "Bus Eta Bot", "username": "BusEtaBot"},
            "chat": {"id": 100000000, "first_name": "fn", "username": "un", "type": "private"},
            "date": 1487566812,
            "text": "message_text",
            "entities": [{"type": "bold", "offset": 0, "length": 25}, {
              "type": "pre",
              "offset": 39,
              "length": 364
            }, {"type": "italic", "offset": 404, "length": 35}]
          },
          "chat_instance": "10000000000000000",
          "data": "{\"t\":\"eta\",\"a\":\"96049 2 24\"}"
        }
      };

      const update_message_spy = sinon.spy();

      bot.prepare_eta_message = function () {
        return Promise.resolve({
          update_message: update_message_spy
        });
      };

      return bot.handle(cbq)
        .then(() => {
          assert.isTrue(update_message_spy.calledOnce, 'update_message should have been called on the prepared message.');
          assert.isTrue(update_message_spy.calledWith(100000000, 85), 'update_message should have been called with the right message_id and chat_id');
          assert.isTrue(answer_cbq_spy.calledOnce, 'cbq.answer should have been called once');
        });

    });

    test('new style eta callback', function () {
      const bot = new BusEtaBot();

      const answer_cbq_spy = sinon.spy();

      bot._callback_query_handlers['eta'].unshift((bot, cbq) => {
        cbq.answer = answer_cbq_spy;
      });

      const cbq = {
        "update_id": 100000000,
        "callback_query": {
          "id": "100000000000000000",
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "message": {
            "message_id": 85,
            "from": {"id": 100000000, "first_name": "Bus Eta Bot", "username": "BusEtaBot"},
            "chat": {"id": 100000000, "first_name": "fn", "username": "un", "type": "private"},
            "date": 1487566812,
            "text": "message_text",
            "entities": [{"type": "bold", "offset": 0, "length": 25}, {
              "type": "pre",
              "offset": 39,
              "length": 364
            }, {"type": "italic", "offset": 404, "length": 35}]
          },
          "chat_instance": "10000000000000000",
          "data": "{\"t\":\"eta\",\"b\":\"96049\",\"s\":[\"2\"]}"
        }
      };

      const update_message_spy = sinon.spy();

      bot.prepare_eta_message = function () {
        return Promise.resolve({
          update_message: update_message_spy
        });
      };

      return bot.handle(cbq)
        .then(() => {
          assert.isTrue(update_message_spy.calledOnce, 'update_message should have been called on the prepared message.');
          assert.isTrue(update_message_spy.calledWith(100000000, 85), 'update_message should have been called with the right message_id and chat_id');
          assert.isTrue(answer_cbq_spy.calledOnce, 'cbq.answer should have been called once');
        });

    });

  });

  suite('refresh callback query handling', function () {
    test('from message', function () {
      const bot = new BusEtaBot();

      const answer_cbq_spy = sinon.spy();

      bot._callback_query_handlers['refresh'].unshift((bot, cbq) => {
        cbq.answer = answer_cbq_spy;
      });

      const cbq = {
        "update_id": 100000000,
        "callback_query": {
          "id": "100000000000000000",
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "message": {
            "message_id": 85,
            "from": {"id": 100000000, "first_name": "Bus Eta Bot", "username": "BusEtaBot"},
            "chat": {"id": 100000000, "first_name": "fn", "username": "un", "type": "private"},
            "date": 1487566812,
            "text": "message_text",
            "entities": [{"type": "bold", "offset": 0, "length": 25}, {
              "type": "pre",
              "offset": 39,
              "length": 364
            }, {"type": "italic", "offset": 404, "length": 35}]
          },
          "chat_instance": "10000000000000000",
          "data": "{\"t\":\"refresh\",\"b\":\"83151\"}"
        }
      };

      const update_message_spy = sinon.spy();

      bot.prepare_eta_message = function () {
        return Promise.resolve({
          update_message: update_message_spy
        });
      };

      return bot.handle(cbq)
        .then(() => {
          assert.isTrue(update_message_spy.calledOnce, 'update_message should have been called on the prepared message.');
          assert.isTrue(update_message_spy.calledWith(100000000, 85), 'update_message should have been called with the right message_id and chat_id');
          assert.isTrue(answer_cbq_spy.calledOnce, 'cbq.answer should have been called once');
        });
    });

    test('from inline query', function () {
      const bot = new BusEtaBot();

      const answer_cbq_spy = sinon.spy();

      bot._callback_query_handlers['refresh'].unshift((bot, cbq) => {
        cbq.answer = answer_cbq_spy;
      });

      const cbq = {
        "update_id": 130514761,
        "callback_query": {
          "id": "1000000000000000",
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "inline_message_id": "aeoudhtns",
          "chat_instance": "1000000000",
          "data": "{\"t\":\"refresh\",\"b\":\"62109\"}"
        }
      };

      const update_inline_message_spy = sinon.spy();

      bot.prepare_eta_message = function () {
        return Promise.resolve({
          update_inline_message: update_inline_message_spy
        });
      };

      return bot.handle(cbq)
        .then(() => {
          assert.isTrue(update_inline_message_spy.calledOnce, 'update_inline_message should have been called once');
          assert.isTrue(update_inline_message_spy.calledWith('aeoudhtns'), 'update_inline_message should have been called with the right inline_message_id');
          assert.isTrue(answer_cbq_spy.calledOnce, 'cbq.answer should have been called once');
        });
    });
  });

  suite('resend callback query handling', function () {
    test('from message', function () {
      const bot = new BusEtaBot();

      const answer_cbq_spy = sinon.spy();

      bot._callback_query_handlers['resend'].unshift((bot, cbq) => {
        cbq.answer = answer_cbq_spy;
      });

      const cbq = {
        "update_id": 100000000,
        "callback_query": {
          "id": "100000000000000000",
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "message": {
            "message_id": 85,
            "from": {"id": 100000000, "first_name": "Bus Eta Bot", "username": "BusEtaBot"},
            "chat": {"id": 100000000, "first_name": "fn", "username": "un", "type": "private"},
            "date": 1487566812,
            "text": "message_text",
            "entities": [{"type": "bold", "offset": 0, "length": 25}, {
              "type": "pre",
              "offset": 39,
              "length": 364
            }, {"type": "italic", "offset": 404, "length": 35}]
          },
          "chat_instance": "10000000000000000",
          "data": "{\"t\":\"resend\",\"b\":\"83151\"}"
        }
      };

      const send_message_spy = sinon.spy();

      bot.prepare_eta_message = function () {
        return Promise.resolve({
          send: send_message_spy
        });
      };

      return bot.handle(cbq)
        .then(() => {
          assert.isTrue(send_message_spy.calledOnce, 'send should have been called on the prepared message.');
          assert.isTrue(send_message_spy.calledWith(100000000), 'update_message should have been called with the right chat_id');
          assert.isTrue(answer_cbq_spy.calledOnce, 'cbq.answer should have been called once');
        });
    });
  });

  suite('start command handling', function () {
    test('reply with welcome message', function () {
      const spy = sinon.spy();

      const original = BusEtaBot.prepare_welcome_message;
      BusEtaBot.prepare_welcome_message = function () {
        return {
          send: spy
        };
      };

      const update = {
        "update_id": 100000000,
        "message": {
          "message_id": 1,
          "from": {"id": 100000000, "first_name": "Jiayu", "username": "un"},
          "chat": {"id": 100000000, "first_name": "Jiayu", "username": "un", "type": "private"},
          "date": 1486817921,
          "text": "/start",
          "entities": [{"type": "bot_command", "offset": 0, "length": 6}]
        }
      };

      const bot = new BusEtaBot();
      return bot.handle(update)
        .then(() => {
          assert.isTrue(spy.calledOnce, 'reply.send should have been called once');
        });
    });
  });

  suite('help command handling', function () {
    test('help command', function () {
      const update = {
        "update_id": 100000000,
        "message": {
          "message_id": 1,
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "chat": {"id": 100000000, "first_name": "fn", "username": "un", "type": "private"},
          "date": 1486817921,
          "text": "/help",
          "entities": [{"type": "bot_command", "offset": 0, "length": 5}]
        }
      };

      const spy = sinon.spy();
      BusEtaBot.prepare_help_message = function () {
        return {
          send: spy
        };
      };
      const bot = new BusEtaBot();
      return bot.handle(update)
        .then(() => {
          assert.isTrue(spy.calledOnce, 'reply.send should have been called');
        });
    });
  });

  suite('about command handling', function () {
    test('/about command', function () {
      const update = {
        "update_id": 100000000,
        "message": {
          "message_id": 1,
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "chat": {"id": 100000000, "first_name": "fn", "username": "un", "type": "private"},
          "date": 1486817921,
          "text": "/about",
          "entities": [{"type": "bot_command", "offset": 0, "length": 6}]
        }
      };

      const spy = sinon.spy();
      BusEtaBot.prepare_about_message = function () {
        return {
          send: spy
        };
      };
      const bot = new BusEtaBot();
      return bot.handle(update)
        .then(() => {
          assert.isTrue(spy.calledOnce, 'reply.send should have been called');
        });
    });
  });

  suite('version command handling', function () {
    test('/version command', function () {
      const update = {
        "update_id": 100000000,
        "message": {
          "message_id": 1,
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "chat": {"id": 100000000, "first_name": "fn", "username": "un", "type": "private"},
          "date": 1486817921,
          "text": "/version",
          "entities": [{"type": "bot_command", "offset": 0, "length": 8}]
        }
      };

      const spy = sinon.spy();
      BusEtaBot.prepare_about_message = function () {
        return {
          send: spy
        };
      };
      const bot = new BusEtaBot();
      return bot.handle(update)
        .then(() => {
          assert.isTrue(spy.calledOnce, 'reply.send should have been called');
        });
    });
  });

  suite('eta request handling', function () {
    test('bus stop only in text message', function () {
      const update = {
        "update_id": 100000000,
        "message": {
          "message_id": 1,
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "chat": {"id": 100000000, "first_name": "fn", "username": "un", "type": "private"},
          "date": 1486817921,
          "text": "96049"
        }
      };

      const bot = new BusEtaBot();

      const send_message_spy = sinon.spy();

      bot.prepare_eta_message = function () {
        return Promise.resolve({
          send: send_message_spy,
        });
      };

      return bot.handle(update)
        .then(() => {
          assert.isTrue(send_message_spy.calledOnce, 'reply.send should have been called once');
          assert.isTrue(send_message_spy.calledWith(100000000), 'reply.send should have been called with the right chat_id');
        });
    });
  });
});
