"use strict";

import bus_eta_bot from '../src/bot/bus-eta-bot';
import { get_etas } from '../src/lib/datamall';
import BusEtaBot from '../src/bot/BusEtaBot';
import { OutgoingTextMessage } from "../src/lib/telegram";

suite('handle', function () {
  test('handle eta command', function () {
    return bus_eta_bot.handle({
      "update_id": 100000000,
      "message": {
        "message_id": 7,
        "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu"},
        "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu", "type": "private"},
        "date": 1486817921,
        "text": "/eta 96049 2 24",
        "entities": [{"type": "bot_command", "offset": 0, "length": 4}]
      }
    }).then(console.log)
      .catch(console.error);
  });

  test('handle voice', function () {
    return bus_eta_bot.handle({
      "update_id": 100000000,
      "message": {
        "message_id": 18,
        "from": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu"},
        "chat": {"id": 100000000, "first_name": "Jiayu", "username": "jiayu", "type": "private"},
        "date": 1486830113,
        "voice": {
          "duration": 1,
          "mime_type": "audio/ogg",
          "file_id": "AwADBQADAQADIGv5VLjzqdplobADAg",
          "file_size": 4924
        }
      }
    }).then(console.log)
      .catch(console.error);
  });

  test('aoeu', function () {
    const bus_stop = '96049';
    const chat_id = '10000';

    return get_etas(bus_stop)
      .then(etas => {
        if (etas.etas.length === 0) {
          const info = bus_eta_bot.datastore.get_bus_stop_info(bus_stop);
          if (info) {
            // if there were no etas for a bus stop but we have information about it
            return BusEtaBot.format_eta_message(etas, {services: service_nos, info})
              .serialise_send(chat_id);
          } else {
            // if there were no etas for a bus stop and it is not in our list of bus stops
            return new OutgoingTextMessage(`Sorry, I couldn't find any information about that bus stop.`)
              .serialise_send(chat_id);
          }
        }

        const info = bus_eta_bot.datastore.get_bus_stop_info(bus_stop);
        return BusEtaBot.format_eta_message(etas, {services: service_nos, info})
          .serialise_send(chat_id);
      })
      .then(console.log)
      .catch(() => console.error('error'));
  });
});
