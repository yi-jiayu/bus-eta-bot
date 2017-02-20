"use strict";

import request from 'request';

const endpoint = 'https://api.telegram.org/bot';


class TelegramMethod {
  constructor(method, params = {}) {
    this.method = method;
    this.params = params;
  }

  serialise() {
    const body = this.params;
    body.method = this.method;
    return body;
  }

  do() {
    if (!process.env.TELEGRAM_BOT_TOKEN) {
      console.error('warning: TELEGRAM_BOT_TOKEN not defined');
    }

    return new Promise((resolve, reject) => {
      request.post({
          uri: endpoint + process.env.TELEGRAM_BOT_TOKEN + '/' + this.method,
          body: this.params,
          json: true
        },
        (err, res, body) => {
          if (err) reject(err);
          else resolve(body);
        })
    });
  }
}

/**
 * Enum for Telegram message types
 * @enum {string}
 */
export const message_types = {
  TEXT: 'text',
  PHOTO: 'photo',
  VOICE: 'voice',
  LOCATION: 'location',
  VENUE: 'venue',
  /** Matches messages which don't match the types here */
  OTHER: 'other',
  /** Matches messages which don't have a handler registered */
  DEFAULT: 'unhandled'
};

export function get_updates(options = {}) {
  if (!process.env.TELEGRAM_BOT_TOKEN) {
    console.error('warning: TELEGRAM_BOT_TOKEN not defined');
  }

  return new Promise((resolve, reject) => {
    request({
      uri: endpoint + process.env.TELEGRAM_BOT_TOKEN + '/getUpdates',
      qs: options
    }, (err, res, body) => {
      if (err) reject(err);
      else if (res.statusCode >= 400) reject(body);
      else resolve(JSON.parse(body));
    });
  });
}

export class IncomingMessage {
  constructor(msg) {
    this.raw = msg;

    this.message_id = msg.message_id;
    this.chat_id = msg.chat.id;
    this.chat_type = msg.chat.type;
    this.user_id = msg.from.id;
    this.first_name = msg.from.first_name;
    this.last_name = msg.from.last_name || null;
    this.username = msg.from.username || null;
  }
}

export class IncomingTextMessage extends IncomingMessage {
  constructor(msg) {
    super(msg);

    this.text = msg.text;
    this.entities = msg.entities || [];
  }
}

export class IncomingPhotoMessage extends IncomingMessage {
  constructor(msg) {
    super(msg);

    this.photo = msg.photo;
    this.caption = msg.caption;
  }
}

export class IncomingVoiceMessage extends IncomingMessage {
  constructor(msg) {
    super(msg);

    this.voice = msg.voice;
  }
}

export class IncomingLocationMessage extends IncomingMessage {
  constructor(msg) {
    super(msg);

    this.location = msg.location;
  }
}

export class IncomingVenueMessage extends IncomingMessage {
  constructor(msg) {
    super(msg);

    this.venue = msg.venue;
    this.location = msg.venue.location;
  }
}

export class CallbackQuery {
  constructor(cbq) {
    this.raw = cbq;

    this.callback_query_id = cbq.id;
    this.message = cbq.hasOwnProperty('message') ? new IncomingMessage(cbq.message) : null;
    this.inline_message_id = cbq.inline_message_id || null;

    this.user_id = cbq.from.id;
    this.first_name = cbq.from.first_name;
    this.last_name = cbq.from.last_name || null;
    this.username = cbq.from.username || null;
    this.data = cbq.data || '';
  }

  /**
   * Updates the message which originated this callback query
   * @param {OutgoingTextMessage} msg
   */
  update(msg) {
    if (this.message) {
      return msg.update_message(this.message.chat, this.message.message_id);
    } else {
      return msg.update_inline_message(this.inline_message_id);
    }
  }
}

