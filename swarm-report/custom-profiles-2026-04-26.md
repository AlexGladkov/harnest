# Report: Custom Profiles

**Дата:** 2026-04-26
**Статус:** Done

## Задача
Добавить возможность создавать/редактировать кастомные профили через CLI, защитить пользовательские изменения при install, перевести автодетект на autodiscovery.

## Что реализовано

### Новые команды
- `harnest profiles add <name>` — интерактивный визард создания кастомного профиля
- `harnest profiles edit <name>` — редактирование через $EDITOR

### Визард (profiles add)
- Ввод стадий по одной: имя → тип агента (single/consilium/bash/none) → роли → переходы
- Генерация markdown с ## Meta (Keywords, Description), Workflow, Transitions, Agents per stage
- Atomic write через temp + rename

### Конфликт-резолюция (install)
- При `harnest install` — если builtin профиль изменён локально, спрашивает: overwrite/skip/diff
- diff показывает unified diff через внешний `diff -u`
- Кастомные профили не затрагиваются

### Autodiscovery
- Global CLAUDE.md больше не содержит хардкод таблицу 6 профилей
- Вместо этого: инструкция "сканируй ~/.claude/profiles/*.md, читай ## Meta"
- Каждый профиль (включая все 6 builtin) содержит ## Meta с Keywords и Description

### Security fixes
- `ValidateName()` — regex `^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`
- `safePath()` — path traversal protection через filepath.Abs + prefix check
- `profilesDir()` возвращает error (раньше игнорировал UserHomeDir ошибку)
- File permissions: 0700 для директорий, 0600 для файлов
- $EDITOR через exec.Command напрямую, без shell (защита от injection)
- Убран --dest флаг (позволял запись в произвольные директории)

### profiles list
- Показывает метку `(builtin)` рядом с встроенными профилями

## Затронутые файлы

| Файл | Изменение |
|------|-----------|
| `internal/profile/profile.go` | Полная переработка: ValidateName, safePath, profilesDir с error, Create (визард), Edit, IsModified, BuiltinNames, IsBuiltin, BuiltinContent, atomic write |
| `internal/profile/templates.go` | Добавлена ## Meta секция во все 6 builtin профилей |
| `internal/install/install.go` | Конфликт-резолюция, BuiltinNames() вместо дубля, promptConflict, showDiff |
| `internal/install/global_template.go` | Хардкод таблица → autodiscovery инструкция |
| `cmd/harnest/main.go` | Новый case edit, переработанный add, обновлённый usage |

## Validation
- `go build ./...` — чисто
- `go vet` по затронутым пакетам — чисто
- `profiles list` — работает, показывает builtin метки
- `profiles add` с существующим именем — корректная ошибка
- Path traversal (`../../../etc/evil`) — заблокирован ValidateName

## Breaking changes
- `profiles add <name>` теперь запускает визард вместо установки builtin. Builtin профили ставятся через `harnest install`.
- Убран флаг `--dest` из `profiles add`.
