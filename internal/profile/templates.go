package profile

// Embedded profile templates.
// These are role-based (no hardcoded agents).
// Each profile includes a ## Meta section for autodiscovery.

const businessFeature = `# Profile: Business Feature

## Meta
- **Keywords:** фича, добавить, реализовать, новый экран, интеграция, API endpoint, feature, implement
- **Description:** Новая функциональность, доработка, интеграция

## Workflow (STRICT)

### Stages
1. **Research** — consilium analyzes task, codebase, dependencies
2. **Plan** — implementation plan
3. **Executing** — write code
4. **Validation** — tests, review, build
5. **Report** — summary
6. **Done**

### Allowed transitions
` + "```" + `
Research   -> Plan
Research   -> Executing
Plan       -> Executing
Executing  -> Validation
Executing  -> Research
Validation -> Report
Validation -> Executing
Validation -> Research
Report     -> Done
` + "```" + `

### Agents per stage

| Stage      | Agents                       | Model  |
|------------|------------------------------|--------|
| Research   | CONSILIUM (see below)        | opus   |
| Plan       | Plan                         | opus   |
| Executing  | exec agents by file scope    | opus   |
| Validation | Bash + playwright-cli        | sonnet |
| Report     | general-purpose              | haiku  |
| Done       | —                            | —      |

### Research — Agent consilium

All agents run **in parallel** via Task tool. Each role resolves to a concrete agent through project CLAUDE.md.

| Role          | Responsibility                   |
|---------------|----------------------------------|
| ` + "`architect`" + `   | Architecture, modules, deps      |
| ` + "`frontend`" + `    | UI/UX, frontend patterns         |
| ` + "`ui`" + `          | Visual design, UX, components    |
| ` + "`security`" + `    | OWASP, vulnerabilities, auth     |
| ` + "`devops`" + `      | Infra, CI/CD, deployment         |
| ` + "`api`" + `         | API contracts, REST/GraphQL      |

### Executing

Agent determined by affected files via exec table in project CLAUDE.md.
If task touches multiple layers — run multiple exec agents in parallel.

### Report contents
- Feature name and date
- Task description
- Research summary (consilium)
- Plan
- What was implemented (files, modules)
- Validation results
- Issues and rollbacks
- Status: Done / Partial
`

const bugHunting = `# Profile: Bug Hunting

## Meta
- **Keywords:** баг, ошибка, краш, не работает, ломается, исключение, stacktrace, NPE, 500, regression, bug, fix
- **Description:** Баг, регрессия, краш, неожиданное поведение

## Workflow (STRICT)

### Stages
1. **Reproduce** — reproduce the bug
2. **Diagnose** — consilium finds root cause
3. **Fix** — write fix
4. **Smoke Test** — run smoke scenarios (mandatory for mobile changes)
5. **Validation** — verify fix, check regressions
6. **Report** — summary
7. **Done**

### Allowed transitions
` + "```" + `
Reproduce  -> Diagnose
Reproduce  -> Report         (not reproducible)
Diagnose   -> Fix
Diagnose   -> Reproduce
Diagnose   -> Report         (diagnosis only)
Fix        -> Smoke Test     (mobile changes)
Fix        -> Validation     (backend only)
Fix        -> Diagnose
Smoke Test -> Validation
Smoke Test -> Fix
Validation -> Report
Validation -> Fix
Validation -> Diagnose
Report     -> Done
` + "```" + `

### Diagnose — Agent consilium

| Role           | Responsibility                        |
|----------------|---------------------------------------|
| ` + "`diagnostics`" + `  | Logs, stacktraces, instrumentation    |
| ` + "`architect`" + `    | Architectural causes, module deps     |
| ` + "`security`" + `     | Vulnerabilities, leaks, auth issues   |
| ` + "`devops`" + `       | Infra, environment, configs           |

### Fix
Agent determined by affected files via exec table in project CLAUDE.md.
`

const research = `# Profile: Research

## Meta
- **Keywords:** как устроено, как работает, как реализовано, объясни, расскажи, что такое, исследуй, покажи архитектуру, explore
- **Description:** Понять как что-то устроено, не планируя делать изменения

## Workflow
1. **Research** — consilium investigates topic in parallel
2. **Done** — structured answer in chat

No Plan, no Executing, no code changes.

### Consilium roles
| Role          | Responsibility                   |
|---------------|----------------------------------|
| ` + "`architect`" + `   | Architecture, modules, deps      |
| ` + "`frontend`" + `    | UI/UX, frontend patterns         |
| ` + "`ui`" + `          | Visual design, UX                |
| ` + "`security`" + `    | OWASP, vulnerabilities           |
| ` + "`devops`" + `      | Infra, CI/CD, deployment         |
| ` + "`api`" + `         | API contracts                    |
`

const refactoring = `# Profile: Refactoring

## Meta
- **Keywords:** рефакторинг, refactor, почисти код, аудит кода, code review, clean up
- **Description:** Улучшить качество существующего кода без смены функционала

## Workflow (STRICT)

### Stages
1. **Audit** — consilium finds all violations
2. **Plan** — structured fix list, user approves
3. **Executing** — refactor by approved plan
4. **Validation** — regression consilium + tester
5. **Report**
6. **Done**

### Audit consilium
| Role           | What to find                             |
|----------------|------------------------------------------|
| ` + "`architect`" + `    | SOLID violations, god classes, cycles    |
| ` + "`diagnostics`" + `  | Duplication, dead code, code smells      |
| ` + "`security`" + `     | OWASP, injection, secret leaks           |
| ` + "`frontend`" + `     | Component duplication, state issues      |
| ` + "`api`" + `          | REST inconsistencies, missing validation |
| ` + "`test`" + `         | Missing tests, weak coverage             |

### Validation — regression consilium
| Role           | What to check                    |
|----------------|----------------------------------|
| ` + "`diagnostics`" + `  | New bugs, regressions, NPE       |
| ` + "`architect`" + `    | Architectural degradation        |
| ` + "`security`" + `     | New vulnerabilities               |
`

const e2eTesting = `# Profile: E2E Testing

## Meta
- **Keywords:** e2e, тестирование, прогнать тесты, smoke, прогнать сценарии, проверить платформы, запустить e2e, протестируй
- **Description:** Прогон smoke/e2e сценариев, автофикс найденных проблем

## Workflow (STRICT)

### Stages
1. **Prepare** — find scenarios in .specs/e2e/
2. **Deploy** — deploy to prod
3. **Run** — run scenarios on selected platforms
4. **Fix** — auto-fix found issues (exec agents by file scope)
5. **Re-run** — re-run failed scenarios (max 3 cycles)
6. **Report**
7. **Done**

### Platform tools (fixed)
| Platform | Tool                        |
|----------|-----------------------------|
| Web      | playwright-cli              |
| Android  | /test-android skill         |
| iOS      | /test-ios skill             |
| Desktop  | /test-desktop skill         |
| Backend  | curl / bash                 |

Fix agents resolve through project CLAUDE.md exec table.
`

const e2eAuthoring = `# Profile: E2E Scenario Authoring

## Meta
- **Keywords:** добавить e2e тест, завести smoke, новый сценарий, создать тест-кейс, написать e2e
- **Description:** Исследование кода и создание нового сценария для прогонов

## Workflow (STRICT)

### Stages
1. **Research** — explore codebase (Explore agent)
2. **Propose** — suggest scenarios to user
3. **Approve** — user refines steps
4. **Save** — write .specs/e2e/ file
5. **Done**

No code changes, no deployment, no test execution.
Scenarios stored in .specs/e2e/ (committed to git).
`
