# HTTP API

The API the web UI uses, for scripts and other programs. A script authenticates with `Authorization: Bearer <token>`; see [access control](security.md).

All paths are under `/api/v1`. Errors are `{"error": {"code": "...", "message": "..."}}`.

| Method and path | Access | Result |
| --- | --- | --- |
| `GET /health` | Public | `{"status": "ok", "version": "..."}` |
| `GET /session` | Public | `{"authRequired": bool, "authenticated": bool}` |
| `POST /session` | Public | Body `{"token": "..."}`; starts a browser session |
| `DELETE /session` | Public | Ends the browser session |
| `POST /session/codes` | Token | Issues a one-time sign-in code: `{"code", "expiresAt", "path"}`; see [Signing a browser in](security.md#signing-a-browser-in) |
| `GET /nodes` | Token | `{"nodes": [node, ...]}`, ordered by node ID |
| `GET /nodes/{id}` | Token | One node |
| `GET /events` | Token | The state of the nodes as Server-Sent Events; see below |
| `POST /nodes/{id}/commands` | Token | Invokes a command on a node; see below |
| `DELETE /nodes/{id}` | Token | Removes a node; see below |
| `POST /commissioning` | Token | Starts adding a device; see below |
| `GET /commissioning` | Token | `{"commissioning": commissioning}`, the latest one, or `null` |
| `GET /hub` | Token | `{"version": "...", "authRequired": bool, "fabricId": "<16 hex digits>", "nodes": {"total": n, "online": n}}` |

A node:

```json
{
  "id": "0000000016A5E84A",
  "vendorId": 4631,
  "productId": 1,
  "commissionedAt": "2026-11-12T09:00:00Z",
  "status": "online",
  "lastReportAt": "2026-11-12T14:20:31.2Z",
  "attributes": [
    {"endpoint": 1, "cluster": 6, "attribute": 0, "value": true, "updatedAt": "2026-11-12T14:20:31.2Z"}
  ]
}
```

- `id` is 16 hexadecimal digits, because a JSON number cannot hold a 64-bit ID exactly.
- `status` is `connecting`, `online` or `offline`; `error` says why a node is offline.
- `attributes` are the last values the node reported, by endpoint, cluster and attribute ID. A value is what the node sent, without the types of the data model applied: a boolean, a number, a string, `null`, an array for a list, and an object keyed by field number for a structure. An integer beyond 2^53 is a string of decimal digits, and an octet string a string of hexadecimal digits. An attribute the node refused to report has `status`, the Interaction Model status code, instead of a value.

## Events

`GET /events` is a `text/event-stream`. The session cookie authenticates a browser's `EventSource`; a script sends the bearer token.

| Event | Data | Meaning |
| --- | --- | --- |
| `nodes` | `{"nodes": [node, ...]}` | Sent first on every connection: all the nodes |
| `status` | `{"node": id, "status": "...", "error": "..."}` | A node's status changed |
| `attributes` | `{"node": id, "priming": bool, "attributes": [...]}` | Attributes changed. With `priming` true they are the whole state of the node and replace what the client holds |
| `node` | node | A node joined the hub |
| `removed` | `{"node": id}` | A node was removed from the hub |
| `commissioning` | `{"commissioning": commissioning}` | A commissioning started or ended |

A comment line is sent every 15 seconds on an idle stream. A client which falls 256 events behind is disconnected; it connects again and starts from a fresh `nodes` event.

```sh
curl -N http://127.0.0.1:8580/api/v1/events
```

## Commands

`POST /nodes/{id}/commands` invokes a command of a cluster on an endpoint of a node:

```sh
curl -X POST -H 'Content-Type: application/json' \
  -d '{"endpoint": 1, "cluster": 8, "command": 4, "fields": {"Level": 127, "TransitionTime": 5, "OptionsMask": 0, "OptionsOverride": 0}}' \
  http://127.0.0.1:8580/api/v1/nodes/0000000016A5E84A/commands
```

- `cluster` and `command` are the numeric IDs; `fields` are the command's fields by the names of the data model, and may be left out for a command without fields.
- The fields are checked against the definition of the command in go-matter's data model catalog and encoded from it. The catalog has the On/Off, Level Control, Color Control, Door Lock, Basic Information, Descriptor and Temperature Measurement clusters; commands of other clusters are refused. Lists, structures and floating point fields are not supported, nor commands which need a timed invoke.
- The command is sent over the session the hub keeps with the node. The answer is `{"ok": true}` once the node has accepted it; the new state arrives over `GET /events`, as every other change does.

| Status | Code | Meaning |
| --- | --- | --- |
| 400 | `unsupported_command` | The command is not in the catalog, or needs a timed invoke |
| 400 | `invalid_fields` | A field is missing, unknown, of the wrong type or out of range |
| 404 | `not_found` | No such node |
| 409 | `node_offline` | The hub has no session with the node right now |
| 502 | `command_failed` | The node refused the command; the message names the Interaction Model status |
| 504 | `node_error` | The node did not answer, or the session failed |

The web UI uses On and Off (On/Off cluster), MoveToLevelWithOnOff (Level Control), MoveToColorTemperature and MoveToHueAndSaturation (Color Control), with a half-second transition. It offers the color temperature and the colors only when the light's ColorCapabilities say it takes them.

## Commissioning

`POST /commissioning` starts adding a device and answers `202` once it has started:

```sh
curl -X POST -H 'Content-Type: application/json' \
  -d '{"code": "MT:...", "wifi": {"ssid": "home", "password": "..."}}' \
  http://127.0.0.1:8580/api/v1/commissioning
```

- `code` is the QR code text or the manual pairing code. `wifi` is for a device found over Bluetooth; leave it out for a device on the network already.
- The progress arrives over `GET /events`: a `commissioning` event when it starts, at each stage it reaches, and when it ends, and a `node` event for the new node. The web UI shows the stages as a list.

```json
{"id": 1, "state": "failed", "overBluetooth": true, "startedAt": "...", "finishedAt": "...",
 "step": {"stage": "case", "attempt": 2, "detail": "192.168.1.80"},
 "error": "...", "errorCode": "failed"}
```

`state` is `running`, `succeeded` (with `node`, the new node's ID) or `failed`, with `errorCode` one of `not_found`, `bluetooth_unavailable` (not found, and the hub could not search over Bluetooth: its adapter is missing, powered off or blocked), `wifi_required` (found over Bluetooth only, and no Wi-Fi given), `wifi_failed` (the device told it could not join the Wi-Fi network; `error` names why, such as `AuthFailure` or `NetworkNotFound`), `not_in_commissioning_mode`, `timeout` and `failed`.

`step` is the last stage reached, the one a failed commissioning failed at. `stage` is, in order: `discovering` (over Bluetooth and on the network, round after round, for up to 30 seconds), `connecting`, `pase` (the session with the setup code), `attestation`, `credentials`, `network` (joining the Wi-Fi network; a device found over Bluetooth only), `operationalDiscovery` (finding the device on the network), `case` (the session with the hub's credentials) and `complete`. `attempt` counts the rounds of `discovering` and the attempts of `operationalDiscovery` and `case`; `detail`, for `case`, holds the addresses tried.

| Status | Code | Meaning |
| --- | --- | --- |
| 400 | `invalid_code` | The setup code is malformed, or has a checksum or passcode which is not valid |
| 400 | `invalid_wifi` | The SSID is empty or longer than 32 bytes, or the password is not 8 to 63 characters or 64 hexadecimal digits |
| 409 | `busy` | Another device is being added |
| 503 | `not_ready` | The hub has not started |

## Sharing a node with another app

`POST /nodes/{id}/share` opens a commissioning window on an online node for 10 minutes, with a one-time passcode, for another app or hub (Google Home, Alexa, Apple Home, SmartThings) to add the node to its own fabric; the hub stays on the node's fabric. It answers `201`:

```json
{"manualCode": "1234-567-8901", "qrCode": "MT:...", "expiresAt": "..."}
```

The hub does not log or keep the codes. `DELETE /nodes/{id}/share` closes the window. A node with a window open already answers `409` `busy`; an offline node `409` `node_offline`; a node which refused or did not answer, `502` `share_failed`. The web UI's device card has **Share with another app…**, with the QR code and the setup code.

## Removing a node

`DELETE /nodes/{id}` removes a node from the fabric and from the hub, and answers `{"fromDevice": true}`. With `?local=true`, a node which is offline or fails to leave the fabric is forgotten anyway, and the answer is `{"fromDevice": false}`: the device still holds the fabric. The clients are told with a `removed` event.

| Status | Code | Meaning |
| --- | --- | --- |
| 404 | `not_found` | No such node |
| 409 | `node_offline` | The node is offline, so it cannot leave the fabric; retry with `?local=true` to forget it |
| 409 | `busy` | The node is being removed already |
| 502 | `remove_failed` | The node did not leave the fabric; retry with `?local=true` to forget it |
| 503 | `not_ready` | The hub has not started |
