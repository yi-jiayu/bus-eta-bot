'use strict';

import request from 'request';

import Analytics from './Analytics';

export default class KeenAnalytics extends Analytics {
  constructor() {
    super();

    const [project_id, write_key] = [process.env.KEEN_PROJECT_ID, process.env.KEEN_WRITE_KEY];

    if (!project_id || !write_key) {
      console.error('warning: KEEN_PROJECT_ID and/or KEEN_WRITE_KEY not set');
    }

    this.project_id = project_id;
    this.write_key = write_key;
  }

  _record_single_event(collection, data) {
    return new Promise((resolve, reject) => request.post({
      uri: `https://api.keen.io/3.0/projects/${this.project_id}/events/${collection}`,
      headers: {
        'Authorization': this.write_key
      },
      body: data,
      json: true
    }, (err, res, body) => {
      if (err) reject(err);
      else if (res.status >= 400) {
        const err = new Error('status code error: ' + res.statusCode);
        err.res = res;
        err.body = body;

        console.error(err);
        reject(err);
      }
      else resolve(body);
    }));
  }

  /**
   * Record an interaction with the bot
   * @param {string} event
   * @param {object} details
   */
  log_event(event, details) {
    const data = details;
    data.event = event;

    return this._record_single_event('events', data);
  }
}
