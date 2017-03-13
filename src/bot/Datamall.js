'use strict';

import { get_etas } from '../lib/datamall';

import EtaProvider from './EtaProvider';

const MILLISECONDS_IN_A_MINUTE = 60 * 1000;

export default class Datamall extends EtaProvider {
  /**
   *
   * @param eta_response
   * @param {Date} [date]
   * @return {ParsedEtas}
   */
  static parse_etas(eta_response, date) {
    const updated = date || new Date();
    const bus_stop_id = eta_response.BusStopID;
    const etas = [];

    const services = eta_response.Services;
    for (const service of services) {
      // if the service is in operation, we use a question mark to signify that etas may be unknown
      // if the service is not in operation, we use a dash to signify that there is no incoming bus
      const placeholder = service.Status === 'Not In Operation' ? '-' : '?';

      const svc_no = service.ServiceNo;
      const next = service.NextBus.EstimatedArrival !== ''
        ? Math.floor((new Date(service.NextBus.EstimatedArrival) - updated) / MILLISECONDS_IN_A_MINUTE)
        : placeholder;
      const subsequent = service.SubsequentBus.EstimatedArrival !== ''
        ? Math.floor((new Date(service.SubsequentBus.EstimatedArrival) - updated) / MILLISECONDS_IN_A_MINUTE)
        : placeholder;
      const third = service.SubsequentBus3.EstimatedArrival !== ''
        ? Math.floor((new Date(service.SubsequentBus3.EstimatedArrival) - updated) / MILLISECONDS_IN_A_MINUTE)
        : placeholder;
      etas.push({svc_no, next, subsequent, third});
    }

    return {
      bus_stop_id,
      etas,
      updated
    };
  }

  get_etas(bus_stop_id) {
    return get_etas(bus_stop_id)
      .then(Datamall.parse_etas);
  }
}
