# Changelog

## [0.14.0](https://github.com/archway-network/relayer_exporter/compare/v0.13.1...v0.14.0) (2024-05-22)


### Features

* use latest release-please ([#39](https://github.com/archway-network/relayer_exporter/issues/39)) ([98a0ef1](https://github.com/archway-network/relayer_exporter/commit/98a0ef1b5d6f3a5f9a7a5d103a49ec78e326c785))

## [0.13.1](https://github.com/archway-network/relayer_exporter/compare/v0.13.0...v0.13.1) (2024-05-22)


### Bug Fixes

* ignore wildcard channels ([#36](https://github.com/archway-network/relayer_exporter/issues/36)) ([f629c41](https://github.com/archway-network/relayer_exporter/commit/f629c41e94ce72e3b9b8f7b752b2a58638de0269))

## [0.13.0](https://github.com/archway-network/relayer_exporter/compare/v0.12.0...v0.13.0) (2024-01-12)


### Features

* add chainName for clientExpiry ([#33](https://github.com/archway-network/relayer_exporter/issues/33)) ([0fea165](https://github.com/archway-network/relayer_exporter/commit/0fea1656a18ef27602938a42419ef69e641482d1))

## [0.12.0](https://github.com/archway-network/relayer_exporter/compare/v0.11.0...v0.12.0) (2023-11-27)


### Features

* Add configurable timeout to rpc endpoints ([#28](https://github.com/archway-network/relayer_exporter/issues/28)) ([1bf89eb](https://github.com/archway-network/relayer_exporter/commit/1bf89eb0309d41d23cc355448f77144d876ec34d))
* Add labels with chain names for cosmos_ibc_stuck_packets metric ([#30](https://github.com/archway-network/relayer_exporter/issues/30)) ([367fdac](https://github.com/archway-network/relayer_exporter/commit/367fdac3f8c0111181be67d6e68324f6dd923464))

## [0.11.0](https://github.com/archway-network/relayer_exporter/compare/v0.10.0...v0.11.0) (2023-11-14)


### Features

* add discord labels to metrics ([#26](https://github.com/archway-network/relayer_exporter/issues/26)) ([1442748](https://github.com/archway-network/relayer_exporter/commit/14427484b9bde6dcfcddcde8268622bf1fc6a443))

## [0.10.0](https://github.com/archway-network/relayer_exporter/compare/v0.9.0...v0.10.0) (2023-10-31)


### Features

* new export for stuck packets ([#24](https://github.com/archway-network/relayer_exporter/issues/24)) ([2bba835](https://github.com/archway-network/relayer_exporter/commit/2bba8359f4488b86230b3c9194fd2cb6011e347e))

## [0.9.0](https://github.com/archway-network/relayer_exporter/compare/v0.8.0...v0.9.0) (2023-10-20)


### Features

* change Juno RPC endpoint ([#22](https://github.com/archway-network/relayer_exporter/issues/22)) ([a630478](https://github.com/archway-network/relayer_exporter/commit/a63047843dbee597f59bb1234b194c3382eb8f6f))
* update nois rpc endpoint ([#21](https://github.com/archway-network/relayer_exporter/issues/21)) ([25bb04f](https://github.com/archway-network/relayer_exporter/commit/25bb04f1f43321792d304b7eef0bf4b79ed411d0))

## [0.8.0](https://github.com/archway-network/relayer_exporter/compare/v0.7.1...v0.8.0) (2023-10-16)


### Features

* Use separate dir for testnets IBC paths ([#19](https://github.com/archway-network/relayer_exporter/issues/19)) ([3ed83c0](https://github.com/archway-network/relayer_exporter/commit/3ed83c03836df28c90889478f068d7be3ae359d5))

## [0.7.1](https://github.com/archway-network/relayer_exporter/compare/v0.7.0...v0.7.1) (2023-10-02)


### Bug Fixes

* Use name instead id for chain name ([#16](https://github.com/archway-network/relayer_exporter/issues/16)) ([ae320e5](https://github.com/archway-network/relayer_exporter/commit/ae320e5a612f14a1ded2f03247a60e9931a40069))

## [0.7.0](https://github.com/archway-network/relayer_exporter/compare/v0.6.0...v0.7.0) (2023-09-26)


### Features

* Add status label for cosmos_ibc_client_expiry metric ([#14](https://github.com/archway-network/relayer_exporter/issues/14)) ([f0dffc0](https://github.com/archway-network/relayer_exporter/commit/f0dffc0e3fd001a107ada8ed4f478ad3d29bf701))

## [0.6.0](https://github.com/archway-network/relayer_exporter/compare/v0.5.0...v0.6.0) (2023-09-13)


### Features

* Add cosmos_wallet_balance metric ([#12](https://github.com/archway-network/relayer_exporter/issues/12)) ([436982c](https://github.com/archway-network/relayer_exporter/commit/436982c5f83a97f73c3e4b6700f97bee68e5c2d6))

## [0.5.0](https://github.com/archway-network/relayer_exporter/compare/v0.4.0...v0.5.0) (2023-09-12)


### Features

* Add authentication support for github client ([#10](https://github.com/archway-network/relayer_exporter/issues/10)) ([825fea4](https://github.com/archway-network/relayer_exporter/commit/825fea4ceb1654b02228f7403724d8db46e9a1a1))

## [0.4.0](https://github.com/archway-network/relayer_exporter/compare/v0.3.0...v0.4.0) (2023-09-07)


### Features

* Gather clients expiry metrics for paths on github ([#8](https://github.com/archway-network/relayer_exporter/issues/8)) ([91fc418](https://github.com/archway-network/relayer_exporter/commit/91fc418769041b6865677f049e48f021738e0329))

## [0.3.0](https://github.com/archway-network/relayer_exporter/compare/v0.2.0...v0.3.0) (2023-05-12)


### Features

* Add support for log levels ([#6](https://github.com/archway-network/relayer_exporter/issues/6)) ([444caba](https://github.com/archway-network/relayer_exporter/commit/444caba6f2526203ea27d543f5123a297e3175a7))

## [0.2.0](https://github.com/archway-network/relayer_exporter/compare/v0.1.1...v0.2.0) (2023-04-28)


### Features

* add cosmos_relayer_configured_chain metric ([#4](https://github.com/archway-network/relayer_exporter/issues/4)) ([c77af43](https://github.com/archway-network/relayer_exporter/commit/c77af4369c3b5911681dfff54e9593ef6fd94548))

## [0.1.1](https://github.com/archway-network/relayer_exporter/compare/v0.1.0...v0.1.1) (2023-04-26)


### Bug Fixes

* do not fail all metrics gathering with single query failure ([#2](https://github.com/archway-network/relayer_exporter/issues/2)) ([8daea53](https://github.com/archway-network/relayer_exporter/commit/8daea535dfff140f607ccdb7dce668c4bfaebc59))
