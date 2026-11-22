# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-11-22

The first stable release, with go-matter v1.0.0: a Matter hub which adds devices over Bluetooth or the network, controls lights, and shares its devices with Google Home, Amazon Alexa and other apps (multi-admin, verified with Google Home). Distributed as .deb packages for amd64 and arm64.

### Added

- **Sharing a device with another app or hub** (Matter multi-admin). **Share with another app…** on a device's card opens a pairing window on the device for 10 minutes, with a QR code and an 11-digit setup code, for Google Home, Amazon Alexa, Apple Home or SmartThings to add the device as well; the hub stays in control. `POST` and `DELETE /api/v1/nodes/{id}/share`. The codes are neither logged nor kept.
- **Adding a device another app shares**, with that app's setup code, over the network.
- **The stages of Add device.** The dialog shows where a commissioning is (discovering, connecting, PASE, attestation, credentials, network, operational discovery, CASE), and where one failed.
- **Why a device could not join Wi-Fi.** A commissioning which fails because the device could not join the network ends with `wifi_failed`, and the reason the device gave, such as a wrong password or a network it does not find, is shown in plain words.
- **An icon, and a launcher in the desktop's applications.** The package installs `homewire.desktop` and the icon in every size; the launcher runs `homewire open -notify`, which tells a failure with a desktop notification. The web UI's header and the browser tab show the icon.
- Screenshots of the web UI in the README.
- The README tells what to do with a device which does not get back on the Wi-Fi network after it drops.

### Changed

- Uses go-matter v1.0.0, with the security updates of go-ble v0.9.3 and go-mdns v1.0.1. The requests to a node take the context of the HTTP request.
- After a device is removed, the hub says that a device another app still has is added again with that app's code, without a reset.
- The snapshot version of a package build is a semver pre-release (`0.9.1-dev.<commit>`); the .deb still sorts it before the release.

### Fixed

- A brightness or color temperature slider no longer jumps back to a value in between while a slow light gets to the new one.
- `homewire status`, `open` and `token` find the data directory of the running service before an idle one of the user's own, and tell how to be given access to the service's (`sudo adduser $USER homewire`).
- Devices which need longer to be found, or which advertise a wrong address right after joining the network, are commissioned (go-matter v0.9.6 and v0.9.7).

## [0.9.0] - 2026-11-15

The first release: a Matter hub with a fabric of its own, a web UI and an HTTP API.

### Added

- A daemon and an embedded Svelte web UI, with access control by a token and browser sessions which are kept across restarts.
- A fabric of the hub's own; the hub connects to each node of it at start and follows its state by subscription, and reconnects to a node it loses.
- The API of the nodes and their state, and live updates over Server-Sent Events; the device list in the web UI.
- Control: on/off, brightness, white color temperature and color, from the web UI and over the API.
- Adding a device with its setup code, over Bluetooth (with the Wi-Fi network to join) or the network; a device which needs a factory reset to be added is told at once.
- Removing a device: it leaves the hub's fabric, or, when it cannot, the hub forgets it.
- `homewire serve`, `open`, `status` and `token`.
- A .deb package for amd64 and arm64 with a hardened systemd service, running as the user `homewire`, with its data in `/var/lib/homewire`.

[Unreleased]: https://github.com/cybergarage/homewire/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/cybergarage/homewire/compare/v0.9.0...v1.0.0
[0.9.0]: https://github.com/cybergarage/homewire/releases/tag/v0.9.0
