# Harnest

AI coding assistant configurator. Detects your project stack and generates optimal agent configs for Claude Code, Cursor, Windsurf, and more.

## Problem

Every AI coding tool needs project context. Each tool has its own config format. Manually maintaining `CLAUDE.md`, `.cursorrules`, `.windsurfrules` is tedious — especially for multi-stack projects where wrong agents get assigned (Java agents reviewing Swift code).

## Solution

```bash
$ cd my-project
$ harnest init

Detected stack:
  - spring-boot (kotlin) [backend/]
  - compose-multiplatform (kotlin) [composeApp/]
  - ios-native (swift) [iosApp/]
  - vue (typescript) [vue-frontend/]

Target harness:
  > 1) claude-code
    2) cursor
    3) windsurf

Select [1]: 1

Generated: CLAUDE.md
  Consilium roles: 8
  Exec agents: 4
```

One command. Correct agents for each part of your stack.

## Install

```bash
brew tap AlexGladkov/tap
brew install harnest
```

Or with Go:

```bash
go install github.com/AlexGladkov/harnest/cmd/harnest@latest
```

Or download binary from [Releases](https://github.com/AlexGladkov/harnest/releases).

## Commands

```bash
# Detect stack without generating
harnest detect [dir]

# Generate AI assistant config (interactive)
harnest init [dir]

# Generate for specific tool
harnest init --harness cursor

# View current agent mappings
harnest agents list

# Override a role
harnest agents set architect voltagent-lang:swift-expert

# Install workflow profiles
harnest profiles list
harnest profiles add business-feature

# Convert between tools
harnest convert --from claude-code --to cursor

# Check for updates
harnest update
```

## How It Works

### Stack Detection

Harnest scans your project for build files and frameworks:

| Indicator | Detected Stack |
|-----------|---------------|
| `build.gradle.kts` + spring | Spring Boot (Kotlin) |
| `composeApp/` | Compose Multiplatform |
| `iosApp/` or `Package.swift` | iOS / Swift |
| `package.json` + vue | Vue.js |
| `package.json` + react | React |
| `package.json` + next | Next.js |
| `pubspec.yaml` | Flutter |
| `go.mod` | Go |
| `Cargo.toml` | Rust |
| `pyproject.toml` + fastapi | FastAPI |
| `pyproject.toml` + django | Django |

### Agent Role System

Generated configs use a **role-based** agent system:

**Consilium roles** (analyze code, don't write it):
- `architect` — architecture, modules, dependencies
- `frontend` — UI/UX review, frontend patterns
- `ui` — visual design, UX, components
- `security` — OWASP, vulnerabilities, auth
- `devops` — infrastructure, CI/CD, deployment
- `api` — API contracts, REST/GraphQL
- `diagnostics` — logs, stacktraces, debugging
- `test` — test coverage, quality

**Exec agents** (write code, matched by file scope):

```
backend/**/*.kt    → builder-spring-feature
composeApp/**/*.kt → kotlin-multiplatform-developer
iosApp/**/*.swift  → voltagent-lang:swift-expert
vue-frontend/**    → voltagent-lang:vue-expert
```

Each role maps to the optimal agent for your specific stack. Swift projects get Swift agents, not Java ones.

### Multi-Harness Output

Same detection, different output formats:

| Harness | Output File | Features |
|---------|------------|----------|
| Claude Code | `CLAUDE.md` | Full consilium + exec scope + profiles |
| Cursor | `.cursorrules` | Expert roles + file ownership |
| Windsurf | `.windsurfrules` | Stack context + code areas |

### Workflow Profiles

Built-in workflow templates (for Claude Code):

- **business-feature** — Research → Plan → Executing → Validation → Report
- **bug-hunting** — Reproduce → Diagnose → Fix → Validation → Report
- **research** — Consilium investigation, no code changes
- **refactoring** — Audit → Plan → Executing → Regression check
- **e2e-testing** — Prepare → Deploy → Run → Fix → Re-run → Report
- **e2e-authoring** — Research → Propose → Approve → Save scenarios

## License

This software is licensed under [CC BY-NC 4.0](LICENSE) — free for non-commercial use.
For commercial licensing, contact the author.
