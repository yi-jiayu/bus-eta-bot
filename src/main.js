"use strict";

const env = require('../.env.json');
Object.keys(env).map(k => process.env[k] = env[k]);

const bus_eta_bot = require('./bot/bus-eta-bot').default;

const platform = process.env.PLATFORM;

if (platform === 'GCLOUD_FUNCTIONS') {
  exports.main = function (req, res) {
    const update = req.body;

    console.log(JSON.stringify(update));

    bus_eta_bot.handle(update)
      .then(result => {
        console.log('success! msg: ' + JSON.stringify(result));
        return res.status(200).end();
      })
      .catch(err => {
        console.error('error! msg: ' + JSON.stringify(err));
        return res.status(200).end();
      });
  }
} else if (platform === 'AWS_LAMBDA') {
  exports.main = function (event, context, callback) {
    const update = event;

    console.log(JSON.stringify(update));
    bus_eta_bot.handle(update)
      .then(result => {
        console.log('success! msg: ' + JSON.stringify(result));
        callback(null);
      })
      .catch(err => {
        console.error('error! msg: ' + JSON.stringify(err));
        callback(null);
      });
  }
}

