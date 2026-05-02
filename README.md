# aoo

Fast terminal notes and command launcher for self-hosted infra.

Search YAML notes, open snippets, run saved commands, and use reusable command templates with prompted arguments.

## Install

```bash
git clone <your-repo-url>
cd aoo
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
```

## First Run

On first start, `aoo` can ask for your notes repo URL and clone it into `~/.local/share/aoo/notes`.

You can also configure it manually:

```bash
aoo set-source
mkdir -p ~/.local/share/aoo/notes
aoo set-folder ~/.local/share/aoo/notes
aoo set-app-dir ~/workspace/aoo
```

If the notes repo is private, `aoo` prints the public key that should be added to Deploy Keys.
`aoo` does not auto-fetch notes repo on every start, so it will not keep asking for your SSH key passphrase.

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

Built-in template variables:

- `{{aoo_notes_dir}}`
- `{{aoo_app_dir}}`

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
