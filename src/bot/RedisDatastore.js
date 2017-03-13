'use strict';

import debug from 'debug';
import * as redis from 'redis';

import Datastore from './Datastore';

const log = debug('redis-datastore');

export default class RedisDatastore extends Datastore {
  constructor(options) {
    super();

    this.client = redis.createClient(options);
  }

  /**
   * Get autocompletions for an inline query
   * @param {string} query
   */
  get_completions(query) {
    query = query.toUpperCase();

    const no_completions = new Promise((resolve, reject) => {
      this.client.zrangebylex('bus_stop_nos', `[${query}`, `[${query}{`, (err, res) => {
        if (err) reject(err);
        else resolve(res);
      });
    });

    const desc_completions = new Promise((resolve, reject) => {
      this.client.zrangebylex('bus_stop_descs', `[${query}`, `[${query}{`, (err, res) => {
        if (err) reject(err);
        else resolve(res);
      });
    });

    return Promise.all([no_completions, desc_completions])
      .then(([nos, descs]) => {
        const completions = [];

        // return desc matches before number matches
        for (const desc of descs) {
          const [_, n, d] = desc.split('|');
          completions.push({code: n, description: d});
        }

        for (const no of nos) {
          const [n, d] = no.split('|');
          completions.push({code: n, description: d});
        }

        return completions;
      });
  }

  get_nearby_bus_stops(location) {
    return super.get_nearby_bus_stops(location);
  }
}
