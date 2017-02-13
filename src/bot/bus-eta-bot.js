"use strict";

import { get_etas } from '../lib/datamall';

import Bot from '../lib/Bot';
import BusEtaBot from './BusEtaBot';
import LocalDatastore from '../lib/LocalDatastore';
import Analytics from '../lib/Analytics';
import {
  message_types,
  OutgoingTextMessage,
  InlineQueryAnswer,
  InlineKeyboardMarkup,
  InlineQueryResultLocation
} from '../lib/telegram';

const bot = new Bot();

bot.datastore = new LocalDatastore(require('../data/bus-stop-info.json'), require('../data/location-data.json'));
bot.analytics = new Analytics();

// start command handler
bot.command('start', (bot, msg) => {
  const first_name = msg.first_name;
  const chat_id = msg.chat_id;

  const text = `*Hello ${first_name}*,
  
Bus Eta Bot is a simple bot which can tell you the estimated time you have to wait for buses in Singapore. Its information comes from the same source as the official LTA MyTransport app and many other bus arrival time apps through the LTA Datamall API.

To get started, send me a bus stop code or try /help to view available commands. I hope you will find Bus Eta Bot useful!`;

  const reply = new OutgoingTextMessage(text);
  reply.parse_mode('markdown');

  return reply.serialise_send(chat_id);
});

// text message handler
bot.message(message_types.TEXT, (bot, msg) => {
  const chat_id = msg.chat_id;

  // trim incoming text to an arbitrary limit of 100 chars
  const text = msg.text.substr(0, 100);

  // let the bus stop code be the first word or the first 5 digits, whichever is shorter
  // and the service numbers be the subsequent words
  const {bus_stop, service_nos} = BusEtaBot.infer_bus_stop_and_service_nos(text);

  return get_etas(bus_stop)
    .then(etas => {
      if (etas.etas.length === 0) {
        return bot.datastore.get_bus_stop_info(bus_stop)
          .then(info => {
            if (info) {
              // if there were no etas for a bus stop but we have information about it
              return BusEtaBot.format_eta_message(etas, {services: service_nos, info})
            } else {
              // if there were no etas for a bus stop and it is not in our list of bus stops
              return new OutgoingTextMessage(`Sorry, I couldn't find any information about that bus stop.`)
            }
          });
      } else {
        return bot.datastore.get_bus_stop_info(bus_stop)
          .then(info => BusEtaBot.format_eta_message(etas, {services: service_nos, info}));
      }
    })
    .then(reply => reply.send(chat_id));
});

// eta command handler
bot.command('eta', (bot, msg, args) => {
  const chat_id = msg.chat_id;
  const message_id = msg.message_id;

  if (args.length === 0) {
    return new OutgoingTextMessage('Please send me a bus stop code to get etas for.',
      {
        reply_to_message_id: message_id,
        reply_markup: JSON.stringify({
          force_reply: true,
          selective: true
        })
      })
      .send(chat_id);
  }

  // trim incoming command args to an arbitrary limit of 100 chars
  args = args.substr(0, 100);

  // let the bus stop code be the first word or the first 5 digits, whichever is shorter
  // and the service numbers be the subsequent words
  const {bus_stop, service_nos} = BusEtaBot.infer_bus_stop_and_service_nos(args);

  return get_etas(bus_stop)
    .then(etas => {
      if (etas.etas.length === 0) {
        return bot.datastore.get_bus_stop_info(bus_stop)
          .then(info => {
            if (info) {
              // if there were no etas for a bus stop but we have information about it
              return BusEtaBot.format_eta_message(etas, {services: service_nos, info})
            } else {
              // if there were no etas for a bus stop and it is not in our list of bus stops
              return new OutgoingTextMessage(`Sorry, I couldn't find any information about that bus stop.`)
            }
          });
      } else {
        return bot.datastore.get_bus_stop_info(bus_stop)
          .then(info => BusEtaBot.format_eta_message(etas, {services: service_nos, info}));
      }
    })
    .then(reply => reply.send(chat_id));
});

// callback query handler
bot.callback_query('eta', (bot, cbq) => {
  const cbq_from_ilq = cbq.inline_message_id !== null;

  const cbq_data = JSON.parse(cbq.data);
  const {b: bus_stop, s: service_nos} = cbq_data;

  return get_etas(bus_stop)
    .then(etas => {
      if (etas.etas.length === 0) {
        return bot.datastore.get_bus_stop_info(bus_stop)
          .then(info => {
            if (info) {
              // if there were no etas for a bus stop but we have information about it
              return BusEtaBot.format_eta_message(etas, {services: service_nos, info})
            } else {
              // if there were no etas for a bus stop and it is not in our list of bus stops
              return new OutgoingTextMessage(`Sorry, I couldn't find any information about that bus stop.`)
            }
          });
      } else {
        return bot.datastore.get_bus_stop_info(bus_stop)
          .then(info => BusEtaBot.format_eta_message(etas, {services: service_nos, info}));
      }
    })
    .then(reply => {
      if (cbq_from_ilq) {
        return reply.update_inline_message(cbq.inline_message_id);
      } else {
        return reply.update_message(cbq.message.chat_id, cbq.message.message_id);
      }
    });
});

bot.inline_query((bot, ilq) => {
  const inline_query_id = ilq.inline_query_id;
  const query = ilq.query;

  return bot.datastore.get_completions(query)
    .then(completions => {
      if (completions.length > 0) {
        const results = [];

        for (const c of completions) {
          const callback_data = {t: 'eta', b: c.id};

          const markup = new InlineKeyboardMarkup([[{
            text: 'Refresh',
            callback_data: JSON.stringify(callback_data)
          }]]);

          const r = new InlineQueryResultLocation(
            c.id,
            `${c.description} (${c.id})`,
            c.latitude,
            c.longitude,
            {
              reply_markup: markup,
              input_message_content: {
                message_text: `*${c.description} (${c.id})*\n${c.road}\n\n\`Fetching etas...\``,
                parse_mode: 'markdown'
              }
            });
          results.push(r);
        }

        const answer = new InlineQueryAnswer(inline_query_id, results, {cache_time: 86400, next_offset: ''});
        return answer.send();
      } else {
        console.log('info: inline query returned no results');
      }
    });
});

bot.chosen_inline_result((bot, cir) => {
  const inline_message_id = cir.inline_message_id;
  if (!inline_message_id) {
    console.error('warning: received a chosen inline result without inline_message_id');
    return;
  }

  const bus_stop = cir.result_id;
  return get_etas(bus_stop)
    .then(etas => {
      if (etas.etas.length === 0) {
        return bot.datastore.get_bus_stop_info(bus_stop)
          .then(info => {
            if (info) {
              // if there were no etas for a bus stop but we have information about it
              return BusEtaBot.format_eta_message(etas, {info})
            } else {
              // if there were no etas for a bus stop and it is not in our list of bus stops
              return new OutgoingTextMessage(`Sorry, I couldn't find any information about that bus stop.`)
            }
          });
      } else {
        return bot.datastore.get_bus_stop_info(bus_stop)
          .then(info => BusEtaBot.format_eta_message(etas, {info}));
      }
    })
    .then(reply => reply.update_inline_message(inline_message_id));
});

export default bot;
