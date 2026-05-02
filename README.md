# aoo

Fast terminal notes and command launcher.

Store notes, commands, hosts, and snippets in YAML. Use fuzzy search across note contents, find what you need in seconds, and run commands directly from notes.

![aoo demo](docs/demo.gif)

## Install

```bash
git clone https://github.com/sergeybychkovvvpgroup-beep/aoo.git
cd aoo
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
```

Install `.deb` manually:

```bash
sudo dpkg -i ./aoo_<version>_amd64.deb
```

Install from APT repository:

```bash
curl -fsSL https://sergeybychkovvvpgroup-beep.github.io/aoo/public.key | \
  sudo gpg --dearmor -o /usr/share/keyrings/aoo-archive-keyring.gpg

echo "deb [signed-by=/usr/share/keyrings/aoo-archive-keyring.gpg] https://sergeybychkovvvpgroup-beep.github.io/aoo stable main" | \
  sudo tee /etc/apt/sources.list.d/aoo.list

sudo apt update
sudo apt install aoo
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

Picker shortcuts:

```text
Enter   open/run selected note
Alt+E   open selected note source in $VISUAL / $EDITOR
Esc     quit
Left/Right switch ALL / RUN / SHOW
Up/Down move
```

## Notes Format

```yaml
- desc: example ssh
  mode: run
  run: ssh admin@example.local

- desc: example url
  mode: show
  note: |
    https://service.example.local

- desc: example nmap template
  mode: run
  tags: [nmap, scan]
  template: sudo nmap -O {{host}}
  args:
    - name: host
      prompt: Host or IP
      example: 192.168.1.1
```

Modes:

- `mode: run` runs a command or a template-built command
- `mode: show` prints text

Behavior:

- `run` runs a command after confirmation
- `template` asks for args, shows the final command, then asks for confirmation
- `note` prints text

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
- `internal/bundled/` built-in help and starter notes
- `docs/demo.gif` quick terminal demo
- `docs/demo.tape` VHS source for the demo

## Releases

- Push a tag like `v0.2.0` to trigger GitHub Release autobuilds
- Release workflow publishes tarballs and `.deb` packages
- The same workflow publishes a signed APT repository to `gh-pages`
- Enable GitHub Pages for the `gh-pages` branch before using the APT URL
