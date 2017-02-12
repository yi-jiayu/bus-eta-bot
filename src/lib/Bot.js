"use strict";

import debug from 'debug';
import {
  message_types,
  IncomingMessage,
  IncomingTextMessage,
  IncomingPhotoMessage,
  IncomingVoiceMessage,
  IncomingLocationMessage,
  IncomingVenueMessage,
  CallbackQuery,
  InlineQuery,
  ChosenInlineResult
} from './telegram';

const log = debug('bot');

/**
 * @callback CommandHandler
 * @param {Bot} bot
 * @param {IncomingTextMessage} msg
 * @param {string} args
 * @returns {Promise}
 */

/**
 * @callback MessageHandler
 * @param {Bot} bot
 * @param {IncomingMessage|IncomingTextMessage} msg
 * @returns {Promise}
 */

/**
 * @callback CallbackQueryHandler
 * @param {Bot} bot
 * @param {CallbackQuery} cbq
 * @returns {Promise}
 */

/**
 * @callback InlineQueryHandler
 * @param {Bot} bot
 * @param {InlineQuery} ilq
 * @returns {Promise}
 */

export default class Bot {
  constructor() {
    this._command_handlers = {};
    this._message_handlers = {
      // workaround so that the bot doesn't throw an error on receiving an unrecognised message type
      [message_types.OTHER]: []
    };
    this._callback_query_handlers = {};
    this._inline_query_handler = [];
    this._chosen_inline_result_handler = [];
  }

  /**
   * Registers a new command handler
   * @param {string} command
   * @param {CommandHandler} handler
   */
  command(command, ...handler) {
    if (this._command_handlers.hasOwnProperty(command)) {
      this._command_handlers = this._command_handlers[command].concat(handler);
    } else {
      this._command_handlers[command] = handler;
    }
  }

  /**
   * Registers a new message handler
   * @param {string} type
   * @param {MessageHandler} handler
   */
  message(type, ...handler) {
    if (this._message_handlers.hasOwnProperty(type)) {
      this._message_handlers[type] = this._message_handlers[type].concat(handler);
    } else {
      this._message_handlers[type] = handler;
    }
  }

  /**
   * Registers a new callback query handler
   * @param {string} type
   * @param {CallbackQueryHandler} handler
   */
  callback_query(type, ...handler) {
    if (this._callback_query_handlers.hasOwnProperty(type)) {
      this._callback_query_handlers[type] = this._callback_query_handlers[type].concat(handler);
    } else {
      this._callback_query_handlers[type] = handler;
    }
  }

  /**
   * Registers an inline query handler
   * @param {InlineQueryHandler} handler
   */
  inline_query(...handler) {
    this._inline_query_handler = this._inline_query_handler.concat(handler);
  }

  /**
   * Registers a handler for chosen inline results
   * @param handler
   */
  chosen_inline_result(...handler) {
    this._chosen_inline_result_handler = this._chosen_inline_result_handler.concat(handler);
  }

  /**
   * Handle an incoming update.
   * @param update
   * @return {Promise}
   */
  handle(update) {
    if (Bot._is_command(update)) {
      return this._dispatch_command(update);
    } else if (Bot._is_message(update)) {
      return this._dispatch_message(update);
    } else if (Bot._is_callback_query(update)) {
      return this._dispatch_callback_query(update);
    } else if (Bot._is_inline_query(update)) {
      return this._dispatch_inline_query(update);
    } else if (Bot._is_chosen_inline_result(update)) {
      return this._dispatch_chosen_inline_result(update);
    } else {
      console.error('warning: unrecognised update ' + JSON.stringify(update));
      return Promise.resolve();
    }
  }

  /**
   * Dispatches an update containing a bot command to the registered command handlers.
   * @param update
   * @return {Promise}
   */
  _dispatch_command(update) {
    const command_entity = update.message.entities.find(e => e.type === 'bot_command');

    let command = update.message.text.substr(command_entity.offset, command_entity.length);
    let args = update.message.text.substr(command_entity.offset + command_entity.length);

    // strip tags and drop the beginning slash from the command
    command = command.replace(/@\w+/g, '').substr(1);

    // collapse whitespace and trim args
    args = args.replace(/\s+/g, ' ').trim();

    if (this._command_handlers.hasOwnProperty(command)) {
      const msg = new IncomingTextMessage(update.message);

      return this._command_handlers[command]
        .reduce((p, next) => p.then(() => next(this, msg, args), err => Promise.reject(err)), Promise.resolve());
    } else {
      console.error(`warning: no command handler registered for ${command}`);
      return Promise.resolve();
    }
  }

