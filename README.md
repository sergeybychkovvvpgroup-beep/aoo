# aoo

Fast terminal notes and command launcher for self-hosted infrastructure notes.

`aoo` scans a directory with YAML notes, gives you instant multi-word fuzzy search, prints notes, and runs commands directly from the terminal. It is designed for infra and support workflows where you need to jump to the right host, tunnel, URL, or snippet with minimal typing.

## Features

- Multi-word fuzzy search across `desc`, file name, tags, command, note body, and banner text
- Single static-friendly Go binary
- Optional `banner` shown before a command runs
- Optional `check` before a command runs
- Validation mode for broken YAML files
- Backward-compatible YAML format with `run` and `note`
- Command templates with prompted arguments
- Zero runtime dependencies after build

## Note Format

```yaml
- desc: proxmox chashnikovo external tunnel localhost 8006
  tags:
    - proxmox
    - tunnel
  banner: |
    Open in browser:
    https://127.0.0.1:8006
  check: ssh -o BatchMode=yes -o ConnectTimeout=5 sergeyb@178.238.119.166 exit
  check_error: |
    SSH auth check failed.
    Verify that your key is loaded.
  run: ssh -L 8006:10.117.0.1:8006 sergeyb@178.238.119.166

- desc: proxmox chashnikovo web local
  note: |
    https://10.117.0.1:8006

- desc: nmap os detect host
  template: sudo nmap -O {{host}}
  args:
    - name: host
      prompt: Host or IP
      example: 192.168.1.1
```

## Requirements

- Linux terminal with interactive TTY
- POSIX shell at `/bin/sh` for running `run:` commands
- A notes directory with `.yaml` or `.yml` files
- Go `1.22+` only if you build from source

Repository layout:

- `cmd/`, `internal/`: application code
- `notes/`: local or example YAML notes

## Install Dependencies

Source:
- Go installation: https://go.dev/doc/install
- GoReleaser installation: https://goreleaser.com/install/

Ubuntu / Debian:

```bash
sudo apt update
sudo apt install -y golang-go
```

Optional release tooling:

```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

## Build

```bash
git clone https://github.com/sergeyb/aoo.git
cd aoo
make tidy
make build
```

## Install

Local install from source:

```bash
git clone https://github.com/sergeyb/aoo.git
cd aoo
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
```

Install with `go install`:

```bash
go install github.com/sergeyb/aoo/cmd/aoo@latest
```

Install from `.deb` package:

```bash
sudo dpkg -i aoo_<version>_linux_amd64.deb
```

Check:

```bash
aoo version
```

## Usage

By default the tool resolves the notes directory in this order:

1. `--dir`
2. `AOO_NOTES_DIR`
3. `TERM_NOTES_DIR` for legacy compatibility
4. `~/.config/aoo/config.yaml`
5. current working directory if it contains `*.yaml`

Initial setup:

```bash
aoo set-folder ~/notes
```

Show current configuration:

```bash
aoo config show
```

For this repository itself, note files are stored under `./notes`.

Run interactive picker:

```bash
aoo
```

Start with a prefilled query:

```bash
aoo --query "chash router"
```

Validate note files:

```bash
aoo validate
```

Command templates:

```bash
aoo --query nmap
```

For `template` entries, `aoo` prompts for args, shows the final command, and asks for confirmation before execution.

Pre-check example:

- `check`: shell command that must exit with code `0` before `run`
- `check_error`: friendly message shown when `check` fails

## Search Behavior

The matcher is intentionally practical rather than "exact only":

- Each word in the query must match somewhere
- Substring matches rank highest
- Approximate subsequence matches also work
- File name tokens are included automatically

That means a query like `chash router` can match a note from `chashnikovo-wh2.yaml` even if `router` is only present in `desc`.

## Release

This repository includes Go module metadata and is ready for:

- GitHub Releases with prebuilt binaries
- `.deb` packaging through GoReleaser / nfpm
- `go install` from source
- CI validation on each push / pull request

Recommended next step before public release: choose a license and add CI for release builds.
