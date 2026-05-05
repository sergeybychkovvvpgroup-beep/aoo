# aoo

Fast terminal notes and command launcher.

Store notes, commands, hosts, and snippets in YAML. Use fuzzy search across note contents, find what you need in seconds, and run commands directly from notes.

![aoo demo](docs/demo.png)

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

## First Run

On first start, `aoo` can ask for your notes repo URL and clone it into `~/.local/share/aoo/notes`.

Main settings now live in one file:

```bash
~/.config/aoo/config.yaml
```

The file is auto-created with comments and examples.

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
aoo config
aoo config show
```

Useful UI settings in `config.yaml`:

```yaml
full_screen: true
picker_height: 14
search_mode: hybrid
show_preview: false
preview_pane: false
```

## Usage

```bash
aoo
aoo validate --dir ~/.local/share/aoo/notes
```

Themes:

```bash
aoo themes
aoo set-theme catppuccin-mocha
```

Picker shortcuts:

```text
Enter   open or run
Ctrl+E  edit selected YAML
Ctrl+N  create a new note from current query
Esc     quit
```

Quick add:

```bash
aoo add "router dhcp"
aoo add cmd "restart nginx"
```

Both commands create a YAML scaffold, open it in `$VISUAL` / `$EDITOR`, and after editor exit auto-commit + auto-push changes when `notes_dir` is a git repo.

## Notes Format

```yaml
- desc: example ssh
  text: main host access
  action: ssh admin@example.local

- desc: example url
  text: https://service.example.local

- desc: example command
  action: ssh admin@example.local

- desc: example nmap template
  tags: [nmap, scan]
  template: sudo nmap -O {{host}}
  args:
  - name: host
    prompt: Host or IP
    example: 192.168.1.1
```

Behavior:

- `Enter` runs the main action directly
- if a note has `cmd`, `Enter` runs it
- if a note has only `text`, `Enter` shows it
- old `actions` format still works as a compatibility fallback
- `text` shows text
- `cmd` runs immediately after selection
- `template` asks for args, renders the final command, and runs it immediately
- `full_screen: true` uses the terminal alternate screen and full terminal height
- when `full_screen: false`, picker height is limited by `picker_height`, so prior terminal output stays visible
- `search_mode: flat` keeps the old behavior and shows individual matching notes/commands
- `search_mode: entry-first` groups matching entries from one YAML file into one result and shows commands after `Enter`
- `search_mode: hybrid` behaves like `entry-first`, but `:query` or `>query` switches to direct flat search
- `show_preview: false` makes each search result a single line
- `preview_pane: true` shows a right-side preview panel like `fzf`
- keep one note = one main thing whenever possible

Built-in template variables:

- `{{aoo_notes_dir}}`
- `{{aoo_app_dir}}`
- `{{aoo_config_file}}`

Config command example:

```yaml
- action: $EDITOR {{aoo_config_file}}
```

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

## Releases

- Push a tag like `v0.2.0` to trigger GitHub Release autobuilds
- Release workflow publishes tarballs and `.deb` packages
