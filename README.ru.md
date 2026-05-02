# aoo

Быстрая терминальная утилита для заметок и запуска команд для self-hosted инфраструктуры.

Ищет YAML-заметки, показывает сниппеты, запускает сохранённые команды и умеет шаблоны команд с вопросами по аргументам.

## Установка

```bash
git clone https://github.com/sergeyb/aoo.git
cd aoo
make build
sudo install -m 0755 bin/aoo /usr/local/bin/aoo
```

## Первый запуск

Храни реальные заметки вне репозитория:

```bash
mkdir -p ~/.local/share/aoo/notes
aoo set-folder ~/.local/share/aoo/notes
```

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

## Формат заметок

```yaml
- desc: example ssh
  run: ssh admin@example.local

- desc: example url
  note: |
    https://service.example.local

- desc: example nmap template
  template: sudo nmap -O {{host}}
  args:
    - name: host
      prompt: Host or IP
      example: 192.168.1.1
```

`run` запускает команду, `note` печатает текст, `template` спрашивает аргументы и собирает итоговую команду.

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
