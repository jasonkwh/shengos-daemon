<p align="center">
  <img src="assets/logos/logo.png" alt="ShengOS logo" width="180">
</p>

<h1 align="center">ShengOS Daemon</h1>

<p align="center">
  Host daemon for ShengOS that syncs machine state across devices over MQTT.
</p>

<p align="center">
  <a href="#quick-start">Quick start</a>
  ·
  <a href="#clipboard">Clipboard</a>
  ·
  <a href="#commands">Commands</a>
</p>

<p align="center">
  <a href="https://github.com/jasonkwh/shengos-daemon/actions/workflows/coverage.yml"><img src="https://img.shields.io/github/actions/workflow/status/jasonkwh/shengos-daemon/coverage.yml?label=coverage" alt="coverage"></a>
  <a href="https://github.com/jasonkwh/shengos-daemon/actions/workflows/lint.yml"><img src="https://img.shields.io/github/actions/workflow/status/jasonkwh/shengos-daemon/lint.yml?label=lint" alt="lint"></a>
  <a href="https://github.com/jasonkwh/shengos-daemon/actions/workflows/test.yml"><img src="https://img.shields.io/github/actions/workflow/status/jasonkwh/shengos-daemon/test.yml?label=test" alt="test"></a>
  <a href="https://github.com/jasonkwh/shengos-daemon/actions/workflows/vet.yml"><img src="https://img.shields.io/github/actions/workflow/status/jasonkwh/shengos-daemon/vet.yml?label=vet" alt="vet"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
</p>

---

Each ShengOS host runs this daemon. It watches local state and publishes it
to the fleet broker; peers apply what they receive. The broker is Eclipse
Mosquitto on the fleet hub (`jasonkwh-bcm2711`, `tailscale0:1883`), declared
in [my-nixos-configurations](https://github.com/jasonkwh/my-nixos-configurations).

- **Clipboard** — text, images, and file lists follow the host you are on.
  A copy on one machine lands on the others.
- **Local changes** — a host publishes clipboard changes made here. It
  skips entries the clipboard marks sensitive, messages whose origin is
  this hostname, and the write it just applied.
- **Files stay put** — a file list is rewritten from local `file://` paths
  to `sftp://<user>@<hostname>/<path>`. A peer opens the file over SSH.
- **One topic** — MessagePack on `shengos/v1/clipboard`, QoS 1, retained.
  The client id is `shengos-<hostname>`.

## Quick start

From a checkout, with the hub broker reachable:

```bash
make serve
```

That runs `go run main.go --config=./config/config.yaml serve`.
`config/config.yaml` sets the log level, optional pprof, the SSH user
embedded in file-list URIs, and the MQTT host, port, keepalive, and
optional username and password.

```yaml
LogLevel: 0          # zap: -1 debug, 0 info
PProfEnabled: false

Clipboard:
  SSHUser: jasonkwh

MQTTConn:
  Host: jasonkwh-bcm2711
  Port: 1883
  KeepAlive: 30
```

The process logs `daemon started` after the MQTT session is up and the
clipboard watch is running. `SIGINT` or `SIGTERM` disconnects from the broker.

Requires Go 1.27.

## Clipboard

`internal/clipboard` watches the session clipboard
([`golang.design/x/clipboard`](https://github.com/golang-design/clipboard))
and publishes a message: origin hostname, format (`text`, `image`, or
`files`), and the bytes. `internal/bus` connects with
[paho.golang](https://github.com/eclipse/paho.golang) (`autopaho`) and
resubscribes whenever the session comes up. `internal/handler` decodes the
payload and writes it to the local clipboard. The broker message is
acknowledged after that write succeeds.

Each connection starts clean. The subscription is `NoLocal`, so the broker
does not deliver a host its own publish. The daemon also remembers the
hash of a clipboard it just applied and drops the watch event that comes
back, so a remote copy is not published again.

|                |                                                                 |
| -------------- | --------------------------------------------------------------- |
| Topic          | `shengos/v1/clipboard`                                          |
| Encoding       | MessagePack                                                     |
| QoS            | 1 (at least once)                                               |
| Retain         | yes                                                             |
| Subscribe      | `NoLocal`; retained messages delivered on subscribe             |
| Client ID      | `shengos-<hostname>`                                            |
| Session expiry | 60 seconds                                                      |

## Commands

| Command                 | Description                                          |
| ----------------------- | ---------------------------------------------------- |
| `make serve`            | Run the daemon with `config/config.yaml`             |
| `make lint`             | `golangci-lint` (`.golangci.yml`)                    |
| `make test-unit`        | `go test ./...`                                      |
| `make test-coverage`    | Atomic coverage profile in `coverage.out`            |
| `make test-clean-cache` | `go clean -testcache`                                |
| `make build-mocks`      | Regenerate `mocks/` with mockery                     |

`serve` is the default target.

## Layout

```
main.go
cmd/                   # cobra: serve, zap, shutdown
config/                # YAML config
internal/
  clipboard/           # watch, apply, file:// → sftp://
  bus/                 # MQTT connection
  handler/             # topic → clipboard apply
  message/             # topic, QoS, MessagePack payload
  adapter/             # bus, clipboard, and handler interfaces
mocks/
version.yaml           # bumped on push to main
.github/workflows/     # lint, vet, test, coverage, version bump
```
