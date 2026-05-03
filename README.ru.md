# aoo

Быстрая терминальная утилита для заметок и запуска команд.

Храни заметки, команды, хосты и сниппеты в YAML. Используй fuzzy search по содержимому заметок, находи нужное за секунды и запускай команды прямо из заметок.

![aoo demo](docs/demo.png)

## Установка

```bash
git clone https://github.com/sergeybychkovvvpgroup-beep/aoo.git
cd aoo
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
```

Установка `.deb` вручную:

```bash
sudo dpkg -i ./aoo_<version>_amd64.deb
```

## Первый запуск

При первом запуске `aoo` может сам спросить URL notes repo и склонировать его в `~/.local/share/aoo/notes`.

Основные настройки теперь живут в одном файле:

```bash
~/.config/aoo/config.yaml
```

Файл создаётся автоматически с комментариями и примерами.

Либо можно настроить вручную:

```bash
aoo set-source
mkdir -p ~/.local/share/aoo/notes
aoo set-folder ~/.local/share/aoo/notes
aoo set-app-dir ~/workspace/aoo
```

Если notes repo приватный, `aoo` покажет public key, который нужно добавить в Deploy Keys.
`aoo` не делает auto-fetch notes repo на каждом старте, поэтому не должен постоянно дёргать passphrase от SSH key.

Проверить конфиг:

```bash
aoo config
aoo config show
```

Полезные UI-настройки в `config.yaml`:

```yaml
full_screen: true
picker_height: 14
show_preview: false
preview_pane: false
```

## Использование

```bash
aoo
aoo --query ssh
aoo --query nmap
aoo validate --dir ~/.local/share/aoo/notes
```

Темы:

```bash
aoo themes
aoo set-theme catppuccin-mocha
aoo --theme nord
```

Горячие клавиши в picker:

```text
Enter   открыть действия для заметки
Alt+E   открыть исходный YAML в $VISUAL / $EDITOR
Esc     выйти
Left/Right переключить фильтр по тегу
Up/Down движение по списку
```

## Формат заметок

```yaml
- desc: example ssh
  actions:
    - desc: show
      text: основной доступ к хосту
    - desc: ssh
      cmd: ssh admin@example.local

- desc: example url
  actions:
    - desc: show
      text: https://service.example.local

- desc: example nmap template
  tags: [nmap, scan]
  actions:
    - desc: nmap
      template: sudo nmap -O {{host}}
      args:
        - name: host
          prompt: Host or IP
          example: 192.168.1.1
```

Поведение:

- `Enter` всегда открывает список действий для заметки
- `actions` это основной формат
- `text` показывает текст
- `cmd` запускается после подтверждения
- `template` спрашивает аргументы, показывает итоговую команду и просит подтверждение
- `full_screen: true` включает alt-screen и использует всю высоту терминала
- когда `full_screen: false`, высота picker ограничивается через `picker_height`, поэтому видно терминал до запуска `aoo`
- `show_preview: false` делает каждый результат поиска однострочным
- `preview_pane: true` показывает правую preview-панель как в `fzf`
- у каждого действия должен быть свой `desc`

Несколько действий в одной заметке:

```yaml
- desc: grep
  actions:
    - desc: show
      text: частые варианты grep
    - desc: grep in syslog
      cmd: grep -i error /var/log/syslog
    - desc: grep with context
      cmd: grep -nC 3 error /var/log/syslog

- desc: himki
  actions:
    - desc: show
      text: основной доступ к площадке
    - desc: ssh vyos
      cmd: ssh vyos@himki
    - desc: ssh tunnel
      cmd: ssh -L 8443:10.10.10.1:443 vyos@himki
      banner: https://127.0.0.1:8443
```

Встроенные переменные шаблонов:

- `{{aoo_notes_dir}}`
- `{{aoo_app_dir}}`
- `{{aoo_config_file}}`

Пример команды для конфига:

```yaml
- desc: edit aoo config
  actions:
    - desc: open
      cmd: $EDITOR {{aoo_config_file}}
```

## Темы

- `catppuccin-mocha`
- `catppuccin-latte`
- `dracula`
- `nord`
- `solarized-dark`
- `solarized-light`

## Структура репозитория

- `examples/notes/` безопасные example notes для демо и валидации
- `cmd/`, `internal/` код утилиты
- `internal/bundled/` встроенные help и starter notes

## Релизы

- Тег вида `v0.2.0` запускает autobuild GitHub Release
- Release workflow публикует tar.gz и `.deb` пакеты
