# Harnest

AI coding assistant configurator. Detects your project stack, runs an interactive agent wizard, and generates configs for Claude Code, Cursor, Windsurf, Codex, OpenCode, and Qwen Code.

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

Or with npm:

```bash
npm install -g harnest
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
  - spring-boot (kotlin) [springApp/]
  - vue (typescript) [webApp/]
  - docker (dockerfile) [.]

── Agent Wizard ──
Found 42 agents on this machine
Enter = accept suggestion, s = skip, ? = search

[Consilium: architect]
  Suggestion: voltagent-lang:java-architect
  Enter=accept, s=skip, ?=search: _
```

- **Enter** — accept suggestion
- **s** — skip role (won't appear in config)
- **?** — search installed agents by name
- **type name** — use your own agent (confirms if not found locally)

### Detect stack without generating

```bash
harnest detect
```

Shows detected technologies without creating any config files.

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

Shows consilium roles and exec scopes from your project config. If no config exists, shows suggestions based on detected stack.

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

### Manage profiles

```bash
harnest profiles list              # show installed profiles
harnest profiles add my-workflow    # interactive wizard to create custom profile
harnest profiles edit my-workflow   # open in $EDITOR
harnest profiles remove research   # delete profile
```

The `add` wizard lets you build a custom profile step by step: pick stages, assign agent types (single / consilium / bash / none), select roles, and auto-generates stage transitions.

## CLI Reference

```
harnest install                                          Install framework
harnest init [dir] [--harness <name>] [--non-interactive]  Generate project config
harnest detect [dir]                                     Show detected stack
harnest profiles list|add|edit|remove [name]             Manage profiles
harnest agents list [dir]                                View agent mappings
harnest agents set <role> <agent> [--dir <path>]         Override agent for role
harnest convert --from <harness> --to <harness> [dir]    Convert between formats
harnest update                                           Check for mapping updates
harnest version                                          Show version
```

## Stack Detection

Harnest auto-detects **92 stacks** across **30+ languages** by scanning build files and dependency manifests. All detectors scan every first-level subdirectory — non-standard folder names like `springApp/` or `webApp/` work out of the box.

### Backend

| Stack | Language | Detection signal |
|-------|----------|-----------------|
| spring-boot | Kotlin | `build.gradle.kts` with spring-boot/springframework |
| ktor | Kotlin | `build.gradle.kts` with io.ktor |
| quarkus | Kotlin/Java | `build.gradle.kts` / `pom.xml` with quarkus |
| micronaut | Kotlin/Java | `build.gradle.kts` / `pom.xml` with micronaut |
| spring-boot-java | Java | `pom.xml` / `build.gradle` with spring-boot |
| vapor | Swift | `Package.swift` with vapor |
| swift-package | Swift | `Package.swift` |
| node | TypeScript | `package.json` with express/fastify/nestjs/koa/hono |
| deno | TypeScript | `deno.json` / `deno.jsonc` |
| bun | TypeScript | `bunfig.toml` |
| strapi | TypeScript | `package.json` with @strapi/strapi |
| fastapi | Python | `pyproject.toml` / `requirements.txt` with fastapi |
| django | Python | `pyproject.toml` / `requirements.txt` with django |
| flask | Python | `pyproject.toml` / `requirements.txt` with flask |
| starlette | Python | `pyproject.toml` / `requirements.txt` with starlette |
| pyramid | Python | `pyproject.toml` / `requirements.txt` with pyramid |
| litestar | Python | `pyproject.toml` / `requirements.txt` with litestar |
| go | Go | `go.mod` |
| gin | Go | `go.mod` with gin-gonic/gin |
| fiber | Go | `go.mod` with gofiber/fiber |
| echo | Go | `go.mod` with labstack/echo |
| chi | Go | `go.mod` with go-chi/chi |
| rust | Rust | `Cargo.toml` |
| axum | Rust | `Cargo.toml` with axum |
| actix | Rust | `Cargo.toml` with actix-web |
| rocket | Rust | `Cargo.toml` with rocket |
| rails | Ruby | `Gemfile` with rails |
| sinatra | Ruby | `Gemfile` with sinatra |
| laravel | PHP | `composer.json` with laravel/framework |
| symfony | PHP | `composer.json` with symfony/framework-bundle |
| wordpress | PHP | `wp-config.php` / `wp-content/` |
| dotnet | C# | `*.csproj` / `*.sln` |
| phoenix | Elixir | `mix.exs` with phoenix |
| elixir | Elixir | `mix.exs` |
| erlang | Erlang | `rebar.config` |
| scala / play / akka | Scala | `build.sbt` |
| clojure | Clojure | `deps.edn` / `project.clj` |
| grails | Groovy | `build.gradle` with grails |
| haskell | Haskell | `stack.yaml` / `*.cabal` |
| ocaml | OCaml | `dune-project` |
| c | C | `CMakeLists.txt` (C project) / `Makefile` with gcc |
| cpp | C++ | `CMakeLists.txt` (CXX) / `meson.build` / `Makefile` with g++ |
| zig | Zig | `build.zig` |
| nim | Nim | `*.nimble` |
| vlang | V | `v.mod` |
| crystal | Crystal | `shard.yml` |
| gleam | Gleam | `gleam.toml` |
| lua | Lua | `*.rockspec` / `.luacheckrc` |
| perl | Perl | `cpanfile` / `Makefile.PL` |

### Frontend

| Stack | Language | Detection signal |
|-------|----------|-----------------|
| vue | TypeScript | `package.json` with `"vue"` (excl. nuxt) |
| nuxt | TypeScript | `package.json` with `"nuxt"` |
| react | TypeScript | `package.json` with `"react"` (excl. next/remix/expo/RN) |
| nextjs | TypeScript | `package.json` with `"next"` |
| gatsby | TypeScript | `package.json` with `"gatsby"` |
| remix | TypeScript | `package.json` with `"@remix-run/react"` |
| angular | TypeScript | `angular.json` |
| svelte | TypeScript | `package.json` with `"svelte"` (excl. sveltekit) |
| sveltekit | TypeScript | `package.json` with `"@sveltejs/kit"` |
| solid | TypeScript | `package.json` with `"solid-js"` |
| qwik | TypeScript | `package.json` with `"@builder.io/qwik"` |
| astro | TypeScript | `package.json` with `"astro"` |
| ember | TypeScript | `package.json` with `"ember-cli"` |
| eleventy | TypeScript | `package.json` with `"@11ty/eleventy"` |
| hugo | Go | `hugo.toml` / `config.toml` with baseURL |
| jekyll | Ruby | `Gemfile` with jekyll |

### Mobile

| Stack | Language | Detection signal |
|-------|----------|-----------------|
| compose-multiplatform | Kotlin | `composeApp/` directory |
| android | Kotlin | `app/build.gradle.kts` (no composeApp) |
| ios-native | Swift | `iosApp/` or `*.xcodeproj` |
| flutter | Dart | `pubspec.yaml` |
| expo | TypeScript | `package.json` with `"expo"` |
| react-native | TypeScript | `package.json` with `"react-native"` (excl. expo) |
| ionic | TypeScript | `package.json` with `"@ionic/*"` |
| capacitor | TypeScript | `package.json` with `"@capacitor/core"` |
| maui | C# | `*.csproj` with Maui |

### Desktop

| Stack | Language | Detection signal |
|-------|----------|-----------------|
| electron | TypeScript | `package.json` with `"electron"` |
| tauri | Rust | `Cargo.toml` with tauri / `tauri.conf.json` |

### Data / Scientific

| Stack | Language | Detection signal |
|-------|----------|-----------------|
| streamlit | Python | `requirements.txt` with streamlit |
| gradio | Python | `requirements.txt` with gradio |
| jupyter | Python | `*.ipynb` files in root |
| julia | Julia | `Project.toml` with uuid |
| r | R | `DESCRIPTION` with Type / `*.Rproj` |

### Infrastructure

| Stack | Language | Detection signal |
|-------|----------|-----------------|
| docker | Dockerfile | `Dockerfile` / `docker-compose.yml` / `compose.yaml` |
| terraform | HCL | `*.tf` files |
| helm | YAML | `Chart.yaml` |
| pulumi | YAML | `Pulumi.yaml` |
| ansible | YAML | `ansible.cfg` / `playbook.yml` / `roles/` |
| github-actions | YAML | `.github/workflows/` |

Multi-stack projects are fully supported — each detected stack gets its own exec agent scope with the correct directory path.

## Agent Discovery

The wizard scans installed agents from multiple sources:

- **Harness agent dirs**: `~/.claude/agents/`, `~/.cursor/agents/`, `~/.windsurf/agents/`, `~/.codex/agents/`, `~/.config/opencode/agents/`, `~/.qwen/agents/`
- **Plugins**: `~/.claude/plugins/cache/*/plugin.json` — reads `agents` field, prefixes with plugin name

Search with `?` in the wizard to filter by substring.

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

9 roles available for agent assignment during `harnest init`:

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
| mobile | Mobile platforms, cross-platform |

### Harness output formats

| Harness | Output File | Features |
|---------|------------|----------|
| Claude Code | `CLAUDE.md` | Full consilium + exec scope + profiles |
| Cursor | `.cursorrules` | Expert roles + file ownership |
| Windsurf | `.windsurfrules` | Stack context + code areas |
| Codex | `AGENTS.md` | Expert perspectives + file ownership |
| OpenCode | `opencode.json` | Subagent declarations + agent files |
| Qwen Code | `QWEN.md` | Expert perspectives + file ownership |

## License

This software is licensed under [CC BY-NC 4.0](LICENSE) — free for non-commercial use.
For commercial licensing, contact the author.
