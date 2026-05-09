# aoo

Terminal notes and command launcher.

## Install

```bash
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
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
aoo config
aoo config show
```

## Usage

```bash
aoo
aoo --query ssh
aoo validate --dir ~/.local/share/aoo/notes
aoo add "router dhcp"
aoo add cmd "restart nginx"
```

Keys:

- `Enter` open or run
- `Ctrl+E` edit
- `Ctrl+N` create a new note from current query
- `Esc` quit

## Note format

```yaml
- desc: ssh router
  text: main access
  cmd: ssh admin@router

- desc: nginx logs
  cmd: journalctl -u nginx -n 100
```

If a note has `cmd`, `Enter` runs it. If it has only `text`, `Enter` prints the note.

Plain query searches notes and files. `:query` or `>query` searches only commands.

## Raw files

You can store snippets as plain files: `netplan.yaml`, `bgp.conf`, `notes.md`.

`aoo` shows them in search and opens the content as a note.
