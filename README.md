<p align="center">
  <img src="assets/logos/logo.png" alt="ShengOS logo" width="180">
</p>

<h1 align="center">ShengOS Daemon</h1>

<p align="center">
  Host daemon for synchronizing clipboard data across ShengOS machines over MQTT.
</p>

<p align="center">
  <a href="#quick-start">Quick start</a>
  ·
  <a href="#configuration">Configuration</a>
  ·
  <a href="#clipboard-sync">Clipboard sync</a>
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

ShengOS Daemon is a Go MQTT client for ShengOS fleet hosts. Its current feature is
clipboard synchronization between graphical desktop sessions. The same daemon
can run on a headless host with clipboard support disabled.

The fleet broker is Eclipse Mosquitto on `jasonkwh-bcm2711`, reachable on the
tailnet at port `1883`. Its NixOS configuration is maintained in
[my-nixos-configurations](https://github.com/jasonkwh/my-nixos-configurations).
The daemon connects to a broker; it does not run or configure Mosquitto itself.

## Quick start

Requirements: Go (see `go.mod`) and a reachable MQTT broker. From the repository
root, run:

```bash
make serve
```

This runs:

```bash
go run main.go --config=./config/config.yaml serve
```

The daemon reads the YAML file passed with `--config` (default: `config.yaml`),
connects to MQTT, and logs `daemon started` once connected. If clipboard sync is
enabled, it also logs `watching clipboard`. `SIGINT` and `SIGTERM` trigger
shutdown and broker disconnect.

## Configuration

`config/config.yaml` is the example configuration:

```yaml
LogLevel: 0 # zap level: -1 debug, 0 info
PProfEnabled: false

Clipboard:
  Enabled: true # disable on headless hosts
  SSHUser: jasonkwh # user embedded in file-list SFTP URIs

MQTTConn:
  Host: jasonkwh-bcm2711
  Port: 1883
  KeepAlive: 30
  # Username and Password are optional.
```

`MQTTConn.Username` and `MQTTConn.Password` may be supplied when the broker
requires credentials. `PProfEnabled` starts the optional profiling server from
the `perf` package. Keep profiling disabled unless it is needed and its network
exposure is understood.

For a headless broker host, set `Clipboard.Enabled: false`. The daemon will
still connect to MQTT but will not initialize a local clipboard or subscribe to
clipboard messages.

## Clipboard sync

When enabled, the daemon watches the local X11 or Wayland clipboard using
[`golang.design/x/clipboard`](https://github.com/golang-design/clipboard) and
publishes text, image, or file-list data. Incoming messages are applied to the
local clipboard. Clipboard content marked sensitive by the clipboard watcher is
not published. The daemon also ignores its own host's messages and suppresses
the clipboard watcher event caused by applying a received update, preventing a
copy-back loop.

File-list clipboard entries are converted from local `file://` URIs into
`sftp://<user>@<origin-host>/<path>` URIs. This shares references to files, not
the file contents; peers need network/SSH access to the origin host and
appropriate permissions to retrieve them.

Clipboard MQTT details:

| Property       | Value                                                 |
| -------------- | ----------------------------------------------------- |
| Topic          | `shengos/v1/clipboard`                                |
| Payload        | MessagePack (`origin`, `format`, `data`)              |
| QoS            | 1 (at least once)                                     |
| Publish retain | No                                                    |
| Subscription   | `NoLocal`; do not send retained messages on subscribe |
| Client ID      | `shengos-<hostname>`                                  |

The MQTT client uses a clean start on its initial connection, a 60-second
session expiry, and automatically reconnects/resubscribes through `autopaho`.
Inbound messages are acknowledged after the topic handler successfully applies
them.

Clipboard data is sent to the configured MQTT broker. Do not treat MQTT as an
end-to-end encrypted channel: secure the broker and its network path (the ShengOS
fleet uses Tailscale), and only connect trusted hosts.

## Nix

The flake provides `shengos-daemon` packages for `x86_64-linux` and
`aarch64-linux`, plus a NixOS module at `nixosModules.default`.

Evaluate a package derivation:

```bash
nix eval --raw .#packages.x86_64-linux.shengos-daemon.drvPath
nix eval --raw .#packages.aarch64-linux.shengos-daemon.drvPath
```

The module provides `services.shengos-daemon.enable`, `package`, and
`configFile` options. Add the flake as an input and enable the module in your
NixOS configuration, for example:

```nix
{
  imports = [ shengos-daemon.nixosModules.default ];
  services.shengos-daemon = {
    enable = true;
    configFile = "/path/to/config.yaml";
  };
}
```

The systemd unit runs `shengosd --config=<configFile> serve` and restarts on
failure. The YAML file must exist on the target host and point to its intended
broker/configuration.

## Commands

| Command                 | Description                               |
| ----------------------- | ----------------------------------------- |
| `make serve`            | Run with `config/config.yaml`             |
| `make lint`             | Run `golangci-lint` using `.golangci.yml` |
| `make test-unit`        | Run `go test -count=1 ./...`              |
| `make test-coverage`    | Write atomic coverage to `coverage.out`   |
| `make test-clean-cache` | Clear the Go test cache                   |
| `make build`            | Build Linux amd64 and arm64 binaries      |
| `make build-amd64`      | Build Linux amd64 binary                  |
| `make build-arm64`      | Build Linux arm64 binary                  |
| `make build-mocks`      | Regenerate mocks with mockery             |
| `make update`           | Update flake inputs                       |

`serve` is the default Make target.
