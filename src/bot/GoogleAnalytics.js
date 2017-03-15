'use strict';

import request from 'request';

import Analytics, { event_types } from './Analytics';

const ga_endpoint = 'https://www.google-analytics.com/collect';
const tid = process.env.GA_TID;
const version = require('../../package.json').version;

if (!tid) {
  console.error('warning: GA_TID not set');
}

export default class GoogleAnalytics extends Analytics {
  log_event(event, details) {
    if (!tid) {
      console.error('warning: GA_TID not set');
      return Promise.resolve();
    }

    let ec, ea;
    switch (event) {
      case event_types.help_command:
        ec = 'bot_command';
        ea = 'help';
        break;
      case event_types.about_command:
        ec = 'bot_command';
        ea = 'about';
        break;
      case event_types.start_command:
        ec = 'bot_command';
        ea = 'start';
        break;
      case event_types.version_command:
        ec = 'bot_command';
        ea = 'version';
        break;
      case event_types.privacy_command:
        ec = 'bot_command';
        ea = 'privacy';
        break;
      case event_types.eta_command:
        ec = 'bot_command';
        ea = 'eta';
        break;
      case event_types.eta_text_message:
        ec = 'text_message';
        ea = 'eta';
        break;
      case event_types.refresh_callback:
        ec = 'callback_query';
        ea = 'refresh';
        break;
      case event_types.resend_callback:
        ec = 'callback_query';
        ea = 'resend';
        break;
      case event_types.eta_demo_callback:
        ec = 'callback_query';
        ea = 'eta_demo';
        break;
      case event_types.inline_query:
        ec = 'inline_query';
        ea = 'inline_query';
        break;
      case event_types.chosen_inline_result:
        ec = 'chosen_inline_result';
        ea = 'chosen_inline_result';
        break;
      default:
        return Promise.reject('invalid event');
    }

    return new Promise((resolve, reject) => {
      request({
        uri: ga_endpoint,
        method: 'POST',
        form: {
          v: 1,
          tid,
          uid: details.user_id,
          t: 'event',
          ec,
          ea,
        },
      }, (err, resp, body) => {
        if (err) {
          reject(err);
        } else if (resp.statusCode < 200 || resp.statusCode >= 300) {
          reject(body);
        } else {
          resolve(body);
        }
      });
    });
  }
}
