"use strict";

import { table, getBorderCharacters } from 'table';
import moment from 'moment-timezone';

import Bot from '../lib/Bot';
import Datastore from './Datastore';
import Analytics from './Analytics';
import EtaProvider from './EtaProvider';
import {
  message_types,
  OutgoingTextMessage,
  InlineQueryAnswer,
  InlineKeyboardMarkup,
  InlineQueryResultLocation
} from '../lib/telegram';

import strings from './strings';

export default class BusEtaBot extends Bot {
  constructor(options = {}) {
    super();

    const datastore = options.datastore || new Datastore();
    const analytics = options.analytics || new Analytics();
    const eta_provider = options.eta_provider || new EtaProvider();

    this.datastore = datastore;
    this.analytics = analytics;
    this.eta_provider = eta_provider;

    // start command handler
    this.command('start', (bot, msg) => {
      const first_name = msg.first_name;
      const chat_id = msg.chat_id;

      return BusEtaBot.prepare_welcome_message(first_name)
        .send(chat_id);
    });

    // version command handler
    this.command('version', (bot, msg) => {
      const chat_id = msg.chat_id;

      const reply = BusEtaBot.prepare_about_message();
      return reply.send(chat_id);
    });

    this.command('about', (bot, msg) => {
      const chat_id = msg.chat_id;

      const reply = BusEtaBot.prepare_about_message();
      return reply.send(chat_id);
    });

    // todo: feedback command handler

    this.command('help', (bot, msg) => {
      const chat_id = msg.chat_id;

      return BusEtaBot.prepare_help_message()
        .send(chat_id);
    });

    // text message handler
    this.message(message_types.TEXT, (bot, msg) => {
      const chat_id = msg.chat_id;

      // trim incoming text to an arbitrary limit of 100 chars
      const text = msg.text.substr(0, 100);

      // let the bus stop code be the first word or the first 5 digits, whichever is shorter
      // and the service numbers be the subsequent words
      const {bus_stop_id, service_nos} = BusEtaBot.infer_bus_stop_and_service_nos(text);

      return this.prepare_eta_message(bus_stop_id, service_nos)
        .then(reply => reply.send(chat_id));
    });

    // eta command handler
    this.command('eta', (bot, msg, args) => {
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
      const {bus_stop_id, service_nos} = BusEtaBot.infer_bus_stop_and_service_nos(args);

      return this.prepare_eta_message(bus_stop_id, service_nos)
        .then(reply => reply.send(chat_id));
    });

    // callback query handler
    this.callback_query('eta', (bot, cbq) => {
      const cbq_from_ilq = cbq.inline_message_id !== null;

      const cbq_data = JSON.parse(cbq.data);
      const {b: bus_stop, s: service_nos} = cbq_data;

      return this.prepare_eta_message(bus_stop, service_nos)
        .then(reply => {
          let update;

          if (cbq_from_ilq) {
            update = reply.update_inline_message(cbq.inline_message_id);
          } else {
            update = reply.update_message(cbq.message.chat_id, cbq.message.message_id);
          }

          return Promise.all([update, cbq.answer({text: 'Etas updated!'})]);
        });
    });

    // eta_demo handler
    this.callback_query('eta_demo', (bot, cbq) => {
      const bus_stop = '96049';

      return this.prepare_eta_message(bus_stop)
        .then(reply => {
          const send = reply.send(cbq.message.chat_id);

          return Promise.all([send, cbq.answer()]);
        });
    });

    // inline query handler
    this.inline_query((bot, ilq) => {
      const inline_query_id = ilq.inline_query_id;

      return this.answer_inline_query(ilq)
        .then(answer => {
          if (answer !== null) {
            return answer.send(inline_query_id);
          } else {
            console.error('info: inline query returned no results');
          }
        });
    });

    // chosen inline result handler
    this.chosen_inline_result((bot, cir) => {
      const inline_message_id = cir.inline_message_id;
      if (!inline_message_id) {
        console.error('warning: received a chosen inline result without inline_message_id');
        return;
      }

      const bus_stop = cir.result_id;
      return this.prepare_eta_message(bus_stop)
        .then(reply => reply.update_inline_message(inline_message_id));
    });
  }

  prepare_eta_message(bus_stop, services) {
    return this.eta_provider.get_etas(bus_stop)
      .then(etas => {
        if (etas.etas.length === 0) {
          return this.datastore.get_bus_stop_info(bus_stop)
            .then(info => {
              if (info) {
                // if there were no etas for a bus stop but we have information about it
                return BusEtaBot.format_eta_message(etas, {services: services, info})
              } else {
                // if there were no etas for a bus stop and it is not in our list of bus stops
                return new OutgoingTextMessage(`Sorry, I couldn't find any information about that bus stop.`)
              }
            });
        } else {
          return this.datastore.get_bus_stop_info(bus_stop)
            .then(info => BusEtaBot.format_eta_message(etas, {services: services, info}));
        }
      });
  }

