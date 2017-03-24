[![Build Status](https://semaphoreci.com/api/v1/jiayu/bus-eta-bot/branches/master/shields_badge.svg)](https://semaphoreci.com/jiayu/bus-eta-bot)
[![Coverage Status](https://coveralls.io/repos/github/yi-jiayu/bus-eta-bot/badge.svg?branch=master)](https://coveralls.io/github/yi-jiayu/bus-eta-bot?branch=master)

# Bus Eta Bot
A Telegram bot for checking bus etas in Singapore

### Check etas
![Checking etas](eta-query.gif)

### Search bus stops
![Searching bus stops](inline-query.gif)

## Getting started

Contact [@BusEtaBot](https://t.me/BusEtaBot/) on Telegram now!

## Features

[Changelog](CHANGELOG.md)

- Request etas for all buses or specific buses at a bus stop by id.
- Search all bus stops by id, description and road name via inline query.
- **(WIP)** Query bus stops by location.

## Architecture
Bus Eta Bot is deployed as a [Cloud Function](https://cloud.google.com/functions/) on [Google Cloud Platform](https://cloud.google.com/) responding to [Telegram Bot API](https://core.telegram.org/bots/api) webhooks.

## [Privacy Policy](PRIVACY.md)
