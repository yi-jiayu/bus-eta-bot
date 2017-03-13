'use strict';

/**
 * @typedef {object} ParsedEtas
 * @prop {Date} updated
 * @prop {string} bus_stop_id
 * @prop {Array} etas
 */

export default class EtaProvider {
  /**
   * Get bus arrival information for a bus stop
   * @param {string} bus_stop_id
   * @returns {Promise.<ParsedEtas>}
   */
  get_etas(bus_stop_id) {
    return Promise.resolve({
      bus_stop_id,
      etas: [],
      updated: new Date()
    });
  }
}
