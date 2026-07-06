# f

Терминальная утилита для заметок и запуска команд. Раньше проект назывался `aoo`; `aoo` всё ещё ставится как совместимый symlink.

## Установка

```bash
make build
sudo make install
```

## Конфиг

Файл: `~/.config/aoo/config.yaml`

Минимальный пример:

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

Что значат ключи:

- `focus_mode` скрывает нижнюю строку с хоткеями для более спокойного интерфейса
- `show_match_context` показывает строку с найденным фрагментом у выбранной записи
- `show_list_on_start` сразу показывает список при пустом запросе
- `two_line_results` включает двухстрочный список; если `false`, `desc` и команда/текст идут в одной строке

Проверка:

```bash
f config
f config show
```

## Использование

```bash
f
f --query ssh
f validate --dir ~/.local/share/aoo/notes
f add "router dhcp"
f add cmd "restart nginx"
f upgrade
```

Горячие клавиши:

- `Enter` открыть или выполнить
- `Ctrl+Enter` вывести команду в терминал без запуска, чтобы скопировать
- `Ctrl+E` редактировать запись
- `Ctrl+N` создать новую запись из текущего запроса
- `Esc` выйти

## Формат заметок

Лучше хранить одну заметку в одном файле, например `ssh-router.yaml`:

```yaml
desc: ssh router
actions:
  - desc: основной доступ
    text: ssh-команда для роутера
  - desc: подключиться
    cmd: ssh admin@router
```

Простой файл с одной командой тоже валиден:

```yaml
desc: nginx logs
cmd: journalctl -u nginx -n 100
```

Если у записи есть `cmd`, `Enter` запускает команду. Если есть только `text`, `Enter` показывает заметку.

Обычный ввод ищет заметки и файлы. `:query` или `>query` ищет только команды.

## Raw-файлы

Можно хранить сниппеты как обычные файлы: `netplan.yaml`, `bgp.conf`, `notes.md`.

`aoo` покажет их в поиске и откроет содержимое как заметку.
