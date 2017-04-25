# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [NEXT_RELEASE]
### Fixed
- Fix the deploy archive building with .teresaignore on Windows

## [0.3.2] - 2017-04-12
### Added
- flags `max-cpu` and `max-memory` in `app create` command to define the `default` limits of the application
(the flags `cpu` and `memory` still define `default requested` quota)

## [0.3.1] - 2017-04-10
### Fixed
- panic error when configuring cluster ([#47](https://github.com/luizalabs/teresa-cli/issues/47))

## [0.3.0] - 2017-04-07
### Changed
- `.teresaignore` pattern type from regex to glob

### Added
- Support for custom process types

## [0.2.1] - 2017-03-14
### Added
- read file `.teresaignore` to ignore some files in deploy process

### Changed
- some code refactors
- default client timeout to 30 minutes

## [0.2.0]
### Changed
- big breaking changing version

## [0.1.2] - 2016-08-18
### Fixed
- adding users to teams

## [0.1.1] - 2016-08-12
### Added
- command `add team-user`
- vendoring dependencies

### Changed
- deployment now returning an error message on fail
- default service port changed from 5000 to 80
- deployment timeout increased
- improved deployment info printed out to user
- go-swagger updated to version 5.0.0

## [0.1.0] - 2016-08-03
