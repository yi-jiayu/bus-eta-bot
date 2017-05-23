# Change Log

## v3.4.5 - 2017-05-22
### Changed
- Refactored bot implementation for testing

## v3.4.4 - 2017-05-22
### Changed
- The bot will now ignore text messages where the first word is longer than 5 characters
- When an eta slash command is invoked directly, the bot will complain if the provided bus stop code is longer than 5 
characters.
- Changed bot's response when a bus stop code is invalid

## v3.4.3 - 2017-05-22
### Added
- More tests.

## v3.4.2 - 2017-05-22
### Added
- Added Google Analytics for tracking usage statistics.

## v3.3.0 - 2017-05-21
### Added
- You can now send the bot a location and it will return a list of nearby bus stops.

### Changed
- The format of eta messages was changed slightly.

### Removed
- Resend buttons will no longer be included in new messages, and the bot will no longer respond to resend button 
callbacks.