export class InlineQuery {
  constructor(ilq) {
    this.raw = ilq;

    this.inline_query_id = ilq.id;
    this.user_id = ilq.from.id;
    this.first_name = ilq.from.first_name;
    this.last_name = ilq.from.last_name || null;
    this.username = ilq.from.username || null;

    this.query = ilq.query;
  }
}

export class ChosenInlineResult {
  constructor(cir) {
    this.raw = cir;

    this.result_id = cir.result_id;
    this.user_id = cir.from.id;
    this.first_name = cir.from.first_name;
    this.last_name = cir.from.last_name;
    this.username = cir.from.username;

    this.inline_message_id = cir.inline_message_id;
    this.query = cir.query;
  }
}

export class OutgoingTextMessage {
  /**
   * Prepare a new text message to be sent by the bot
   * @param {string} text
   * @param {object} [config]
   */
  constructor(text = '', config = {}) {
    this.text = text;
    this.params = config;
  }

  /**
   * Sets the parse mode for this message
   * @param {string} mode
   */
  parse_mode(mode) {
    this.params.parse_mode = mode;
  }

  disable_web_page_preview(bool = true) {
    this.params.disable_web_page_preview = bool;
  }

  disable_notification(bool = true) {
    this.params.disable_notification = bool;
  }

  reply_to_message_id(id) {
    this.params.reply_to_message_id = id;
  }

  reply_markup(markup) {
    this.params.reply_markup = JSON.stringify(markup);
  }

  /**
   * Returns a sendMessage TelegramMethod object which can be executed or serialised.
   * @param chat_id
   * @return {TelegramMethod}
   * @private
   */
  _prepare_send(chat_id) {
    const params = {
      chat_id,
      text: this.text
    };

    const optional = ['parse_mode', 'disable_web_page_preview', 'disable_notification', 'reply_to_message_id', 'reply_markup'];
    for (const param of optional) {
      if (this.params[param]) {
        params[param] = this.params[param];
      }
    }

    return new TelegramMethod('sendMessage', params);
  }

  /**
   * Returns an updateMessage TelegramMethod object which can be executed or serialised.
   * @return {TelegramMethod}
   * @private
   */
  _prepare_update() {
    const params = {
      text: this.text
    };

    const optional = ['parse_mode', 'disable_web_page_preview', 'reply_markup'];
    for (const param of optional) {
      if (this.params[param]) {
        params[param] = this.params[param];
      }
    }

    return new TelegramMethod('editMessageText', params);
  }

  send(chat_id) {
    return this._prepare_send(chat_id).do();
  }

  /**
   * Update the message with chat_id and message_id with this message.
   * @param chat_id
   * @param message_id
   * @return {Promise}
   */
  update_message(chat_id, message_id) {
    const update = this._prepare_update();
    update.params.chat_id = chat_id;
    update.params.message_id = message_id;

    return update.do();
  }

  /**
   * Update the inline message with inline_message_id with this message.
   * @param inline_message_id
   * @return {Promise}
   */
  update_inline_message(inline_message_id) {
    const update = this._prepare_update();
    update.params.inline_message_id = inline_message_id;

    return update.do();
  }

  serialise_send(chat_id) {
    return this._prepare_send(chat_id).serialise();
  }

  serialise_update_message(chat_id, message_id) {
    const update = this._prepare_update();
    update.params.chat_id = chat_id;
    update.params.message_id = message_id;

    return update.serialise();
  }

  serialise_update_inline_message(inline_message_id) {
    const update = this._prepare_update();
    update.params.inline_message_id = inline_message_id;

    return update.serialise();
  }
}

/**
 * @typedef {object} InlineKeyboardButton
 * @prop {string} text
 * @prop {string} url
 * @prop {string} callback_data
 * @prop {string} switch_inline_query
 * @prop {string} switch_inline_query_current_chat
 * @prop callback_game
 */

export class InlineKeyboardMarkup {
  /**
   * Create a new inline keyboard markup options object
   * @param {InlineKeyboardButton[][]} buttons - Array of array of inline keyboard buttons
   */
  constructor(buttons) {
    this.inline_keyboard = buttons;
  }
}

