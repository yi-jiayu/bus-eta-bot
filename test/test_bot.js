"use strict";

import { assert } from 'chai';
import * as sinon from 'sinon';

import Bot from '../src/lib/Bot';
import { message_types } from '../src/lib/telegram';

const updates = {
  command: {
    "update_id": 100000000, "message": {
      "message_id": 5,
      "from": {"id": 100000000, "first_name": "first_name", "username": "username"},
      "chat": {"id": 100000000, "first_name": "first_name", "username": "username", "type": "private"},
      "date": 1000000000,
      "text": "/version",
      "entities": [{"type": "bot_command", "offset": 0, "length": 8}]
    }
  },
  command_with_args: {
    "update_id": 100000000, "message": {
      "message_id": 5,
      "from": {"id": 100000000, "first_name": "first_name", "username": "username"},
      "chat": {"id": 100000000, "first_name": "first_name", "username": "username", "type": "private"},
      "date": 1000000000,
      "text": "/version two  two",
      "entities": [{"type": "bot_command", "offset": 0, "length": 8}]
    }
  },
  text_message: {
    "update_id": 100000000, "message": {
      "message_id": 5,
      "from": {"id": 100000000, "first_name": "first_name", "username": "username"},
      "chat": {"id": 100000000, "first_name": "first_name", "username": "username", "type": "private"},
      "date": 1000000000,
      "text": "version"
    }
  },
  callback_query: {
    "update_id": 100000000, "callback_query": {
      "id": "100000000000000000",
      "from": {"id": 100000000, "first_name": "first_name", "username": "username"},
      "message": {
        "message_id": 6,
        "from": {"id": 100000000, "first_name": "first_name", "username": "username"},
        "chat": {"id": 100000000, "first_name": "first_name", "username": "username", "type": "private"},
        "date": 1000000000,
        "text": "text",
        "entities": [{"type": "pre", "offset": 24, "length": 89}]
      },
      "chat_instance": "-100000000000000000",
      "data": "{\"t\": \"eta\", \"d\":false, \"a\": \"81111 155\"}"
    }
  },
  inline_query: {
    "update_id": 100000000,
    "inline_query": {
      "id": "100000000000000000",
      "from": {"id": 100000000, "first_name": "first_name", "username": "username"},
      "location": {"latitude": 1.0, "longitude": 100.0},
      "query": "query",
      "offset": ""
    }
  },
  chosen_inline_result: {
    "update_id": 100000000,
    "chosen_inline_result": {
      "from": {"id": 100000000, "first_name": "fn", "username": "un"},
      "location": {"latitude": 1.0, "longitude": 100.0},
      "query": "q",
      "result_id": "10000"
    }
  }
};


suite('Bot', function () {
  suite('update classification', function () {
    test('bot command', function () {
      assert.isTrue(Bot._is_command(updates.command));
    });

    test('message', function () {
      assert.isTrue(Bot._is_message(updates.text_message));
    });

    test('callback query', function () {
      assert.isTrue(Bot._is_callback_query(updates.callback_query));
    });

    test('inline query', function () {
      assert.isTrue(Bot._is_inline_query(updates.inline_query));
    });

    test('chosen inline result', function () {
      assert.isTrue(Bot._is_chosen_inline_result(updates.chosen_inline_result));
    });
  });

  suite('update dispatch', function () {
    test('dispatch command without args', function () {
      const spy = sinon.spy();
      const bot = new Bot();

      bot.command('version', spy);

      return bot.handle(updates.command)
        .then(() => {
          assert.strictEqual(spy.args[0][0], bot, 'first arg of handler should refer to bot');
          assert.strictEqual(spy.args[0][1].raw, updates.command.message, 'second arg should contain the original message');
          assert.strictEqual(spy.args[0][2], '', 'third arg should be an empty string since there were no args behind the command');
        });
    });

    test('dispatch command with args', function () {
      const spy = sinon.spy();
      const bot = new Bot();

      bot.command('version', spy);

      return bot.handle(updates.command_with_args)
        .then(() => {
          assert.strictEqual(spy.args[0][0], bot, 'first arg of handler should refer to bot');
          assert.strictEqual(spy.args[0][1].raw, updates.command_with_args.message, 'second arg should contain the original message');
          assert.strictEqual(spy.args[0][2], 'two two', 'should have args, args should be trimmed');
        });
    });

    test('dispatch text message', function () {
      const spy = sinon.spy();
      const bot = new Bot();

      bot.message(message_types.TEXT, spy);

      return bot.handle(updates.text_message)
        .then(() => {
          assert.strictEqual(spy.args[0][0], bot, 'spy should be called with bot');
          assert.strictEqual(spy.args[0][1].raw, updates.text_message.message);
        });
    });

    test('dispatch callback query', function () {
      const spy = sinon.spy();
      const bot = new Bot();

      bot.callback_query('eta', spy);

      return bot.handle(updates.callback_query)
        .then(() => {
          assert.strictEqual(spy.args[0][0], bot);
          assert.strictEqual(spy.args[0][1].raw, updates.callback_query.callback_query);
        });
    });

    test('dispatch inline query', function () {
      const spy = sinon.spy();
      const bot = new Bot();

      bot.inline_query(spy);

      return bot.handle(updates.inline_query)
        .then(() => {
          assert.isTrue(spy.calledWith(bot));
          assert.strictEqual(spy.args[0][1].raw, updates.inline_query.inline_query);
        });
    });

    test('dispatch chosen inline result', function () {
      const spy = sinon.spy();
      const bot = new Bot();

      bot.chosen_inline_result(spy);

      return bot.handle(updates.chosen_inline_result)
        .then(() => {
          assert.isTrue(spy.calledWith(bot));
          assert.strictEqual(spy.args[0][1].raw, updates.chosen_inline_result.chosen_inline_result);
        });
    });
  });
});
