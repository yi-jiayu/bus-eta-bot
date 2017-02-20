/**
 * Created by fn on 2/20/2017.
 */

import { assert } from 'chai';
import sinon from 'sinon';

import BusEtaBot from '../../src/bot/BusEtaBot';

suite('BusEtaBot', function () {
  suite('callback query handling', function () {
    test('callback query from message', function () {
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
          "data": "{\"t\":\"eta\",\"b\":\"83151\"}"
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

    test('callback query from inline query', function () {
      const bot = new BusEtaBot();

      const answer_cbq_spy = sinon.spy();

      bot._callback_query_handlers['eta'].unshift((bot, cbq) => {
        cbq.answer = answer_cbq_spy;
      });

      const cbq = {
        "update_id": 130514761,
        "callback_query": {
          "id": "1000000000000000",
          "from": {"id": 100000000, "first_name": "fn", "username": "un"},
          "inline_message_id": "aeoudhtns",
          "chat_instance": "1000000000",
          "data": "{\"t\":\"eta\",\"b\":\"62109\"}"
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
});
