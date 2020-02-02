# Bus Eta Bot
[![Weekly Users](https://us-central1-bus-eta-bot.cloudfunctions.net/bus-eta-bot-weekly-users)](https://us-central1-bus-eta-bot.cloudfunctions.net/bus-eta-bot-weekly-users)
[![Build Status](https://travis-ci.org/yi-jiayu/bus-eta-bot.svg?branch=master)](https://travis-ci.org/yi-jiayu/bus-eta-bot)
[![CircleCI](https://circleci.com/gh/yi-jiayu/bus-eta-bot.svg?style=svg)](https://circleci.com/gh/yi-jiayu/bus-eta-bot)
[![codecov](https://codecov.io/gh/yi-jiayu/bus-eta-bot/branch/master/graph/badge.svg)](https://codecov.io/gh/yi-jiayu/bus-eta-bot)
[![Go Report Card](https://goreportcard.com/badge/github.com/yi-jiayu/bus-eta-bot)](https://goreportcard.com/report/github.com/yi-jiayu/bus-eta-bot)

A practical Telegram bot for getting bus etas in Singapore supporting refreshing, bus stop search and showing nearby bus stops.

## Getting started
Telegram: [@BusEtaBot](https://t.me/BusEtaBot)

## Features
### Get etas by bus stop code
![Etas by bus stop code](screenshots/eta-query.png)

### Get etas by bus stop code and service
![Etas by bus stop code and service](screenshots/eta-query-filtered.png)

### Send a location to get nearby bus stops
![Looking up nearby bus stops](screenshots/nearby-bus-stops.png)

### Search bus stops by bus stop code, description or road name
![Searching bus stops](screenshots/search-bus-stops.png)

### Get nearby bus stops
![Location-based inline query](screenshots/nearby-bus-stops-inline.png)

### Send etas as inline messages
![Inline message](screenshots/inline-message.png)

## Updating bus stop data

From the `scripts` directory:

```
# remove datamall.sqlite if it exists
# rm datamall.sqlite
./create_db.sh
./create_bus_stops_json.py > ../data/bus_stops.json
```