  /**
   * Respond to an inline query
   * @param {InlineQuery} ilq - Inline query
   * @return {Promise}
   */
  answer_inline_query(ilq) {
    const query = ilq.query;
    const location = ilq.location;

    if (query.length === 0 && location) {
      // if the inline query is empty and we have a location, we respond with nearby bus stops
      return this.datastore.get_nearby_bus_stops(location.latitude, location.longitude)
        .then(nearby => {
          if (nearby.length === 0) {
            // if we can't find any nearby bus stops, just default to sending the completions for a blank query
            console.log("prepare_inline_query: couldn't find any nearby bus stops, defaulting to completions");
            return this.datastore.get_completions(query)
            // don't cache results for empty queries or queries using location
              .then(completions => BusEtaBot.prepare_inline_query_answer(completions, {
                cache_time: 0,
                next_offset: ''
              }));
          } else {
            // don't cache results for empty queries or queries using location
            console.log(`prepare_inline_query: sending ${nearby.length} results for nearby bus stops`);
            return BusEtaBot.prepare_inline_query_answer(nearby, {cache_time: 0, is_personal: true, next_offset: ''});
          }
        });
    }

    return this.datastore.get_completions(query)
      .then(BusEtaBot.prepare_inline_query_answer);
  };

  static prepare_help_message() {
    const help = strings.help;
    return new OutgoingTextMessage(help);
  }

  static prepare_about_message() {
    const about = strings.about;
    const text = about.replace('$VERSION', require('../../package.json').version);

    return new OutgoingTextMessage(text);
  }

  static prepare_welcome_message(first_name) {
    const welcome = strings.welcome;
    const text = welcome.replace('FIRST_NAME', first_name);

    const reply = new OutgoingTextMessage(text);
    reply.parse_mode('markdown');

    const markup = new InlineKeyboardMarkup([[{
      text: 'Get etas for bus stop 96049',
      callback_data: JSON.stringify({t: 'eta_demo'})
    }], [{
      text: 'Try an inline query',
      switch_inline_query_current_chat: 'Changi'
    }]]);
    reply.reply_markup(markup);

    return reply;
  }

  /**
   * Convert the result of an eta query to a message to be sent or updated on Telegram
   * @param {ParsedEtas} etas
   * @param {object} [options]
   * @param {string[]} [options.services]
   * @param {object} [options.info}
   * @param {string} options.info.description
   * @param {string} options.info.road
   * @returns {OutgoingTextMessage}
   */
  static format_eta_message(etas, options = {}) {
    const services = options.services || [];
    const info = options.info;
    if (info) {
      if (!info.road || !info.description) {
        throw new TypeError('info must contain description and road');
      }
    }

    const eta_array = [['Svc', 'Next', '2nd', '3rd']];
    let excluded_services = 0;
    let valid_services = [];
    for (const eta of etas.etas) {
      // if services is provided, we return only etas for the specified services
      if (services.length > 0) {
        if (services.indexOf(eta.svc_no) === -1) {
          excluded_services++;
          continue
        } else {
          valid_services.push(eta.svc_no);
        }
      }

      eta_array.push([eta.svc_no, eta.next, eta.subsequent, eta.third]);
    }

    let eta_table;
    if (eta_array.length > 8) {
      // if there are more than 8 buses to be shown, leave out the borders to save space
      eta_table = table(eta_array, {
        border: getBorderCharacters('void'),
        columnDefault: {
          paddingLeft: 0,
          paddingRight: 2
        },
        drawHorizontalLine: () => {
          return false
        }
      })
    } else {
      eta_table = table(eta_array, {
        border: getBorderCharacters('ramac')
      });
    }


    if (excluded_services > 0) {
      eta_table += `${excluded_services} more ${excluded_services === 1 ? 'service' : 'services'} not shown.`;
    }

    const updated_time = `Last updated: ${moment(etas.updated).tz('Asia/Singapore').format('lll')}.`;

    let header;
    if (options.info) {
      header = `*${info.description} (${etas.bus_stop_id})*
${info.road}`;
    } else {
      header = `*${etas.bus_stop_id}*`;
    }

    const text = `${header}
\`\`\`\n${eta_table}\`\`\`
_${updated_time}_`;

    const message = new OutgoingTextMessage(text);
    const callback_data = {
      t: 'eta',
      b: etas.bus_stop_id
    };

    if (valid_services.length > 0) {
      callback_data.s = services;
    }

    const markup = new InlineKeyboardMarkup([[{
      text: 'Refresh',
      callback_data: JSON.stringify(callback_data)
    }]]);
    message.reply_markup(markup);
    message.parse_mode('markdown');

    return message;
  }

  /**
   * Infer a bus stop code and an array of service numbers to get etas for from an incoming text
   * @param {string} text
   * @returns {{bus_stop_id: string, service_nos: string[]}}
   */
  static infer_bus_stop_and_service_nos(text) {
    text = text.toUpperCase();
    let [bus_stop_id, ...service_nos] = text.split(' ');
    bus_stop_id = bus_stop_id.substr(0, 5);

    return {bus_stop_id, service_nos: service_nos};
  }

  /**
   *
   * @param {BusStopInfo[]} completions
   * @param options
   * @return {?InlineQueryAnswer}
   */
  static prepare_inline_query_answer(completions, options) {
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

      options = options || {cache_time: 86400, next_offset: ''};

      return new InlineQueryAnswer(results, options);
    } else {
      return null;
    }
  }
}
