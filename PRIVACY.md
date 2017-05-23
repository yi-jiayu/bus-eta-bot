# Bus Eta Bot Privacy Policy
Last modified: 2017-05-23. Applicable to Bus Eta Bot v3.0.0 and above.

## What data does Bus Eta Bot collect?
Bus Eta Bot collects two types of data: application logs and usage statistics. 

### Application logs
The contents of each update received from the Telegram Bot are logged. This includes, but is not restricted to, message contents, timestamps, user identifiers and public profile information. A detailed description of what these updates contain can be found in the [Telegram Bot API documentation].

### Usage statistics
The type, timestamp and user identifier of each user interaction with the bot are recorded. The type of user interaction corresponds to the action taken on Telegram, for example sending a bot commmand, making an inline query, or pressing an inline keyboard button, while the user identifier is a unique number assigned to each user.

## How is this data collected?
Application logs are written to the application log, while usage statistics are recorded using the [Google Analytics Measurement Protocol].

## What is this data collected for?
Application logs are used to monitor the status of the bot and to facilitate error identification, diagnosis and rectification. Usage statistics help to reveal patterns in user engagement with the bot, such as which features are more or less popular and how they can be improved or removed. It also makes the bot creator happy to see his bot being used more.

## Who will have access to this data, and how is it protected.
Only the bot creator has access to this data, and best practices such as using randomly generated strong passwords and multi-factor authentication are taken to ensure there is no unauthorised access to the Google Cloud Platform project and Google Analytics account containing this data. However the bot runs on Google Cloud Platform and the Measurement Protocol is provided by Google Analytics, and their respective terms of service also apply:
- [Google Cloud Platform TOS]
- [Google Analytics TOS]

## How can I be notified if this privacy policy changes?
This document is an authoritative reference for the Bus Eta Bot privacy policy. Any changes to it are updates to the privacy policy and will be reflected in this repository.

## Who can I contact with questions about this privacy policy?
You may raise an issue in this repository or contact the bot creator at privacy@bus-eta-bot.jiayu.io.

[Telegram Bot API documentation]: https://core.telegram.org/bots/api
[Google Analytics Measurement Protocol]: https://developers.google.com/analytics/devguides/collection/protocol/v1/
[Google Cloud Platform TOS]: https://cloud.google.com/terms/
[Google Analytics TOS]: https://www.google.com/analytics/terms/
