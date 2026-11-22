# Development

Building homewire from source, testing it, and the layout of the repository.

## Building from source

Go 1.25 or later, and Node.js 22 or later with npm. The hub depends on go-matter v1.0.0.

```sh
make build             # builds web/dist, then bin/homewire
bin/homewire serve     # http://127.0.0.1:8580
```

Always build with `make build`: a plain `go build` embeds whatever is left in `web/dist`, which may be an older UI.

## Testing

```sh
make test       # gofmt, go vet, go test -race
go test -short ./...  # without the test which needs mDNS on the network
make check-web  # svelte-check

## Third-party notices

`THIRD_PARTY_NOTICES.md` lists the Go modules built into the executable (`go version -m bin/homewire`), the Go standard library and the JavaScript bundled into the web UI, with their licenses. Update it when a dependency is added, removed or upgraded.

## Layout

| Path | Role |
| --- | --- |
| `cmd/homewire` | The command: `serve` (the daemon: flags, listener, shutdown) and the other subcommands |
| `internal/hub` | The hub: its fabric, the sessions and subscriptions with the nodes, their state, and the commands sent to them |
| `internal/auth` | Access token file, bearer and session authentication |
| `internal/server` | HTTP API, event stream, Host and Origin checks, web UI serving |
| `packaging` | systemd unit, `/etc/default/homewire`, the desktop launcher and the icons (`packaging/icons`, made from `homewire.svg` and, for 16 to 24 px, `homewire-small.svg`), the package's install scripts, and the test of the installed package run by CI |
| `web` | Svelte + Vite + TypeScript UI; `web/dist` is embedded |
| `docs` | The documentation, and the screenshots of the README (`docs/images`) |
