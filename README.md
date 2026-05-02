# aoo

Fast terminal notes and command launcher for self-hosted infra.

Search YAML notes, open snippets, run saved commands, and use reusable command templates with prompted arguments.

## Install

```bash
git clone https://github.com/sergeyb/aoo.git
cd aoo
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
```

## First Run

Store your real notes outside the repo:

```bash
mkdir -p ~/.local/share/aoo/notes
aoo set-folder ~/.local/share/aoo/notes
```

Check config:

```bash
aoo config show
```

## Usage

```bash
aoo
aoo --query ssh
aoo --query nmap
aoo validate --dir ~/.local/share/aoo/notes
```

Themes:

```bash
aoo themes
aoo set-theme catppuccin-mocha
aoo --theme nord
```

## Notes Format

```yaml
- desc: example ssh
  run: ssh admin@example.local

- desc: example url
  note: |
    https://service.example.local

- desc: example nmap template
  template: sudo nmap -O {{host}}
  args:
    - name: host
      prompt: Host or IP
      example: 192.168.1.1
```

`run` runs a command, `note` prints text, `template` asks for args and then builds the final command.

## Themes

- `catppuccin-mocha`
- `catppuccin-latte`
- `dracula`
- `nord`
- `solarized-dark`
- `solarized-light`

## Repo Layout

- `examples/notes/` safe example notes for validation and demo
- `cmd/`, `internal/` application code

