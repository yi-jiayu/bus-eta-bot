"use strict";

import Geo from 'geo-nearby';

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
 */

export default class Datastore {
  /**
   * Create and populate a new datastore
   * @param {{i: string, d: string, r: string}[]}info
   * @param {{i: string, g: number}[]} locations
   */
  constructor(info, locations) {
    this.info = info;
    this.locations = locations;
  }

  /**
   * Retrieve information about a bus stop
   * @param {string} code
   * @returns {?BusStopInfo}
   */
  get_bus_stop_info(code) {
    const bs = this.info[code];

    if (bs) {
      return {
        id: code,
        description: bs.d,
        road: bs.r,
        latitude: bs.lat,
        longitude: bs.lon
      }
    } else {
      // throw or return null?
      return null;
    }
  }

  /**
   * Get autocompletions for an inline query
   * @param {string} query
   * @returns {BusStopInfo[]}
   */
  get_completions(query) {
    query = query.toUpperCase().trim();
    const matches = [];

    for (const bs in this.info) {
      //noinspection JSUnfilteredForInLoop
      if (bs.toUpperCase().indexOf(query) !== -1 || this.info[bs].d.toUpperCase().indexOf(query) !== -1) {
        //noinspection JSUnfilteredForInLoop
        matches.push({
          id: bs,
          description: this.info[bs].d,
          road: this.info[bs].r,
          latitude: this.info[bs].lat,
          longitude: this.info[bs].lon
        });
      }
    }

    return matches;
  }

  /**
   * Get nearby bus stops
   * @param {number} lat
   * @param {number} lon
   * @returns {BusStopInfo[]}
   */
  get_nearby_bus_stops(lat, lon) {
    // limit of 50 is is the max number of results for an inline query
    // note: no need for sorted: true because bsearch is actually slower with our dataset
    const geo = new Geo(this.locations, {limit: 50});

    // try arbitrary 100m radius for now
    const nearby = geo.nearBy(lat, lon, 100);

    return nearby.map(nb => {
      const bs = this.info[nb.i];
      return {
        id: nb.i,
        description: bs.d,
        road: bs.r,
        latitude: bs.lat,
        longitude: bs.lon
      };
    });
  }
}
