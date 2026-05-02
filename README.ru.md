# aoo

Быстрая терминальная утилита для заметок и запуска команд из YAML-файлов.

`aoo` сканирует каталог с заметками, даёт быстрый поиск по нескольким словам, показывает текстовые заметки и запускает команды прямо из терминала. Утилита ориентирована на инфраструктурные и support-сценарии: SSH, туннели, URL, сниппеты, jump-host, локальные подсказки.

## Возможности

- Поиск по нескольким словам с неточным совпадением
- Один быстрый Go-бинарь без runtime-зависимостей после сборки
- Поле `banner` перед выполнением команды
- Поле `check` перед выполнением команды
- Режим `validate` для проверки YAML
- Поддержка заметок `note` и команд `run`
- Поддержка шаблонов команд `template`
- Автоматический поиск по `desc`, имени файла, тегам, команде, тексту заметки и баннеру

## Формат заметок

```yaml
- desc: proxmox chashnikovo external tunnel localhost 8006
  tags:
    - proxmox
    - tunnel
  banner: |
    Open in browser:
    https://127.0.0.1:8006
  check: ssh -o BatchMode=yes -o ConnectTimeout=5 sergeyb@178.238.119.166 exit
  check_error: |
    SSH auth check failed.
    Verify that your key is loaded.
  run: ssh -L 8006:10.117.0.1:8006 sergeyb@178.238.119.166

- desc: proxmox chashnikovo web local
  note: |
    https://10.117.0.0.1:8006
```

Поддерживаемые поля:

- `desc` — обязательное короткое описание
- `run` — команда для выполнения
- `template` — шаблон команды с аргументами
- `args` — список аргументов для `template`
- `note` — текст заметки
- `banner` — текстовый баннер перед `run`
- `tags` — дополнительные теги для поиска

У каждой записи должен быть `desc` и хотя бы одно из полей: `run` или `note`.

## Требования

- Linux-терминал с интерактивным TTY
- `/bin/sh` для выполнения `run:` команд
- каталог с файлами `.yaml` или `.yml`
- Go `1.22+`, только если собираешь из исходников

Структура репозитория:

- `cmd/`, `internal/`: код утилиты
- `notes/`: локальные или примерные YAML-заметки

## Установка зависимостей

Источники:

- Go: https://go.dev/doc/install
- GoReleaser: https://goreleaser.com/getting-started/install/

Ubuntu / Debian:

```bash
sudo apt update
sudo apt install -y golang-go
```

Опционально, для сборки релизов:

```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

## Сборка

```bash
git clone https://github.com/sergeyb/aoo.git
cd aoo
make tidy
make build
```

Бинарь появится в `bin/aoo`.

## Установка

Локальная установка из исходников:

```bash
git clone https://github.com/sergeyb/aoo.git
cd aoo
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
```

Установка через `go install`:

```bash
go install github.com/sergeyb/aoo/cmd/aoo@latest
```

Установка из `.deb` пакета:

```bash
sudo dpkg -i aoo_<version>_linux_amd64.deb
```

Проверка:

```bash
aoo version
```

## Использование

Порядок определения каталога с заметками:

1. `--dir`
2. `AOO_NOTES_DIR`
3. `TERM_NOTES_DIR` для обратной совместимости
4. `~/.config/aoo/config.yaml`
5. текущий каталог, если в нём есть `*.yaml`

Первичная настройка:

```bash
aoo set-folder ~/notes
```

Посмотреть текущую конфигурацию:

```bash
aoo config show
```

Внутри этого репозитория заметки лежат в `./notes`.

Запуск интерактивного поиска:

```bash
aoo
```

Запуск с предзаполненным запросом:

```bash
aoo --query "chash router"
```

Шаблонные команды:

```bash
aoo --query nmap
```

После выбора `template` утилита:
- спросит значения аргументов
- покажет итоговую команду
- попросит подтверждение перед запуском

Проверка YAML-заметок:

```bash
aoo validate
```

Пример pre-check:

- `check`: shell-команда, которая должна завершиться с кодом `0` до запуска `run`
- `check_error`: понятное сообщение, которое показывается при провале `check`

## Как работает поиск

Поиск сделан практично, а не строго по полному совпадению:

- каждое слово запроса должно где-то найтись
- полные подстроки ранжируются выше
- неточные subsequence-совпадения тоже учитываются
- токены из имени файла автоматически участвуют в поиске

Пример: запрос `chash router` может найти заметку из файла `chashnikovo-wh2.yaml`, даже если `chash` есть только как часть имени файла.

## Публикация в GitHub

Базовая публикация исходников:

```bash
git add .
git commit -m "Initial Go version of aoo"
git branch -M main
git remote add origin git@github.com:YOUR_USERNAME/aoo.git
git push -u origin main
```

Релиз по тегу:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Что уже подготовлено в репозитории

- `Makefile` для типовых действий
- `.goreleaser.yaml` для сборки архивов и `.deb`
- GitHub Actions CI
- валидация YAML-заметок при сборке

## Следующий практичный шаг

Если нужна полноценная публикация бинарей в GitHub Releases, добавь `release.yml`, который будет запускать GoReleaser по тегу. Если нужен именно `apt install`, потребуется отдельный APT-репозиторий или внешний сервис публикации пакетов.
