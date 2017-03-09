# Change log

## 2.0.10 -2017-3-9

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
