"use strict";

import request from 'request';

const BUS_ETA_ENDPOINT = 'http://datamall2.mytransport.sg/ltaodataservice/BusArrival';

function log(msg) {
  console.error('datamall: ' + msg);
}

/**
 * A response from the LTA Bus Arrival API
 * @typedef {object} BusEtaResponse
 * @prop {string} Metadata
 * @prop {string} BusStopID
 * @prop {ServiceInfo[]} Services
 */

/**
 * Information about a particular bus service
 * @typedef {object} ServiceInfo
 * @prop {string} ServiceNo
 * @prop {string} Status
 * @prop {string} Operator
 * @prop {string} OriginatingID
 * @prop {string} TerminatingID
 * @prop {ArrivingBusInfo} NextBus
 * @prop {ArrivingBusInfo} SubsequentBus
 * @prop {ArrivingBusInfo} SubsequentBus3
 */

/**
 * Information about an incoming bus
 * @typedef {object} ArrivingBusInfo
 * @prop {string} EstimatedArrival
 * @prop {string} Latitude
 * @prop {string} Longitude
 * @prop {string} VisitNumber
 * @prop {string} Load
 * @prop {string} Feature
 */

/**
 * Make a request to the Datamall bus eta endpoint for bus_stop
 * @param bus_stop
 * @return {Promise.<BusEtaResponse>}
 * @private
 */
function _get_etas_from_api(bus_stop) {
  if (!process.env.DATAMALL_ACCOUNT_KEY || !process.env.DATAMALL_USER_ID) {
    console.error('warning: no datamall credentials');
  }

  const options = {
    url: BUS_ETA_ENDPOINT,
    headers: {
      AccountKey: process.env.DATAMALL_ACCOUNT_KEY,
      UniqueUserId: process.env.DATAMALL_USER_ID,
      accept: 'application/json'
    },
    qs: {
      BusStopID: bus_stop
    }
  };

  return new Promise((resolve, reject) => {
    request(options, (err, res, body) => {
      if (err) {
        log(err);
        reject(err);
      } else if (res.statusCode >= 400) {
        log(res.statusCode);
        log(body);
        reject({res, body});
      } else {
        resolve(JSON.parse(body));
      }
    });
  });

}

/**
 * Get etas for bus_stop
 * @param {string} bus_stop
 * @return {Promise.<ParsedEtas>}
 */
export function get_etas(bus_stop) {
  return _get_etas_from_api(bus_stop);
}