  /**
   * Dispatches an update containing a message to the registered message handlers.
   * @param update
   * @return {Promise}
   */
  _dispatch_message(update) {
    const message = update.message;
    const message_type = Bot._get_message_type(update);

    if (this._message_handlers.hasOwnProperty(message_type)) {
      let msg;

      switch (message_type) {
        case (message_types.TEXT):
          msg = new IncomingTextMessage(message);
          break;
        case (message_types.PHOTO):
          msg = new IncomingPhotoMessage(message);
          break;
        case (message_types.VOICE):
          msg = new IncomingVoiceMessage(message);
          break;
        case (message_types.LOCATION):
          msg = new IncomingLocationMessage(message);
          break;
        case (message_types.VENUE):
          msg = new IncomingVenueMessage(message);
          break;
        case (message_types.OTHER):
        default:
          msg = new IncomingMessage(message);
      }

      return this._message_handlers[message_type]
        .reduce((p, next) => p.then(() => next(this, msg), err => Promise.reject(err)), Promise.resolve());
    } else {
      console.error('warning: no message handler registered for message ' + JSON.stringify(update));
      return Promise.resolve();
    }
  }

  /**
   * Dispatches an update containing an inline query to the registered inline query handlers.
   * @param update
   * @return {Promise}
   * @private
   */
  _dispatch_callback_query(update) {
    const cbq_type = Bot._get_cbq_type(update);

    if (this._callback_query_handlers.hasOwnProperty(cbq_type)) {
      const cbq = new CallbackQuery(update.callback_query);

      return this._callback_query_handlers[cbq_type]
        .reduce((p, next) => p.then(() => next(this, cbq), err => Promise.reject(err)), Promise.resolve());
    } else {
      console.error(`warning: no callback query handler registered for "${update.callback_query.data}"`);
      return Promise.resolve();
    }
  }

  /**
   * Dispatches an update containing an inline query to the registered inline query handler
   * @param update
   * @return {Promise}
   * @private
   */
  _dispatch_inline_query(update) {
    if (this._inline_query_handler.length > 0) {
      const ilq = new InlineQuery(update.inline_query);

      return this._inline_query_handler
        .reduce((p, next) => p.then(() => next(this, ilq), err => Promise.reject(err)), Promise.resolve());
    } else {
      console.error('warning: no inline query handler registered');
      return Promise.resolve();
    }
  }

  _dispatch_chosen_inline_result(update) {
    if (this._chosen_inline_result_handler.length > 0) {
      const cir = new ChosenInlineResult(update.chosen_inline_result);

      return this._chosen_inline_result_handler
        .reduce((p, next) => p.then(() => next(this, cir), err => Promise.reject(err)), Promise.resolve());
    } else {
      console.error('warning: no chosen inline result handler registered');
      return Promise.resolve();
    }
  }

  static _is_command(update) {
    if (update.hasOwnProperty('message') && update.message.hasOwnProperty('text')) {
      if (update.message.hasOwnProperty('entities')) {
        return !!update.message.entities.find(e => e.type === 'bot_command');
      }
    }

    return false;
  }

  static _is_message(update) {
    return update.hasOwnProperty('message');
  }

  static _is_callback_query(update) {
    return update.hasOwnProperty('callback_query');
  }

  static _is_inline_query(update) {
    return update.hasOwnProperty('inline_query');
  }

  static _is_chosen_inline_result(update) {
    return update.hasOwnProperty('chosen_inline_result');
  }

  static _get_message_type(update) {
    const msg = update.message;

    if (msg.hasOwnProperty('text')) {
      return message_types.TEXT;
    } else if (msg.hasOwnProperty('photo')) {
      return message_types.PHOTO;
    } else if (msg.hasOwnProperty('voice')) {
      return message_types.VOICE;
    } else if (msg.hasOwnProperty('location')) {
      return message_types.LOCATION;
    } else if (msg.hasOwnProperty('venue')) {
      return message_types.VENUE;
    } else {
      return message_types.OTHER;
    }
  }

  static _get_cbq_type(update) {
    let data;
    try {
      data = JSON.parse(update.callback_query.data);
    } catch (e) {
      return null;
    }

    return data.t || null;
  }
}
