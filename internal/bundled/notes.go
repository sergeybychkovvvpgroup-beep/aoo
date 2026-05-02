package bundled

import "aoo/internal/notes"

const serviceYAML = `
- desc: aoo-help
  note: |
    Built-in notes are available even on a clean host.

    First setup:
      1. run: aoo set source
      2. choose repo or local folder
      3. reopen aoo

    Useful searches:
      aoo
      notes
      git
      make
      theme

- desc: aoo set source
  run: aoo set-source
  tags: [aoo, setup, source]

- desc: aoo config show
  run: aoo config show
  tags: [aoo, config]

- desc: aoo list themes
  run: aoo themes
  tags: [aoo, theme]

- desc: aoo version
  run: aoo version
  tags: [aoo, version]

- desc: aoo make build
  run: make build
  tags: [aoo, make, build]

- desc: aoo make update
  run: sudo make update
  tags: [aoo, make, install, update]

- desc: aoo make test
  run: make test
  tags: [aoo, make, test]

- desc: aoo make validate
  run: make validate
  tags: [aoo, make, validate]
`

const templatesYAML = `
- desc: aoo rebuild and install
  template: cd {{aoo_app_dir}} && make update
  args: []
  tags: [aoo, make, update]

- desc: aoo git add commit push
  template: cd {{aoo_app_dir}} && git add . && git commit -m "{{message}}" && git push
  args:
    - name: message
      prompt: Commit message
      example: update aoo
  tags: [aoo, git, push]

- desc: notes git pull
  template: git -C {{aoo_notes_dir}} pull --rebase
  args: []
  tags: [notes, git, pull]

- desc: notes git add commit push
  template: git -C {{aoo_notes_dir}} add . && git -C {{aoo_notes_dir}} commit -m "{{message}}" && git -C {{aoo_notes_dir}} push
  args:
    - name: message
      prompt: Commit message
      example: update notes
  tags: [notes, git, push]

- desc: ssh to host
  template: ssh {{user}}@{{host}}
  args:
    - name: user
      prompt: SSH user
      example: root
    - name: host
      prompt: Host or IP
      example: 192.168.1.10
  tags: [ssh]
`

func Load() notes.LoadResult {
	service := notes.LoadBytes("builtin/service.yaml", []byte(serviceYAML))
	templates := notes.LoadBytes("builtin/command_templates.yaml", []byte(templatesYAML))

	return notes.LoadResult{
		Entries: append(service.Entries, templates.Entries...),
		Errors:  append(service.Errors, templates.Errors...),
	}
}
