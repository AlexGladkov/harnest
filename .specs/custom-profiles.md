# Spec: Custom Profiles

## Summary

Добавить возможность создавать кастомные профили через интерактивный визард, редактировать существующие профили через `$EDITOR`, защитить пользовательские изменения при `harnest install`, перевести автодетект профилей на autodiscovery (сканирование `~/.claude/profiles/`).

## Решения из интервью

| Вопрос | Решение |
|--------|---------|
| Создание профиля | `harnest profiles add <name>` — интерактивный визард |
| Визард стадий | По одной стадии: name → agent type (single/consilium/bash/none) → roles → transitions → next? |
| Transitions | Спрашивать для каждой стадии: "After <stage>, where can you go?" |
| Builtin профили через add | Убрать. Builtin ставятся через `harnest install` |
| Редактирование | `harnest profiles edit <name>` — открывает `$EDITOR` (fallback: vim) |
| Защита при install | Сравнивать контент файла с builtin шаблоном. При различии — спрашивать: overwrite / skip / diff |
| Автодетект | Global CLAUDE.md говорит AI "сканируй ~/.claude/profiles/". Убрать хардкод таблицу |
| Keywords формат | `## Meta` секция в каждом .md профиле: `**Keywords:** ...`, `**Description:** ...` |
| Builtin meta | Добавить `## Meta` во все 6 builtin профилей |
| Команды | `add`, `edit`, `remove`, `list` — достаточно |

## Затронутые файлы

### Изменения

| Файл | Что меняется |
|------|-------------|
| `internal/profile/profile.go` | Новые функции: `Create()` (визард), `Edit()` ($EDITOR), `IsModified()` (сравнение с builtin). Изменение `Install()` — проверка конфликтов |
| `internal/profile/templates.go` | Добавить `## Meta` секцию во все 6 builtin профилей. Добавить scaffold-шаблон для новых профилей |
| `cmd/harnest/main.go` | Новые subcommands: `profiles add` (визард), `profiles edit`. Изменить `profiles add` — убрать прямую установку builtin |
| `internal/install/install.go` | `InstallAll()` — перед записью проверять файл на диске, если отличается — интерактивный prompt |
| `internal/install/global_template.go` | Убрать хардкод таблицу профилей, заменить на autodiscovery инструкцию |

### Новые файлы

Нет новых файлов. Вся логика добавляется в существующие.

## Детальный дизайн

### 1. `harnest profiles add <name>` — интерактивный визард

```
$ harnest profiles add release

Creating profile: release

--- Stage 1 ---
Stage name: > Prepare
Agent type (single/consilium/bash/none): > consilium
Roles (comma-separated): > architect, devops
After Prepare, allowed transitions (comma-separated stage names): > Deploy

Add another stage? (y/n): > y

--- Stage 2 ---
Stage name: > Deploy
Agent type: > bash
After Deploy, allowed transitions: > Verify

Add another stage? (y/n): > y

--- Stage 3 ---
Stage name: > Verify
Agent type: > consilium
Roles: > diagnostics, security
After Verify, allowed transitions: > Report, Deploy

Add another stage? (y/n): > y

--- Stage 4 ---
Stage name: > Report
Agent type: > single
Agent name (or Enter for 'general-purpose'): >
After Report, allowed transitions: > Done

Add another stage? (y/n): > y

--- Stage 5 ---
Stage name: > Done
Agent type: > none
After Done, allowed transitions: >

Add another stage? (y/n): > n

Keywords (comma-separated): > release, deploy, выкатка, релиз, выпуск
Description: > Релизный цикл: подготовка, деплой, верификация

Profile 'release' created.
  → ~/.claude/profiles/release.md
```

### 2. Генерируемый файл

