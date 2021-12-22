# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
### Changed
### Removed
### Fixed

## [0.12.1] - 2021-12-22
### Changed
 * Bump https://github.com/bitpoke/build to 0.7.1
### Fixed
 * Fix the app version in the published Helm charts

## [0.12.0] - 2021-12-22
### Added
### Changed
 * Minimum required Kubernetes version is 1.19
 * Use `networking.k8s.io/v1` for `Ingress` resources
 * Run WordPress Operator as non-root user
 * Bump https://github.com/bitpoke/build to 0.7.0
### Removed
### Fixed

## [0.11.1] - 2021-11-22
### Changed
 * Change the default image to `docker.io/bitpoke/wordpress-runtime:5.8.2`

## [0.11.0] - 2021-11-15
### Changed
 * Use [Bitpoke Build](https://github.com/bitpoke/build) for building the
   project
### Removed
 * Drop support for Helm v2
### Fixed
