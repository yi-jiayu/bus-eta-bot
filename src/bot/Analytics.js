"use strict";

/*
 Things to track:

 1. user
 2. action
 3. timing

 that's all

 no need to track:
 1. bus stop
 2. bus service
 */

export const event_types = {
  eta_command: 'eta_command',
  about_command: 'about_command',
  start_command: 'start_command',
  help_command: 'help_command',
  version_command: 'version_command',
  eta_text_message: 'eta_text_message',
  refresh_callback: 'refresh_callback',
  resend_callback: 'resend_callback',
  eta_demo_callback: 'eta_demo_callback',
  inline_query: 'inline_query',
  chosen_inline_result: 'chosen_inline_result',
};

export default class Analytics {
  log_event(event, details) {
    return Promise.resolve();
  }
}
