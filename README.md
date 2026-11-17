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
  <a href="https://github.com/jasonkwh/shengos-daemon/actions/workflows/eval.yml"><img src="https://img.shields.io/github/actions/workflow/status/jasonkwh/shengos-daemon/eval.yml?label=eval" alt="eval"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
</p>

---

Each ShengOS host runs this daemon and connects to the fleet broker over
MQTT. Desktop machines watch local state, publish it, and apply what peers
send. The fleet hub runs the same daemon with clipboard disabled so it can
host Mosquitto and take later message types without an X11 or Wayland
session. The broker is Eclipse Mosquitto on the hub (`jasonkwh-bcm2711`,
`tailscale0:1883`), declared in
[my-nixos-configurations](https://github.com/jasonkwh/my-nixos-configurations).

- **Clipboard** — copy and paste across desktop machines. X11 and Wayland
  only. Text, images, and file lists follow the host you are on. Sensitive
  copies stay on the machine that made them, and a file list points peers
  at the files over SSH.

## Quick start

From a checkout, with the hub broker reachable:

```bash
make serve
```

That runs `go run main.go --config=./config/config.yaml serve`.
`config/config.yaml` sets the log level, optional pprof, whether clipboard
sync is enabled, the SSH user embedded in file-list URIs, and the MQTT host,
port, keepalive, and optional username and password.

```yaml
LogLevel: 0 # zap: -1 debug, 0 info
PProfEnabled: false

Clipboard:
  Enabled: true # false on the headless fleet hub
  SSHUser: jasonkwh

MQTTConn:
  Host: jasonkwh-bcm2711
  Port: 1883
  KeepAlive: 30
```

The process logs `daemon started` after the MQTT session is up. With clipboard
enabled it also logs `watching clipboard` and subscribes to
`shengos/v1/clipboard`. With `Clipboard.Enabled: false` it stays connected
and does not subscribe to clipboard topics. `SIGINT` or `SIGTERM` disconnects
from the broker.

## Clipboard

`internal/clipboard` watches the X11 or Wayland session clipboard
([`golang.design/x/clipboard`](https://github.com/golang-design/clipboard))
and publishes a message: origin hostname, format (`text`, `image`, or
`files`), and the bytes. Set `Clipboard.Enabled: false` on headless hosts
(the Pi hub has no display). The daemon still runs there for MQTT and future
message types; it does not open a clipboard session or subscribe to the
clipboard topic. `internal/bus` connects with
[paho.golang](https://github.com/eclipse/paho.golang) (`autopaho`) and
resubscribes whenever the session comes up. Inbound publishes are routed by
MQTT topic to a registered `DaemonTopicHandler`; the clipboard handler
(`internal/handler`) decodes the payload and writes it to the local clipboard.
The broker message is acknowledged after that write succeeds.

Each connection starts clean. The subscription is `NoLocal`, so the broker
does not deliver a host its own publish. The daemon also remembers the
hash of a clipboard it just applied and drops the watch event that comes
back, so a remote copy is not published again.

|                |                                                        |
| -------------- | ------------------------------------------------------ |
| Topic          | `shengos/v1/clipboard`                                 |
| Encoding       | MessagePack                                            |
| QoS            | 1 (at least once)                                      |
| Retain         | no                                                     |
| Subscribe      | `NoLocal`; retained messages are not sent on subscribe |
| Client ID      | `shengos-<hostname>`                                   |
| Session expiry | 60 seconds                                             |

## Commands

| Command                 | Description                               |
| ----------------------- | ----------------------------------------- |
| `make serve`            | Run the daemon with `config/config.yaml`  |
| `make lint`             | `golangci-lint` (`.golangci.yml`)         |
| `make test-unit`        | `go test ./...`                           |
| `make test-coverage`    | Atomic coverage profile in `coverage.out` |
| `make test-clean-cache` | `go clean -testcache`                     |
| `make build-mocks`      | Regenerate `mocks/` with mockery          |

`serve` is the default target.

## Layout

```
main.go
cmd/                   # cobra: serve wires clipboard, topic handlers, subs, bus
config/                # YAML config
internal/
  clipboard/           # watch, apply, file:// → sftp://
  bus/                 # MQTT connect, publish, subscribe, dispatch by topic
  handler/             # per-topic handlers (e.g. clipboard inbound apply)
  message/             # topics, subscriptions, MessagePack payloads
  adapter/             # DaemonBus, Clipboard, DaemonTopicHandler
mocks/
version.yaml           # bumped on push to main
.github/workflows/     # lint, vet, test, coverage, version bump
```
