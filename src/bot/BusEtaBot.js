"use strict";

import { table, getBorderCharacters } from 'table';
import moment from 'moment-timezone';
import { message_types, OutgoingTextMessage, InlineKeyboardMarkup } from '../lib/telegram';

export default class BusEtaBot {
  /**
   * Returns an OutgoingTextMessage which can be sent or used to refresh bus etas
   * @param {string} query
   * @returns {OutgoingTextMessage}
   */
  static answer_eta_query_timed(query) {
    const timings = [];

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
      header = `*${info.description} (${etas.bus_stop})*
${info.road}`;
    } else {
      header = `*${etas.bus_stop}*`;
    }

    const text = `${header}
\`\`\`\n${eta_table}\`\`\`
_${updated_time}_`;

    const message = new OutgoingTextMessage(text);
    const callback_data = {
      t: 'eta',
      b: etas.bus_stop
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
   * @returns {{bus_stop: string, service_nos: string[]}}
   */
  static infer_bus_stop_and_service_nos(text) {
    text = text.toUpperCase();
    let [bus_stop, ...service_nos] = text.split(' ');
    bus_stop = bus_stop.substr(0, 5);

    return {bus_stop, service_nos: service_nos};
  }
}
