"use strict";

const request = require('request');

const get_updates = require('../transpiled/lib/telegram').get_updates;

const bus_eta_bot = require('../transpiled/bot/bus-eta-bot').default;
const bot_token = process.env.TELEGRAM_BOT_TOKEN;

let polling = true;

const bot = bus_eta_bot;

get_updates()
  .then(updates => {
    console.log(updates);
    for (const update of updates.result) {
      console.log(update);
      bot.handle(update)
        .then(console.log)
        .catch(console.error);
    }
  });

