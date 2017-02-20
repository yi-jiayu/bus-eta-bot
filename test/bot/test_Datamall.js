"use strict";

import { assert } from 'chai';
import Datamall from "../../src/bot/Datamall";

suite('Datamall', function () {
  test('parse api response', function () {
    const api_response = {
      "odata.metadata": "http://datamall2.mytransport.sg/ltaodataservice/$metadata#BusArrival/@Element",
      "BusStopID": "96049",
      "Services": [{
        "ServiceNo": "2",
        "Status": "In Operation",
        "Operator": "GAS",
        "OriginatingID": "99009",
        "TerminatingID": "10589",
        "NextBus": {
          "EstimatedArrival": "2017-02-15T02:45:05+00:00",
          "Latitude": "0",
          "Longitude": "0",
          "VisitNumber": "1",
          "Load": "Seats Available",
          "Feature": "WAB"
        },
        "SubsequentBus": {
          "EstimatedArrival": "2017-02-15T03:00:05+00:00",
          "Latitude": "0",
          "Longitude": "0",
          "VisitNumber": "1",
          "Load": "Seats Available",
          "Feature": "WAB"
        },
        "SubsequentBus3": {
          "EstimatedArrival": "",
          "Latitude": "",
          "Longitude": "",
          "VisitNumber": "",
          "Load": "",
          "Feature": ""
        }
      }, {
        "ServiceNo": "24",
        "Status": "In Operation",
        "Operator": "SBST",
        "OriginatingID": "54009",
        "TerminatingID": "54009",
        "NextBus": {
          "EstimatedArrival": "2017-02-15T02:26:15+00:00",
          "Latitude": "1.3470085",
          "Longitude": "103.96316633333333",
          "VisitNumber": "1",
          "Load": "Seats Available",
          "Feature": "WAB"
        },
        "SubsequentBus": {
          "EstimatedArrival": "2017-02-15T02:35:07+00:00",
          "Latitude": "1.3549335",
          "Longitude": "103.98832433333334",
          "VisitNumber": "1",
          "Load": "Seats Available",
          "Feature": "WAB"
        },
        "SubsequentBus3": {
          "EstimatedArrival": "2017-02-15T02:41:17+00:00",
          "Latitude": "1.361278",
          "Longitude": "103.990838",
          "VisitNumber": "1",
          "Load": "Seats Available",
          "Feature": "WAB"
        }
      }, {
        "ServiceNo": "5",
        "Status": "In Operation",
        "Operator": "SBST",
        "OriginatingID": "77009",
        "TerminatingID": "10009",
        "NextBus": {
          "EstimatedArrival": "2017-02-15T02:34:21+00:00",
          "Latitude": "1.3646705",
          "Longitude": "103.96770783333334",
          "VisitNumber": "1",
          "Load": "Seats Available",
          "Feature": "WAB"
        },
        "SubsequentBus": {
          "EstimatedArrival": "2017-02-15T02:45:51+00:00",
          "Latitude": "0",
          "Longitude": "0",
          "VisitNumber": "1",
          "Load": "Seats Available",
          "Feature": "WAB"
        },
        "SubsequentBus3": {
          "EstimatedArrival": "2017-02-15T03:02:51+00:00",
          "Latitude": "0",
          "Longitude": "0",
          "VisitNumber": "1",
          "Load": "Seats Available",
          "Feature": "WAB"
        }
      }]
    };

    const date = new Date("2017-02-15T02:24:27.706Z");
    const expected = {
      "bus_stop_id": "96049",
      "etas": [{"svc_no": "2", "next": 20, "subsequent": 35, "third": "?"}, {
        "svc_no": "24",
        "next": 1,
        "subsequent": 10,
        "third": 16
      }, {"svc_no": "5", "next": 9, "subsequent": 21, "third": 38}],
      "updated": date
    };

    const actual = Datamall.parse_etas(api_response, date);

    assert.deepEqual(actual, expected);
  });
});
