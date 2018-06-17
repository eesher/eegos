# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
## [0.0.1] - 2018-06-17
### Changed
- Move netpack func to session.go
### Added
- Add heartbeat in client
- Add timeout when use call
- New data type rpc.heartBeatRet
### Deprecated
- rpc.packRegister

## [0.0.1] - 2018-06-11
### Changed
- Make rpc as one way message flow, all behavior begin with client.
### Added
- Add data type in netpack

## [0.0.1] - 2018-05-26
### Added
- Complete a basic rpc lib

## [0.0.0] - 2018-05-09
### Added
- All ideas base on High-availability cluster, so let's start with rpc
