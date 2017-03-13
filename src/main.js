'use strict';

import 'source-map-support/register';
import './loadenv';

import LocalDatastore from './bot/LocalDatastore';
import Datamall from './bot/Datamall';
import BusEtaBot from './bot/BusEtaBot';
import KeenAnalytics from './bot/KeenAnalytics';

const datastore = new LocalDatastore(
  require('./data/bus-stop-info.json'),
  require('./data/location-data.json'));
const datamall = new Datamall();
const analytics = new KeenAnalytics();

const bus_eta_bot = new BusEtaBot({
  datastore,
  eta_provider: datamall,
  analytics
});

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
  };
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
  };
}

