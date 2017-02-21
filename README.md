[![CircleCI](https://circleci.com/gh/yi-jiayu/bus-eta-bot.svg?style=shield)](https://circleci.com/gh/yi-jiayu/bus-eta-bot)
[![codecov](https://codecov.io/gh/yi-jiayu/bus-eta-bot/branch/master/graph/badge.svg)](https://codecov.io/gh/yi-jiayu/bus-eta-bot)

# Bus Eta Bot
A Telegram bot for checking bus etas in Singapore

## About
This is a rewrite of my [previous bus eta bot](https://github.com/yi-jiayu/bus-eta-bot-sg) for maintainability and new features, while dropping other features which turned out to be less popular than expected. It is currently available on Telegram at [@DevBuildBusEtaBot](https://t.me/DevBuildBusEtaBot) and is running on Google Cloud Platform as a Cloud Function, however it is not yet production-ready and will likely be broken often. Once the important v1 features are implemented in v2 and it is generally stable, it will be promoted to take over [@BusEtaBot](https://t.me/BusEtaBot).

## Features

[Changelog](CHANGELOG.md)

**Core**
- **(WIP)** Request etas for all buses or specific buses at a bus stop by id.
- **(WIP)** Search all bus stops by id, description and road name via inline query.
- **(WIP)** Query bus stops by location. Default inline query results will also show nearby bus stops.

**Gimmicks**
- **(NOT IMPLEMENTED)** Send the bot a bus stop id in a voice message.
- **(NOT IMPLEMENTED)** Send the bot a photo of the bus stop info plate to get etas.

**Dropped** (No one used these)
- Eta query history
- Favourite eta queries
- 'Done' button on eta messages

## Architecture
Bus Eta Bot v1 was deployed as an AWS Lambda function and used DynamoDB for persisting user state such as whether the user was in the middle of a /eta command, eta histories and saved queries. V2 avoids having to maintain state by using Telegram message replies to handle user interaction and dropping the history and favourites features. However, v2 does have to store bus stop information. While it is possible to do so in an external database, the number of bus stops (~5300) was low enough to just load a local JSON file with the bus stop details. Further performance monitoring is required to find out if this solution is satisfactory or a RDBMS- or Redis-based would be better.
