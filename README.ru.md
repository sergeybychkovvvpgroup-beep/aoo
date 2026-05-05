# aoo

Быстрая терминальная утилита для заметок и запуска команд.

Храни команды в YAML, а реальные сниппеты и конфиги как обычные файлы. Используй fuzzy search по содержимому заметок и файлов, находи нужное за секунды и запускай команды прямо из `aoo`.

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
Если `notes_dir` это git repo, `aoo` сам делает `fetch/pull/push` при старте и после редактирования заметок.
Если SSH key с passphrase, лучше держать его в `ssh-agent`.

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
Enter   открыть или выполнить
Ctrl+E  редактировать YAML
Ctrl+N  создать новую заметку из текущего запроса
Esc     выйти
```

Быстрое добавление:

```bash
aoo add "router dhcp"
aoo add cmd "restart nginx"
```

Обе команды создают YAML-скелет, открывают его в `$VISUAL` / `$EDITOR`, а после выхода из редактора автоматически делают commit/pull/push, если `notes_dir` это git repo.

## Формат заметок

`aoo` поддерживает два простых формата:

- YAML-записи для команд и launcher-заметок
- raw-файлы для реальных сниппетов и конфигов

### YAML-записи

```yaml
- desc: example ssh
  text: основной доступ к хосту
  cmd: ssh admin@example.local

- desc: example url
  text: https://service.example.local

- desc: example command
  cmd: dig nas.example.local A +short
```

Поведение:

- `Enter` сразу делает главное действие
- если у записи есть `cmd`, то `Enter` запускает его
- если у записи только `text`, то `Enter` показывает заметку
- старый формат `actions` остаётся только для совместимости
- `text` показывает текст
- `cmd` запускается сразу после выбора
- `full_screen: true` включает alt-screen и использует всю высоту терминала
- когда `full_screen: false`, высота picker ограничивается через `picker_height`, поэтому видно терминал до запуска `aoo`
- обычный ввод ищет только заметки и файлы
- `:query` или `>query` показывает только команды
- `show_preview: false` делает каждый результат поиска однострочным
- `preview_pane: true` показывает правую preview-панель как в `fzf`
- старайся держать одну запись = одна основная задача

### Raw-файлы

Для сниппетов не обязательно использовать YAML-обёртку. Можно положить реальный файл как есть:

- `netplan-prod.yaml`
- `router-bgp.conf`
- `deploy.py`
- `notes.md`

Пример `netplan-prod.yaml`:

```yaml
network:
  version: 2
  ethernets:
    ens18:
      dhcp4: true
```

Поведение raw-файлов:

- если `*.yaml` / `*.yml` не содержит поля `aoo` (`desc`, `cmd`, `text`, `run` и т.д.), файл считается raw-сниппетом
- raw-файл показывается в поиске как отдельная заметка
- `desc` строится из имени файла
- `Enter` печатает содержимое как есть
- в action label и hint показывается расширение файла: `yaml`, `py`, `conf`; если расширения нет, используется `raw`

Пример команды для конфига:

```yaml
- desc: edit aoo config
  cmd: $EDITOR ~/.config/aoo/config.yaml
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
