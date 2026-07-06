# f

Terminal notes and command launcher. The project was formerly called `aoo`; `aoo` is still installed as a compatibility symlink.

## Install

```bash
make build
sudo make install
```

## Config

File: `~/.config/aoo/config.yaml`

Minimal example:

```yaml
notes_dir: ~/.local/share/aoo/notes
theme: fzf-dark
layout: bottom
full_screen: true
picker_height: 14
focus_mode: false
show_match_context: false
show_list_on_start: true
two_line_results: true
```

- `focus_mode` hides the hotkey footer for a cleaner picker
- `show_match_context` shows the matched line for the selected entry
- `show_list_on_start` shows notes immediately with an empty query
- `two_line_results` keeps description and command/text on separate lines

```bash
f config
f config show
```

## Usage

```bash
f
f --query ssh
f validate --dir ~/.local/share/aoo/notes
f add "router dhcp"
f add cmd "restart nginx"
f upgrade
```

Keys:

- `Enter` open or run
- `Alt+Enter` or `Ctrl+Y` print the command to the terminal without running it, so it can be copied (`Ctrl+Enter` is also supported when the terminal can distinguish it from plain Enter)
- `Ctrl+E` edit
- `Ctrl+N` create a new note from current query
- `Esc` quit

## Note format

Prefer one note per file, for example `ssh-router.yaml`:

```yaml
desc: ssh router
actions:
  - desc: main access
    text: ssh command for router
  - desc: connect
    cmd: ssh admin@router
```

Simple command-only files are also valid:

```yaml
desc: nginx logs
cmd: journalctl -u nginx -n 100
```

If a note has `cmd`, `Enter` runs it. If it has only `text`, `Enter` prints the note.

Plain query searches notes and files. `:query` or `>query` searches only commands.

## Raw files

You can store snippets as plain files: `netplan.yaml`, `bgp.conf`, `notes.md`.

`aoo` shows them in search and opens the content as a note.
