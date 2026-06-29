# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
### Changed
### Fixed

## [1.3.0] - 2026-06-11

### Added

- Add support for building and publishing multi-architecture container images for `linux/amd64` and `linux/arm64`.
- Add container image build support for AWS Graviton.
- Add GitHub Actions release workflow using `ko` to publish multi-architecture images to GitHub Container Registry.

### Changed

- Use Go 1.26.2 in the image build workflow to align with Kubernetes 1.36.
- Use Ubuntu 24.04 as the `ko` container base image.
- Use `ko` for local image, Docker load, and release image build targets.
- Remove the old Travis CI and Dockerfile image build paths.
- Switch the default image repository from Docker Hub to GitHub Container Registry.
- Update vulnerable Go dependencies reported by Dependabot.

## [1.2.0] - 2021-06-10

### Changed

- Update for kubernetes 1.20+
- Update Prometheus Go client library
- Bump up Go version

## [1.1.0] - 2019-09-18
### Changed
- Use container started events instead of pod updates
- Update golang docker image to 1.13.0

## [1.0.1] - 2019-01-16
### Changed
- Provide release documentation in README
- Provided a Changelog
### Fixed
- Change the event reason `PreviousPodWasOOMKilled` to the correct `PreviousContainerWasOOMKilled`

## [1.0.0] - 2019-01-11
### Added
- Initial release as Open-Source under the Apache License v2.0

[Unreleased]: https://github.com/xing/kubernetes-oom-event-generator/compare/v1.3.0...HEAD
[1.3.0]: https://github.com/xing/kubernetes-oom-event-generator/compare/v1.2.0...v1.3.0
[1.2.0] https://github.com/xing/kubernetes-oom-event-generator/compare/v1.1.0...v1.2.0 
[1.1.0]: https://github.com/xing/kubernetes-oom-event-generator/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/xing/kubernetes-oom-event-generator/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/xing/kubernetes-oom-event-generator/compare/afe6c88c3a8925c7c72ccecf4f52ff1addbbba2d...v1.0.0
