"use strict";

// datastore purposes:
// 2. autocomplete

// no need:
// 1. multi message commands eg. '/eta' followed by '96049' DON'T EVEN NEED THIS
// history
// favourites

/**
 * Information about a specific bus stop
 * @typedef {object} BusStopInfo
 * @prop {string} id
 * @prop {string} description
 * @prop {string} road
 * @prop {number} latitude
 * @prop {number} longitude
 */

export default class Datastore {
  /**
   * Retrieve information about a bus stop
   * @param {string} id
   * @returns {Promise.<?BusStopInfo>}
   */
  get_bus_stop_info(id) {
    return Promise.resolve();
  }

  /**
   * Get autocompletions for an inline query
   * @param {string} query
   * @returns {Promise.<BusStopInfo[]>}
   */
  get_completions(query) {
    return Promise.resolve();
  }

  /**
   * Get nearby bus stops
   * @param {number} lat
   * @param {number} lon
   * @returns {Promise.<BusStopInfo[]>}
   */
  get_nearby_bus_stops(lat, lon) {
    return Promise.resolve();
  }
}
