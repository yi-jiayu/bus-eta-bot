"use strict";

import Geo from 'geo-nearby';

import Datastore from "./Datastore";

export default class LocalDatastore extends Datastore {
  /**
   * Creates a datastore locally
   * @param {object} info - object containing bus stop information keyed by bus stop id
   * @param {object[]} locations - array containing bus stop ids and locations
   */
  constructor(info, locations) {
    super();

    this.info = info;
    this.locations = locations;
  }

  get_bus_stop_info(id) {
    return Promise.resolve()
      .then(() => {
        const bs = this.info[id];

        if (bs) {
          return {
            id: id,
            description: bs.d,
            road: bs.r,
            latitude: bs.lat,
            longitude: bs.lon
          };
        } else {
          return null;
        }
      });
  }

  get_completions(query) {
    return Promise.resolve()
      .then(() => {
        // normalise case
        query = query.toUpperCase().trim();
        const matches = [];

        for (const id in this.info) {
          if (matches.length >= 50) {
            break;
          }

          // if the bus stop info object was parsed from json we should have no additional properties
          //noinspection JSUnfilteredForInLoop
          const bs = this.info[id];

          //noinspection JSUnfilteredForInLoop
          if (id.startsWith(query) ||
            bs.d.toUpperCase().indexOf(query) !== -1 ||
            bs.r.toUpperCase().indexOf(query) !== -1) {
            //noinspection JSUnfilteredForInLoop
            matches.push({
              id: id,
              description: bs.d,
              road: bs.r,
              latitude: bs.lat,
              longitude: bs.lon
            });
          }
        }

        return matches;
      });
  }

  get_nearby_bus_stops(lat, lon) {
    return Promise.resolve()
      .then(() => {
        // limit of 50 is is the max number of results for an inline query
        // no need for sorted: true because bsearch seems to be slower with our dataset
        const geo = new Geo(this.locations, {limit: 50});

        // arbitrary 100m radius for now
        // what if there are no bus stops within 100m?
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
      });
  }
}
