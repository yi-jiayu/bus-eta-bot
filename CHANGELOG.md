# Change log

## 2.2.0 - 2017-3-23

### Removed

- Bus Eta Bot does not request inline location anymore, and empty inline queries will not return nearby bus stops. This 
feature will be replaced later on by directly sending the bot a location.

## 2.1.4 - 2017-3-17

### Fixed

- Made sure not to cache results for empty inline queries

## 2.1.3 - 2017-3-13

### Fixed

- Made bot backwards compatible with older callback query which used different callback_data formats

## 2.1.2 - 2017-3-13

### Fixed

- Fixed broken resend button on eta results sent by the bot

## 2.1.1 - 2017-3-13

### Added

- When the bot is unable to find etas for a bus stop, it will mention what bus code it tried to look up.
- The bot will now mention if there are no bus services serving a bus stop, even if it has information about it. (This 
is the case for bus stop codes such as "TPE" or "TENN9".)

### Fixed

- Bot will not respond with an error message when sent an inline query result from itself

## 2.1.0-rc1 - 2017-3-9

### Added

- Added /privacy command which links to the privacy policy.

## 2.0.10 - 2017-3-9

### Added

- Bot will respond to /feedback commands with a placeholder message while the feedback command is still being 
implemented.
- Added privacy policy [here](PRIVACY.md) with regards to the addition of analytics code.

### Under the hood

- Added analytics code to track anonymous usage statics.

## 2.0.9 - 2017-3-7

### Added

- Eta results sent directly by the bot will have a resend button to have the bot send that eta query as a new message.

## 2.0.8 - 2017-2-28

### Added

- Bot with respond with /help commands with a link to a helpful article.

## 2.0.7 - 2017-2-21

### Changed

- Empty inline queries sent with location will now return nearby bus stops sorted by closest first

## 2.0.6 - 2017-2-21
## 2.0.5 - 2017-2-21

### Fixed

- Fixed bug causing bot to never send nearby bus stops

## 2.0.4 - 2017-2-21

### Added

- Bot will respond to empty inline queries with bus stops within a 1 km radius.

## 2.0.3 - 2017-2-21

### Added

- Bot will respond with a helpful (hopefully) welcome message after receiving /start.

## 2.0.2 - 2017-2-20

### Fixed

- Fixed bot not responding to callback queries

### Added

- Remove progress bar on telegram clients after handling callback queries
