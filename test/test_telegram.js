"use strict";

import { OutgoingTextMessage } from '../src/lib/telegram';

suite('telegram', function () {
  suite('outgoing text message', function () {
    test('send message', function () {
      this.timeout(10000);

      const om = new OutgoingTextMessage('hello');
      return om.send(100000000)
        .then(console.log);
    });
  });
});
