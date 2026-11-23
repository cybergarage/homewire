# Commands and options

The subcommands of `homewire`, and the options of `homewire serve`. The system service takes the options from `/etc/default/homewire`, as `HOMEWIRE_` variables.

## Commands

| Command | Does |
| --- | --- |
| `homewire serve` | Runs the hub: the controller, its web UI and its HTTP API |
| `homewire open` | Opens the web UI of the running hub in a browser, signed in; see [Signing a browser in](security.md#signing-a-browser-in) |
| `homewire status` | Prints the state of the running hub and of its nodes |
| `homewire token` | Prints the access token, for scripts which call the API |
| `homewire version` | Prints the version |

`open`, `status` and `token` take `-data-dir` too, and find the running hub through `serve.json`, which `homewire serve` writes into the data directory while it runs. Without `-data-dir` (or `HOMEWIRE_DATA_DIR`), they use the first of `~/.homewire`, `~/.go-matter-hub` and `/var/lib/homewire`, the data directory of the system service, in which a hub is running; when none is, the first which exists.

## Options of `homewire serve`


| Flag | Default | Meaning |
| --- | --- | --- |
| `-listen` | `127.0.0.1:8580` | Address to listen on |
| `-data-dir` | `~/.homewire` | Directory for the hub's fabric, its commissioned nodes and the access token. When only `~/.go-matter-hub` exists, the data directory from before the hub was named homewire, that one is used, with a warning: rename it to `~/.homewire` |
| `-vendor-id` | `0xFFF1` | Administrator vendor ID of the fabric the hub creates at its first start |
| `-log-level` | `info` | `debug`, `info`, `warn` or `error`. go-matter's own messages are logged with `component=go-matter`: all of them at `debug` (each step of commissioning and of each session), only its warnings and errors otherwise |
| `-allowed-host` | | Host name by which browsers may reach the hub; repeatable |
| `-require-auth` | off | Require the access token on the loopback interface too |

Each option can also be given in the environment, as `HOMEWIRE_` and the option's name in capitals with `_` for `-`: `HOMEWIRE_LISTEN=0.0.0.0:8580`, `HOMEWIRE_DATA_DIR=/var/lib/homewire`. `HOMEWIRE_ALLOWED_HOST` takes a list, separated by commas or spaces. An option on the command line wins over its variable.

## The system service

The .deb package installs `homewire` as a systemd service, `homewire.service`, and starts it.

- **User.** The service runs as the system user `homewire`, created by the package. It is added to the group `bluetooth`, so that it may use the Bluetooth adapter through BlueZ (package `bluez`, a dependency of the package).
- **Data.** The fabric, the commissioned nodes and the access token are kept in `/var/lib/homewire` (mode 0750, group `homewire`).
- **Options.** `/etc/default/homewire` gives the options of `homewire serve`, as `HOMEWIRE_` variables; see [the options](#options-of-homewire-serve). By default the hub listens on `127.0.0.1:8580`, this host only; set `HOMEWIRE_LISTEN=0.0.0.0:8580` to reach it from other machines, then restart it with `sudo systemctl restart homewire`. An upgrade keeps the file.
- **Commands.** `homewire open`, `status` and `token` read `/var/lib/homewire`, so they need membership of the group `homewire`:

```sh
sudo adduser $USER homewire     # then sign in again
homewire status
homewire open -print
```

- **Launcher.** The package also adds **homewire** to the desktop's applications (in GNOME, the app grid), with its icon. It runs `homewire open -notify`, which tells with a desktop notification when the hub cannot be opened, such as when it is not running or the user is not yet in the group `homewire`.

`sudo apt remove homewire` stops the service and keeps its data; `sudo apt purge homewire` deletes `/var/lib/homewire` and the user too. A device commissioned into a purged fabric has to be reset to factory defaults to be added again.