```markdown
# Profile: Release

## Meta
- **Keywords:** release, deploy, выкатка, релиз, выпуск
- **Description:** Релизный цикл: подготовка, деплой, верификация

## Workflow (STRICT)

### Stages
1. **Prepare** — consilium analyzes task
2. **Deploy** — bash execution
3. **Verify** — consilium verifies result
4. **Report** — summary
5. **Done**

### Allowed transitions
```
Prepare -> Deploy
Deploy  -> Verify
Verify  -> Report
Verify  -> Deploy
Report  -> Done
```

### Agents per stage

| Stage   | Agents                    | Model  |
|---------|---------------------------|--------|
| Prepare | CONSILIUM (see below)     | opus   |
| Deploy  | Bash                      | sonnet |
| Verify  | CONSILIUM (see below)     | opus   |
| Report  | general-purpose           | haiku  |
| Done    | —                         | —      |

### Prepare — Agent consilium

| Role       | Responsibility                |
|------------|-------------------------------|
| `architect` | Architecture, modules, deps  |
| `devops`    | Infra, CI/CD, deployment     |

### Verify — Agent consilium

| Role          | Responsibility                    |
|---------------|-----------------------------------|
| `diagnostics` | Logs, stacktraces, instrumentation |
| `security`    | OWASP, vulnerabilities, auth      |
```

### 3. `harnest profiles edit <name>`

```go
func Edit(name string) error {
    path := filepath.Join(profilesDir(), name+".md")
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Errorf("profile not found: %s", name)
    }
    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = "vim"
    }
    cmd := exec.Command(editor, path)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

### 4. `harnest install` — конфликт-резолюция

```
Installing profiles...
  → ~/.claude/profiles/business-feature.md
  → ~/.claude/profiles/bug-hunting.md (modified locally)
    [o]verwrite  [s]kip  [d]iff: > d

--- diff ---
-  Old line
+  New line
---

    [o]verwrite  [s]kip: > s
  → skipped bug-hunting.md
  → ~/.claude/profiles/research.md
  ...
```

Логика:
1. Для каждого builtin профиля: прочитать файл с диска
2. Если файл не существует — записать
3. Если контент совпадает с builtin — записать (обновление)
4. Если контент отличается — спросить: overwrite / skip / diff
5. diff показывает разницу и спрашивает снова: overwrite / skip
6. Кастомные профили (нет в builtinProfiles map) — не трогать

### 5. Global CLAUDE.md — autodiscovery

Заменить хардкод таблицу:

```
| Профиль          | Файл                                          | Когда использовать |
|------------------|-----------------------------------------------|-------------------|
| Бизнес-фича      | ~/.claude/profiles/business-feature.md        | ...               |
...
```

На:

```
### Доступные профили

Профили хранятся в `~/.claude/profiles/*.md`. Для определения профиля:
1. Прочитать все `.md` файлы в `~/.claude/profiles/`
2. В каждом найти секцию `## Meta` с полями **Keywords** и **Description**
3. Матчить ключевые слова из запроса пользователя с Keywords профилей
4. Если совпадение найдено — предложить этот профиль
5. Если совпадений нет или несколько — показать список и спросить пользователя
```

### 6. Builtin Meta — добавить во все 6 профилей

Пример для business-feature:

```markdown
# Profile: Business Feature

## Meta
- **Keywords:** фича, добавить, реализовать, новый экран, интеграция, API endpoint, feature, implement
- **Description:** Новая функциональность, доработка, интеграция

## Workflow (STRICT)
...
```

## Edge Cases

1. `profiles add` с именем существующего профиля → ошибка: "profile already exists. Use 'harnest profiles edit <name>'"
2. `profiles edit` несуществующего → ошибка: "profile not found"
3. `profiles remove` builtin → работает (удаляет файл, можно восстановить через install)
4. Пустые keywords → профиль не матчится автодетектом, только ручной выбор
5. `$EDITOR` не установлен → fallback на `vim`
6. Визард: пустое имя стадии → пропустить, спросить снова
7. Визард: transition на несуществующую стадию → warning, но разрешить (стадия может быть добавлена позже)

## Обратная совместимость

- `profiles list` — без изменений (уже сканирует директорию)
- `profiles remove` — без изменений
- `profiles add <builtin-name>` — поведение меняется: раньше устанавливал builtin, теперь → ошибка "profile already exists" если файл есть, или создаст кастомный профиль с таким именем если файла нет. **BREAKING** но приемлемо — builtin ставятся через install.
