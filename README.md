[![Build Status](https://semaphoreci.com/api/v1/jiayu/bus-eta-bot/branches/master/shields_badge.svg)](https://semaphoreci.com/jiayu/bus-eta-bot)
[![Coverage Status](https://coveralls.io/repos/github/yi-jiayu/bus-eta-bot/badge.svg?branch=master)](https://coveralls.io/github/yi-jiayu/bus-eta-bot?branch=master)

# Bus Eta Bot
A Telegram bot for checking bus etas in Singapore

## About
This is a rewrite of my [previous bus eta bot](https://github.com/yi-jiayu/bus-eta-bot-sg) for maintainability and new features, while dropping other features which turned out to be less popular than expected. It is currently available on Telegram at [@DevBuildBusEtaBot](https://t.me/DevBuildBusEtaBot) and is running on Google Cloud Platform as a Cloud Function, however it is not yet production-ready and will likely be broken often. Once the important v1 features are implemented in v2 and it is generally stable, it will be promoted to take over [@BusEtaBot](https://t.me/BusEtaBot).

## Features

[Changelog](CHANGELOG.md)

- Request etas for all buses or specific buses at a bus stop by id.
- Search all bus stops by id, description and road name via inline query.
- **(WIP)** Query bus stops by location.

## Architecture
Bus Eta Bot is deployed as a [Cloud Function](https://cloud.google.com/functions/) on [Google Cloud Platform](https://cloud.google.com/) responding to [Telegram Bot API](https://core.telegram.org/bots/api) webhooks.

## [Privacy Policy](PRIVACY.md)
