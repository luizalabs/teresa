# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

## [NEXT_RELEASE]
### Changed
- Change Method `CreateSecret` to `CreateOrUpdateSecret` on k8s interfaces

### Added
- Support Nginx as sidecar
- support for internal apps

## [0.16.0] - 2018-03-07
### Added
- `app logs` support to filter by pod name
- `app logs` support to print the logs of the previous pod instance
- [server] `replace-storage-secret` admin command
- CronJob experimental support
- CONTRIBUTING.md and related FAQ entry

### Changed
- Better error message for invalid app name error
- Better error message for invalid env var name error
- Refactor specs to be more in line with k8s concepts
- The slugrunner doesn't mount the storage keys anymore. An init container is
  responsible for downloading the slug
- Bump the default slugrunner version to v3.0.1

## [0.15.0] - 2018-02-14
### Changed
- [server] using multistage build on teresa dockerfile
- [server] update default slugbuilder version to v3.2.0
- Upgrade golang version to 1.9

### Added
- `exec` command
- [helm] Add HealthChecks
- [server] graceful shutdown
- [client] support to deploy remote (http and https) and local files

## [0.14.0] - 2018-02-06
### Changed
- [server] decouple k8s client interface from the domain ones
- [client] rewrite client tar pkg and change the deploy cmd accordingly

### Fixed
- Pod list on `app info` command for apps without HPA
- [server] make the build process stop on client cancellation

### Added
- [server] default deploy lifecycle with 10s drain timeout
- [server] configurable env (`TERESA_DB_SHOW_LOGS`) to show (or not) database logs (default `false`)

## [0.13.0] - 2018-01-22
### Fixed
- [client] On deploys create the tar file before connecting to the server and
  remove it before exiting
- [server] null pointer exception on deploy

### Added
- app Start/Stop commands
- `app delete-pods` command
- set-password now support the --user flag to set password of another user(needs admin)

## [0.12.0] - 2018-01-12
### Changed
- Back to Godep for dependencies management
- Update `client-go` lib to version 4.0
- [server] Don't use the default service account on builds and deploys

### Added
- alias `app log` to `app logs`
- `app logs` now  support shorthand `-f` to `--follow` and `-n` to `--lines`

### Fixed
- Env vars being set or unset for the app on deploy update errors
- Typo on server pkg
- [server] Return all deploy errors to the client
- [client] teresaignore on Windows

## [0.11.0] - 2017-12-11
### Added
- [helm] support rbac
- Support ingress on app expose
- app create now support --vhost to define a virtual host to app
- Helm support ingress and service type config

### Fixed
- Don't return error on `app info` command if the app doesn't have HPA

## [0.10.0] - 2017-11-08
### Added
- Pod readiness to the `app info` output
- [helm] minio as dependency
- Team rename command

### Changed
- Using `dep` instead of `Godep` for dependencies management
- deploy list now print revision in reverse order and remove current column

### Fixed
- `app info` now counts only pods with defined state

## [0.9.0] - 2017-10-26
### Fixed
- login now use the --cluster flag to save the token to config file
- Don't return error on `app info` command if the app doesn't have HPA

### Changed
- Upgrade `slugbuilder` version to `v3.0.0`
- Timeout of `PodRun` process (deploy and release) is now configurable

### Added
- Allows developers to set the JWT auth token expiration period

## [0.8.0] - 2017-09-18
### Added
- Command 'deploy list'
- Command 'team remove-user'
- Command 'deploy rollback'
- Command 'app delete'

### Changed
- Deploys are now performed using the 'deploy create' command
- app info now print env vars sorted

### Fixed
- App info don't print pod without status

## [0.7.0] - 2017-08-29
### Added
- [server] --debug flag. For now only print the stack trace on panic/recover.
- [client] --cluster flag. To use a cluster different of current-cluster.

### Changed
- [client] Commands 'env-set' and 'env-unset' show the current cluster.

### Fixed
- [server] Restart count on 'app info' output
- [client] Validate 'deploy' command parameters
- [client] Hanging deploys when the app dir doesn't exist

## [0.6.0] - 2017-08-23
### Added
- [client] App name length validation
- autoscale command
- [server] Support for Teresa yaml per process type

### Changed
- [server] Specific CPU and Memory limits for both deploy and release pods
- [client] Change default `max-cpu` to `200m` (instead of `500m`) in command `create app`
- [server] Doesn't log request content on error middleware if the route is `Login`
- 'app info' command shows the pods age and restart count

## [0.5.0] - 2017-08-15
### Changed
- Use gRPC instead of go-swagger
- Changed project structure to be more k8s like
- Standardize error messages format and server logs
- 'app create' command doesn't create a welcome deploy anymore
- [client] Refactored config and bash completion
- [client] Flag --admin removed, admin users are only created with the
  'create-super-user' command
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
