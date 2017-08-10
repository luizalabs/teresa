# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [0.5.0] - Release Date
### Changed
- Use gRPC instead of go-swagger
- Changed project structure to be more k8s like
- Standardize error messages format and server logs
- 'app create' command doesn't create a welcome deploy anymore
- [client] Refactored config and bash completion
- [server] Delete build pod after deploy on success

### Added
- Unit tests
- Travis CI
- 'set-password' command
- Builtin TLS (thanks gRPC)
- Versioning by git tag
- [server] 'create-super-user' command
- [server] added so called 'release' command, which allows developers to run a
  command right after the build is finished
- [server] Healthcheck
- [server] Support for minio storage
- [server] Added deploy drain timeout in teresa.yaml, which makes the POD sleep
  for a configurable amount of time before receiving SIGTERM

### Fixed
- [server] Don't return stale app data to the user
- [server] Every deploy has at least one replica
- [server] 'app logs' command timeouts

## [0.3.2] - 2017-05-09
### Fixed
- Finish the merge with `teresa-cli` by removing all references to the old repo
- Multi team tokens. Before this fix a user would get access only to apps
  from one of its teams.

## [0.3.1] - 2017-04-24
### Fixed
- Fix deploy timeouts by sending a few bytes at a constant time interval to the
  network connection in use (as required by the idle timeout of the aws classic
  elb for instance)

### Changed
- Merge `teresa-cli` and `teresa-api`

## [0.3.0] - 2017-04-07
### Added
- Support for non-web process types

### Fixed
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
