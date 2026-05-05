# aoo

Fast terminal notes and command launcher.

Store commands in YAML and keep real snippets/configs as plain files. Use fuzzy search across notes and files, find what you need in seconds, and run commands directly from `aoo`.

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
If `notes_dir` is a git repo, `aoo` auto-runs `fetch/pull/push` on startup and after note edits.
If your SSH key has a passphrase, keep it loaded in `ssh-agent`.

Check config:

```bash
aoo config
aoo config show
```

Useful UI settings in `config.yaml`:

```yaml
full_screen: true
picker_height: 14
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

Both commands create a YAML scaffold, open it in `$VISUAL` / `$EDITOR`, and after editor exit auto-run commit/pull/push when `notes_dir` is a git repo.

## Notes Format

```yaml
- desc: example ssh
  text: main host access
  action: ssh admin@example.local

- desc: example url
  text: https://service.example.local

- desc: example command
  cmd: dig nas.example.local A +short
```

Behavior:

- `Enter` runs the main action directly
- if a note has `cmd`, `Enter` runs it
- if a note has only `text`, `Enter` shows it
- old `actions` format still works as a compatibility fallback
- `text` shows text
- `cmd` runs immediately after selection
- `full_screen: true` uses the terminal alternate screen and full terminal height
- when `full_screen: false`, picker height is limited by `picker_height`, so prior terminal output stays visible
- plain query searches only notes and files
- `:query` or `>query` searches only commands
- `show_preview: false` makes each search result a single line
- `preview_pane: true` shows a right-side preview panel like `fzf`
- keep one note = one main thing whenever possible

Config command example:

```yaml
- cmd: $EDITOR ~/.config/aoo/config.yaml
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
