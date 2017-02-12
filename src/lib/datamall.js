"use strict";

import debug from 'debug';
import request from 'request';

const log = debug('datamall');
const BUS_ETA_ENDPOINT = 'http://datamall2.mytransport.sg/ltaodataservice/BusArrival';
const MILLISECONDS_IN_A_MINUTE = 60 * 1000;

const account_key = process.env.DATAMALL_ACCOUNT_KEY;
const user_id = process.env.DATAMALL_USER_ID;

if (!account_key || !user_id) {
  console.error('warning: no datamall credentials');
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

export class ParsedEtas {
  /**
   * Parse a response from the Datamall bus eta endpoint
   * @param {BusEtaResponse} eta_response
   */
  constructor(eta_response) {
    this.updated = new Date();
    this.bus_stop = eta_response.BusStopID;
    this.etas = [];

    const services = eta_response.Services;
    for (const service of services) {
      // if the service is in operation, we use a question mark to signify that etas may be unknown
      // if the service is not in operation, we use a dash to signify that there is no incoming bus
      const placeholder = service.Status === 'Not In Operation' ? '-' : '?';

      const svc_no = service.ServiceNo;
      const next = service.NextBus.EstimatedArrival !== ''
        ? Math.floor((new Date(service.NextBus.EstimatedArrival) - this.updated) / MILLISECONDS_IN_A_MINUTE)
        : placeholder;
      const subsequent = service.SubsequentBus.EstimatedArrival != ''
        ? Math.floor((new Date(service.SubsequentBus.EstimatedArrival) - this.updated) / MILLISECONDS_IN_A_MINUTE)
        : placeholder;
      const third = service.SubsequentBus3.EstimatedArrival != ''
        ? Math.floor((new Date(service.SubsequentBus3.EstimatedArrival) - this.updated) / MILLISECONDS_IN_A_MINUTE)
        : placeholder;
      this.etas.push({svc_no, next, subsequent, third});
    }
  }
}

/**
 * Make a request to the Datamall bus eta endpoint for bus_stop
 * @param bus_stop
 * @return {Promise.<BusEtaResponse>}
 * @private
 */
function _get_etas_from_api(bus_stop) {
  const options = {
    url: BUS_ETA_ENDPOINT,
    headers: {
      AccountKey: account_key,
      UniqueUserId: user_id,
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
  return _get_etas_from_api(bus_stop)
    .then(etas => new ParsedEtas(etas));
}
