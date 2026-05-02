# aoo

Быстрая терминальная утилита для заметок и запуска команд.

Храни заметки, команды, хосты и сниппеты в YAML. Используй fuzzy search по содержимому заметок, находи нужное за секунды и запускай команды прямо из заметок.

![aoo demo](docs/demo.gif)

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
aoo config show
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
Enter   открыть / запустить заметку
Alt+E   открыть исходный YAML в $VISUAL / $EDITOR
Esc     выйти
Left/Right переключить ALL / RUN / SHOW
Up/Down движение по списку
```

## Формат заметок

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

Режимы:

- `mode: run` запускает команду или template-команду
- `mode: show` печатает текст

Поведение:

- `run` запускается после подтверждения
- `template` спрашивает аргументы, показывает итоговую команду и просит подтверждение
- `note` печатает текст

Встроенные переменные шаблонов:

- `{{aoo_notes_dir}}`
- `{{aoo_app_dir}}`

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
- `docs/demo.gif` короткое терминальное демо
- `docs/demo.tape` VHS-исходник для демо

## Релизы

- Тег вида `v0.2.0` запускает autobuild GitHub Release
- Release workflow публикует tar.gz и `.deb` пакеты
