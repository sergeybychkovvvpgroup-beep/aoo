# Notes storage layout

Recommended layout for a notes repository:

```text
notes/
  commands/
    git/status.yaml
    ssh/router.yaml
  notes/
    proxmox/backup.md
    network/ipam.md
  snippets/
    vyos/router-beria-wan.config
```

Rules:

- keep one YAML note object per file;
- use directories by topic instead of large `*_cmds.yaml` or `*_notes.yaml` files;
- use Markdown/raw files for long prose and snippets;
- command files can be tiny:

```yaml
desc: git status short
cmd: git status --short
```

`f` scans directories recursively, so this layout works without extra config.
