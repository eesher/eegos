# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
## [0.0.2] - 2020-01-29
### Changed
- change mod name to go pkg
- gate better to be network
### Removed
- old reader
- unpack function

## [0.0.2] - 2020-01-11
### Added
- TcpConn as base tcp connection
### Changed
- move TcpServer listen to Start()
- isolate Data in package
- set TcpClient ticker to time.Timer for reset
- use local log package
- session Write in connection
- move TcpConn counter to TcpClient
- use session message handle
### Deprecated
- old reader in session
### Removed
- old rpc
### Fixed
- use new reader in session
- for loop with full gorountine will stuck in reflect.Method.Call

## [0.0.2] - 2020-01-01
### Changed
- complete rpc example still have bugs
### Added
- TcpClient
- new rpc client

## [0.0.2] - 2019-12-31
### Changed
- keep session simple
- keep session inside package
- process heartbeat in listening server
### Added
- gate handle 
- new package util
- new rpc use gate
- outData channel for write buff
### Deprecated
- old rpc

## [0.0.2] - 2019-12-29
### Changed
- make session as gate session
### Added
- use go mod
- new package gate, log

## [0.0.1] - 2018-06-17
### Changed
- Move netpack func to session.go
- Use append byte instead of make length byte, small memory needed, maybe faster.
- Declare session id uint16 that no need use (if max uint16)
- marshal/unmarshal func to set different proto
### Added
- Add heartbeat in client
- Add timeout when use call
- New data type rpc.heartBeatRet
### Deprecated
- rpc.packRegister
- Session.HandleWriteTest for test, will remove soon
### Fixed
- map[key] = nil cannot delete key, use delete()

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
