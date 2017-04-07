# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [0.3.0] - 2017-04-07
### Added
- Support for non-web process types

#### Fixed
- Get current namespace name from environment variable instead of a constant

### Changed
- Location of slugbuilder and slugrunner images
- Read keys from k8s secrets
- Increase deploy timeouts and make them configurable

## [0.2.2] - 2017-03-16
### Changed
- Upgrade slugbuilder and slugrunner

## [0.2.1] - 2017-03-14
### Added
- use `TERESADB_DATABASE` variable to define location of `teresa.sqlite`

### Changed
- update README to use an api to generate password instead of a python script
- remove dead code
- remove users/me endpoint

## [0.2.0]
### Changed
- big breaking changing version

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
