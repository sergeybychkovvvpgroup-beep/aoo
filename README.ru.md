# aoo

Терминальная утилита для заметок и запуска команд.

## Установка

```bash
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
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
aoo config
aoo config show
```

## Использование

```bash
aoo
aoo --query ssh
aoo validate --dir ~/.local/share/aoo/notes
aoo add "router dhcp"
aoo add cmd "restart nginx"
```

Горячие клавиши:

- `Enter` открыть или выполнить
- `Ctrl+E` редактировать запись
- `Ctrl+N` создать новую запись из текущего запроса
- `Esc` выйти

## Формат заметок

```yaml
- desc: ssh router
  text: основной доступ
  cmd: ssh admin@router

- desc: nginx logs
  cmd: journalctl -u nginx -n 100
```

Если у записи есть `cmd`, `Enter` запускает команду. Если есть только `text`, `Enter` показывает заметку.

Обычный ввод ищет заметки и файлы. `:query` или `>query` ищет только команды.

## Raw-файлы

Можно хранить сниппеты как обычные файлы: `netplan.yaml`, `bgp.conf`, `notes.md`.

`aoo` покажет их в поиске и откроет содержимое как заметку.
