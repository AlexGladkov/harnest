# Harnest

AI coding assistant configurator. Detects your project stack, runs an interactive agent wizard, and generates configs for Claude Code, Cursor, and Windsurf.

## About

Every AI coding tool needs project context. Manually maintaining `CLAUDE.md`, `.cursorrules`, `.windsurfrules` is tedious — especially for multi-stack projects where wrong agents get assigned.

Harnest solves this with a three-layer system:

<table>
<tr>
<td><img src="docs/layers.jpg" alt="Three layers of the system" /></td>
<td><img src="docs/flow.jpg" alt="Request processing flow" /></td>
</tr>
</table>

## Install

**1. Get the binary**

```bash
brew tap AlexGladkov/tap
brew install harnest
```

Or with Go:

```bash
go install github.com/AlexGladkov/harnest/cmd/harnest@latest
```

Or download from [Releases](https://github.com/AlexGladkov/harnest/releases).

**2. Install the framework**

```bash
harnest install
```

Installs 6 workflow profiles and global CLAUDE.md framework to `~/.claude/`. Uses `<!-- harnest-managed -->` markers — your custom content is preserved on updates.

## Scenarios

### First-time setup

```bash
# 1. Install profiles + global CLAUDE.md
harnest install

# 2. Go to your project and generate config
cd my-project
harnest init
```

The wizard detects your stack, then asks for each role and exec scope:

```
Detected stack:
  - spring-boot (kotlin) [backend/]
  - vue (typescript) [vue-frontend/]

── Agent Wizard ──
Enter = accept suggestion, s = skip, or type agent name

[Consilium: architect]
  Suggestion: voltagent-lang:java-architect
  Enter agent (Enter=suggestion, s=skip): _
```

- **Enter** — accept suggestion
- **s** — skip role (won't appear in config)
- **type name** — use your own agent

### Change agents after init

Generated config wrong agent? Override it:

```bash
harnest agents set architect my-custom-architect
```

Or edit the generated file directly (`CLAUDE.md`, `.cursorrules`, `.windsurfrules`) — it's plain markdown.

### View current agent mappings

```bash
harnest agents list
```

Shows consilium roles and exec scopes from your project config.

### Switch to a different harness

Already have `CLAUDE.md` but need `.cursorrules` too?

```bash
harnest convert --from claude-code --to cursor
```

Or re-run init for a specific harness:

```bash
harnest init --harness windsurf
```

### CI / scripts (no wizard)

```bash
harnest init --non-interactive
```

Uses suggested agents automatically. Defaults to Claude Code.

### Update profiles after Harnest upgrade

```bash
harnest install
```

Re-running `install` updates profiles and the managed block in global CLAUDE.md. Your custom content outside `<!-- harnest-managed -->` stays intact.

### Add or remove individual profiles

```bash
harnest profiles list
harnest profiles add e2e-testing
harnest profiles remove research
```

## Package

### Workflow profiles

Installed to `~/.claude/profiles/`. Each profile defines stages and roles — no hardcoded agents.

| Profile | Stages |
|---------|--------|
| business-feature | Research → Plan → Executing → Validation → Report |
| bug-hunting | Reproduce → Diagnose → Fix → Validation → Report |
| research | Consilium investigation, no code changes |
| refactoring | Audit → Plan → Executing → Regression check |
| e2e-testing | Prepare → Deploy → Run → Fix → Re-run → Report |
| e2e-authoring | Research → Propose → Approve → Save scenarios |

### Consilium roles

8 roles available for agent assignment during `harnest init`:

| Role | Purpose |
|------|---------|
| architect | Architecture, modules, dependencies, SOLID |
| frontend | UI/UX review, frontend patterns |
| ui | Visual design, UX, components |
| security | OWASP, vulnerabilities, auth |
| devops | Infrastructure, CI/CD, deployment |
| api | API contracts, REST/GraphQL |
| diagnostics | Logs, stacktraces, debugging |
| test | Test coverage, quality |

### Harness output formats

| Harness | Output File | Features |
|---------|------------|----------|
| Claude Code | `CLAUDE.md` | Full consilium + exec scope + profiles |
| Cursor | `.cursorrules` | Expert roles + file ownership |
| Windsurf | `.windsurfrules` | Stack context + code areas |

## License

This software is licensed under [CC BY-NC 4.0](LICENSE) — free for non-commercial use.
For commercial licensing, contact the author.