/**
 * @typedef {object} InputTextMessageContent
 * @prop {string} message_text
 * @prop {string} [parse_mode]
 * @prop {bool} [disable_web_page_previews]
 */

class InlineQueryResult {
  /**
   * Base class for InlineQueryResults
   * @param {string} type
   * @param {string} id
   */
  constructor(type, id) {
    this.type = type;
    this.id = id;
  }
}

export class InlineQueryResultLocation extends InlineQueryResult {
  /**
   * Create a new InlineQueryResultLocation
   * @param {string} id
   * @param {string} title
   * @param {number} lat
   * @param {number} lon
   * @param {object} [options]
   * @param {InlineKeyboardMarkup} [options.reply_markup]
   * @param {InputTextMessageContent} [options.input_message_content]
   * @param {string} [options.thumb_url]
   */
  constructor(id, title, lat, lon, options = {}) {
    super('location', id);

    this.title = title;
    this.latitude = lat;
    this.longitude = lon;

    // optional attributes
    const optional = ['reply_markup', 'input_message_content', 'thumb_url', 'thumb_width', 'thumb_height'];
    for (const attr of optional) {
      if (options.hasOwnProperty(attr)) {
        this[attr] = options[attr];
      }
    }
  }
}

export class InlineQueryResultVenue extends InlineQueryResult {
  /**
   * Create a new InlineQueryResultVenue
   * @param {string} id
   * @param {string} title
   * @param {string} address
   * @param {number} lat
   * @param {number} lon
   * @param {object} [options]
   * @param {InlineKeyboardMarkup} [options.reply_markup]
   * @param {InputTextMessageContent} [options.input_message_content]
   * @param {string} [options.thumb_url]
   */
  constructor(id, title, address, lat, lon, options) {
    super('venue', id);

    this.title = title;
    this.address = address;
    this.latitude = lat;
    this.longitude = lon;

    // optional attributes
    const optional = ['foursquare_id', 'reply_markup', 'input_message_content', 'thumb_url', 'thumb_width', 'thumb_height'];
    for (const attr of optional) {
      if (options.hasOwnProperty(attr)) {
        this[attr] = options[attr];
      }
    }
  }
}

export class InlineQueryResultArticle extends InlineQueryResult {
  /**
   * Create a new InlineQueryResultArticle
   * @param {string} id - Unique identifier for this result, 1-64 Bytes
   * @param {string} title - Title of the result
   * @param {InputTextMessageContent} content - Content of the message to be sent
   * @param {object} [options]
   * @param {InlineKeyboardMarkup} [options.reply_markup]
   * @param {string} [options.url]
   * @param {boolean} [options.hide_url]
   * @param {string} [options.description]
   */
  constructor(id, title, content, options) {
    super('article', id);

    this.title = title;
    this.input_message_content = content;

    // optional attributes
    const optional = ['reply_markup', 'url', 'hide_url', 'description'];
    for (const attr of optional) {
      if (options[attr]) {
        this[attr] = options[attr];
      }
    }
  }
}

export class InlineQueryAnswer {
  constructor(results, options = {}) {
    this.results = results;
    this.params = options;
  }

  _prepare_answer(inline_query_id) {
    const params = {
      inline_query_id,
      results: JSON.stringify(this.results)
    };

    const optional = ['cache_time', 'is_personal', 'next_offset', 'switch_pm_text', 'switch_pm_parameter'];
    for (const param of optional) {
      if (this.params.hasOwnProperty(param)) {
        params[param] = this.params[param];
      }
    }

    return new TelegramMethod('answerInlineQuery', params);
  }

  send(inline_query_id) {
    return this._prepare_answer(inline_query_id).do();
  }

  serialise(inline_query_id) {
    return this._prepare_answer(inline_query_id).serialise();
  }
}
