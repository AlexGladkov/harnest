# Report: Agent Discovery Fixes & Language Maps Expansion

**Дата:** 2026-06-26
**Статус:** Done
**Версия:** 0.11.0 → 0.12.0

## Задача

1. Плагиновые агенты не находились — баг в `scanPlugins`
2. Поиск проектных агентов для всех harness dirs
3. Языковые карты только для Kotlin — расширить на 8 языков
4. Scope `**/` съедался в `buildScopeFromKeywords`
5. Detector: `.slnx`, multi-project, глубокая вложенность

## Что реализовано

### Agent Discovery

**Файл:** `internal/agents/discover.go`

- `Discover(projectDir)` — единая точка входа: проектные (приоритет) → глобальные → плагины
- Порядок: insertion order, без сортировки. Проектные всегда первые → выигрывают tie в `MatchAgent`
- Проектные агенты: все 6 harness dirs (`.claude/agents/`, `.cursor/agents/`, `.windsurf/agents/`, `.codex/agents/`, `.config/opencode/agents/`, `.qwen/agents/`)
- YAML frontmatter: `name:` → приоритет, нет `name:` → fallback на имя файла, нет frontmatter → fallback на имя файла
- Защита: size limit (64KB), null-byte детект, frontmatter limit (4KB)

**Баг scanPlugins:**
- Плагины не находились — код ждал поле `agents` в `plugin.json`, но современные плагины кладут агентов в `agents/*.md`
- Фикс: всегда сканируется `<version>/agents/*.md`. Поле `agents` — обратная совместимость
- Namespace: `plugin-name:agent-name`
- Результат: 14 агентов (11 ai-sdlc + 3 caveman), было 0

**Wizard fix:**
- `wizard.Run()` принимает `available []string` извне — больше не вызывает `Discover()` внутри себя
- Проектные агенты попадают в wizard

### Mapping

**Файл:** `internal/mapping/mapping.go`

**Языковые карты:**
- `securityKeywords`, `diagnosticsKeywords`, `testKeywords`: было только `kotlin`, стало 8 языков
- `defaultRoleKeywords[devops]` — `devops, infra, sre, platform, deploy`
- `defaultRoleKeywords[test]` — `test, qa, tester`
- `testKeywords[*]` — `qa` во все языки

**matchRole fix:**
- Было: `kw = langKW` (замена default)
- Стало: `kw = append(defaultKW, langKW...)` (дополнение)
- Эффект: role-name агенты получают +1 bonus за совпадение с именем роли. `security` бьёт `backend-csharp` для security role

**Scope fix:**
- `buildScopeFromKeywords` для `**/*.py` + path `backend/` → `backend/**/*.py` (было `backend/*.py`)
- `**/` wildcard сохраняется

**Exec keywords:**
- `dotnet` — добавлен `csharp`
- `docker` — добавлены `devops, infra`

### Detector

**Файл:** `internal/detector/detector.go`

- **`.slnx`** — поддержка нового формата .NET решений
- **Multi-project:** `detectDotNet` больше не выходит после первого стека. Дедупликация по path
- **Multi-module:** `detectKotlin` и `detectJava` — убраны `break`/`return stacks` внутри циклов
- **Глубокая вложенность:** `subdirs` → 3 уровня (было 1)
- **Нормализация путей:** `relPath` → `filepath.ToSlash`

## Тесты

| Пакет | Тестов | Покрытие |
|-------|--------|----------|
| agents | 13 | Discover, DiscoverProject (multi-harness), ParseAgentName (11 кейсов), ScanPlugins (4) |
| mapping | 32 | MatchAgent (13), MatchRole (10), BuildScope (5), ExecKeywords (2), TieBreaker, Fallback |
| detector | 21 | DotNet (11), Kotlin (5), Java (5) |

**Всего: 66 тестов.** `go vet` чист.

## Makefile

- Версия 0.12.0
- Добавлены Windows-билды: `windows-amd64.exe`, `windows-arm64.exe`
- `clean` чистит `.exe`
- `checksum` fallback: `shasum || sha256sum`

## Затронутые файлы

```
cmd/harnest/main.go              Discover(dir), wizard.Run(available), 0.12.0
internal/agents/discover.go      Discover(dir), scanPlugins fix, parseAgentName
internal/agents/discover_test.go 13 тестов
internal/mapping/mapping.go      языковые карты, scope fix, matchRole fix, exec keywords
internal/mapping/mapping_test.go 32 теста
internal/detector/detector.go    .slnx, multi-project, subdirs 3 уровня
internal/detector/detector_test.go 21 тест
internal/wizard/wizard.go        available []string параметр
internal/converter/converter.go  Discover(dir)
internal/drift/drift.go          Discover(dir)
Makefile                         0.12.0, +Windows, clean fix
README.md                        Agent Discovery обновлён
go.mod                           yaml.v3 → прямой dependency
```
