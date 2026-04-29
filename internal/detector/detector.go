package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type Stack struct {
	Name     string // e.g. "spring-boot", "vue", "axum"
	Lang     string // e.g. "kotlin", "swift", "typescript"
	Category string // "backend", "frontend", "mobile", "shared", "desktop", "infra", "data"
	Path     string // relative path from project root
}

func Detect(root string) []Stack {
	var stacks []Stack

	detectors := []func(string) []Stack{
		detectKotlin,
		detectJava,
		detectSwift,
		detectJavaScript,
		detectPython,
		detectGo,
		detectRust,
		detectFlutter,
		detectRuby,
		detectPHP,
		detectDotNet,
		detectElixir,
		detectErlang,
		detectScala,
		detectClojure,
		detectHaskell,
		detectOCaml,
		detectCCpp,
		detectZig,
		detectNim,
		detectVlang,
		detectCrystal,
		detectGleam,
		detectJulia,
		detectR,
		detectLua,
		detectPerl,
		detectGroovy,
		detectWordPress,
		detectInfra,
	}

	for _, d := range detectors {
		stacks = append(stacks, d(root)...)
	}

	return stacks
}

// ===========================================================================
// Kotlin
// ===========================================================================

func detectKotlin(root string) []Stack {
	var stacks []Stack

	// Spring Boot (Kotlin)
	for _, dir := range allCandidateDirs(root) {
		gradle := filepath.Join(root, dir, "build.gradle.kts")
		if isSpringGradle(gradle) {
			stacks = append(stacks, Stack{
				Name: "spring-boot", Lang: "kotlin", Category: "backend", Path: relPath(dir),
			})
			break
		}
	}

	// Ktor backend
	for _, dir := range allCandidateDirs(root) {
		gradle := filepath.Join(root, dir, "build.gradle.kts")
		if fileContains(gradle, "io.ktor") {
			if !hasStackInDir(stacks, dir) {
				stacks = append(stacks, Stack{
					Name: "ktor", Lang: "kotlin", Category: "backend", Path: relPath(dir),
				})
				break
			}
		}
	}

	// Quarkus (Kotlin/Gradle)
	for _, dir := range allCandidateDirs(root) {
		gradle := filepath.Join(root, dir, "build.gradle.kts")
		if fileContains(gradle, "quarkus") && !hasStackInDir(stacks, dir) {
			stacks = append(stacks, Stack{
				Name: "quarkus", Lang: "kotlin", Category: "backend", Path: relPath(dir),
			})
			break
		}
	}

	// Micronaut (Kotlin/Gradle)
	for _, dir := range allCandidateDirs(root) {
		gradle := filepath.Join(root, dir, "build.gradle.kts")
		if fileContains(gradle, "micronaut") && !hasStackInDir(stacks, dir) {
			stacks = append(stacks, Stack{
				Name: "micronaut", Lang: "kotlin", Category: "backend", Path: relPath(dir),
			})
			break
		}
	}

	// Compose Multiplatform / KMP
	composeFound := false
	for _, dir := range subdirs(root) {
		if strings.HasPrefix(strings.ToLower(dir), "compose") && dirExists(filepath.Join(root, dir)) {
			if fileExists(filepath.Join(root, dir, "build.gradle.kts")) {
				stacks = append(stacks, Stack{
					Name: "compose-multiplatform", Lang: "kotlin", Category: "shared", Path: dir + "/",
				})
				composeFound = true
				break
			}
		}
	}
	if !composeFound && fileContains(filepath.Join(root, "build.gradle.kts"), "compose") {
		stacks = append(stacks, Stack{
			Name: "compose-multiplatform", Lang: "kotlin", Category: "shared", Path: ".",
		})
		composeFound = true
	}

	// Android-only (skip if Compose Multiplatform found)
	if dirExists(filepath.Join(root, "app")) &&
		fileExists(filepath.Join(root, "app", "build.gradle.kts")) &&
		!composeFound {
		stacks = append(stacks, Stack{
			Name: "android", Lang: "kotlin", Category: "mobile", Path: "app/",
		})
	}

	return stacks
}

// ===========================================================================
// Java (Maven pom.xml or Gradle .gradle)
// ===========================================================================

func detectJava(root string) []Stack {
	var stacks []Stack

	for _, dir := range allCandidateDirs(root) {
		pom := filepath.Join(root, dir, "pom.xml")
		gradle := filepath.Join(root, dir, "build.gradle")

		// Spring Boot (Java)
		if fileContains(pom, "spring-boot") || fileContains(pom, "springframework") ||
			isSpringGradle(gradle) {
			stacks = append(stacks, Stack{
				Name: "spring-boot-java", Lang: "java", Category: "backend", Path: relPath(dir),
			})
			return stacks
		}

		// Quarkus (Java/Maven)
		if fileContains(pom, "quarkus") || fileContains(gradle, "quarkus") {
			stacks = append(stacks, Stack{
				Name: "quarkus", Lang: "java", Category: "backend", Path: relPath(dir),
			})
			return stacks
		}

		// Micronaut (Java/Maven)
		if fileContains(pom, "micronaut") || fileContains(gradle, "micronaut") {
			stacks = append(stacks, Stack{
				Name: "micronaut", Lang: "java", Category: "backend", Path: relPath(dir),
			})
			return stacks
		}

		// Plain Maven/Gradle Java (no specific framework)
		if fileExists(pom) || fileExists(gradle) {
			// Only detect if it's clearly a Java project (has src/main/java or .java files referenced)
			if fileContains(pom, "java") || dirExists(filepath.Join(root, dir, "src", "main", "java")) {
				stacks = append(stacks, Stack{
					Name: "java", Lang: "java", Category: "backend", Path: relPath(dir),
				})
				return stacks
			}
		}
	}

	return stacks
}

// ===========================================================================
// Swift
// ===========================================================================

func detectSwift(root string) []Stack {
	var stacks []Stack

	// iOS native app (iosApp/ convention from KMP projects)
	if dirExists(filepath.Join(root, "iosApp")) {
		stacks = append(stacks, Stack{
			Name: "ios-native", Lang: "swift", Category: "mobile", Path: "iosApp/",
		})
	}

	// Xcode project without iosApp convention
	for _, dir := range subdirs(root) {
		if strings.HasSuffix(dir, ".xcodeproj") || strings.HasSuffix(dir, ".xcworkspace") {
			if !dirExists(filepath.Join(root, "iosApp")) {
				stacks = append(stacks, Stack{
					Name: "ios-native", Lang: "swift", Category: "mobile", Path: ".",
				})
			}
			break
		}
	}

	// Vapor (Swift server)
	if fileExists(filepath.Join(root, "Package.swift")) {
		if fileContains(filepath.Join(root, "Package.swift"), "vapor") {
			stacks = append(stacks, Stack{
				Name: "vapor", Lang: "swift", Category: "backend", Path: ".",
			})
		} else {
			stacks = append(stacks, Stack{
				Name: "swift-package", Lang: "swift", Category: "backend", Path: ".",
			})
		}
	}

	return stacks
}

// ===========================================================================
// JavaScript / TypeScript
// ===========================================================================

type jsFramework struct {
	name     string
	category string
	markers  []string
	exclude  []string
	files    []string
}

var frontendFrameworks = []jsFramework{
	{name: "nuxt", category: "frontend", markers: []string{"\"nuxt\""}},
	{name: "sveltekit", category: "frontend", markers: []string{"\"@sveltejs/kit\""}},
	{name: "nextjs", category: "frontend", markers: []string{"\"next\""}},
	{name: "remix", category: "frontend", markers: []string{"\"@remix-run/react\"", "\"@remix-run/node\""}},
	{name: "gatsby", category: "frontend", markers: []string{"\"gatsby\""}},
	{name: "astro", category: "frontend", markers: []string{"\"astro\""}},
	{name: "qwik", category: "frontend", markers: []string{"\"@builder.io/qwik\""}},
	{name: "svelte", category: "frontend", markers: []string{"\"svelte\""}, exclude: []string{"\"@sveltejs/kit\""}},
	{name: "vue", category: "frontend", markers: []string{"\"vue\""}, exclude: []string{"\"nuxt\""}},
	{name: "react", category: "frontend", markers: []string{"\"react\""}, exclude: []string{"\"next\"", "\"@remix-run/react\"", "\"react-native\"", "\"expo\"", "\"gatsby\""}},
	{name: "solid", category: "frontend", markers: []string{"\"solid-js\""}},
	{name: "ember", category: "frontend", markers: []string{"\"ember-cli\"", "\"ember-source\""}},
	{name: "eleventy", category: "frontend", markers: []string{"\"@11ty/eleventy\""}},
	{name: "angular", category: "frontend", files: []string{"angular.json"}},
}

var backendJSFrameworks = []jsFramework{
	{name: "strapi", category: "backend", markers: []string{"\"@strapi/strapi\""}},
	{name: "node", category: "backend", markers: []string{"\"express\"", "\"fastify\"", "\"@nestjs/core\"", "\"nestjs\"", "\"koa\"", "\"hono\"", "\"@hono/node-server\"", "\"@adonisjs/core\""}},
}

var mobileJSFrameworks = []jsFramework{
	{name: "expo", category: "mobile", markers: []string{"\"expo\""}},
	{name: "react-native", category: "mobile", markers: []string{"\"react-native\""}, exclude: []string{"\"expo\""}},
	{name: "ionic", category: "mobile", markers: []string{"\"@ionic/core\"", "\"@ionic/react\"", "\"@ionic/angular\"", "\"@ionic/vue\""}},
	{name: "capacitor", category: "mobile", markers: []string{"\"@capacitor/core\""}, exclude: []string{"\"@ionic/core\""}},
}

var desktopJSFrameworks = []jsFramework{
	{name: "electron", category: "desktop", markers: []string{"\"electron\""}},
}

func detectJavaScript(root string) []Stack {
	var stacks []Stack
	candidates := allCandidateDirs(root)

	// Deno detection (deno.json / deno.jsonc)
	for _, dir := range candidates {
		if fileExists(filepath.Join(root, dir, "deno.json")) || fileExists(filepath.Join(root, dir, "deno.jsonc")) {
			stacks = append(stacks, Stack{
				Name: "deno", Lang: "typescript", Category: "backend", Path: relPath(dir),
			})
			break
		}
	}

	// Bun detection (bunfig.toml or bun.lockb without other markers)
	for _, dir := range candidates {
		if fileExists(filepath.Join(root, dir, "bunfig.toml")) {
			stacks = append(stacks, Stack{
				Name: "bun", Lang: "typescript", Category: "backend", Path: relPath(dir),
			})
			break
		}
	}

	// Framework detection from package.json
	groups := [][]jsFramework{frontendFrameworks, backendJSFrameworks, mobileJSFrameworks, desktopJSFrameworks}

	for _, group := range groups {
		found := false
		for _, dir := range candidates {
			if found {
				break
			}
			pkgPath := filepath.Join(root, dir, "package.json")
			if !fileExists(pkgPath) {
				continue
			}
			for _, fw := range group {
				if matchFramework(root, dir, pkgPath, fw) {
					stacks = append(stacks, Stack{
						Name: fw.name, Lang: "typescript", Category: fw.category, Path: relPath(dir),
					})
					found = true
					break
				}
			}
		}
	}

	return stacks
}

func matchFramework(root, dir, pkgPath string, fw jsFramework) bool {
	if len(fw.files) > 0 {
		for _, f := range fw.files {
			if fileExists(filepath.Join(root, dir, f)) {
				return true
			}
		}
		return false
	}

	matched := false
	for _, marker := range fw.markers {
		if fileContains(pkgPath, marker) {
			matched = true
			break
		}
	}
	if !matched {
		return false
	}

	for _, ex := range fw.exclude {
		if fileContains(pkgPath, ex) {
			return false
		}
	}
	return true
}

// ===========================================================================
// Python
// ===========================================================================

func detectPython(root string) []Stack {
	var stacks []Stack

	type pyFramework struct {
		name     string
		category string
	}

	frameworks := []pyFramework{
		{"fastapi", "backend"},
		{"django", "backend"},
		{"flask", "backend"},
		{"starlette", "backend"},
		{"pyramid", "backend"},
		{"litestar", "backend"},
		{"streamlit", "data"},
		{"gradio", "data"},
	}

	for _, dir := range allCandidateDirs(root) {
		pyproject := filepath.Join(root, dir, "pyproject.toml")
		requirements := filepath.Join(root, dir, "requirements.txt")
		setupPy := filepath.Join(root, dir, "setup.py")

		for _, fw := range frameworks {
			if fileContains(pyproject, fw.name) || fileContains(requirements, fw.name) || fileContains(setupPy, fw.name) {
				stacks = append(stacks, Stack{
					Name: fw.name, Lang: "python", Category: fw.category, Path: relPath(dir),
				})
				return stacks
			}
		}
	}

	// Jupyter notebooks (*.ipynb in root)
	entries, err := os.ReadDir(root)
	if err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".ipynb") {
				stacks = append(stacks, Stack{
					Name: "jupyter", Lang: "python", Category: "data", Path: ".",
				})
				break
			}
		}
	}

	return stacks
}

// ===========================================================================
// Go (with framework sub-detection)
// ===========================================================================

func detectGo(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		goMod := filepath.Join(root, dir, "go.mod")
		if !fileExists(goMod) {
			continue
		}

		// Check for specific frameworks
		type goFw struct {
			name   string
			marker string
		}
		frameworks := []goFw{
			{"gin", "github.com/gin-gonic/gin"},
			{"fiber", "github.com/gofiber/fiber"},
			{"echo", "github.com/labstack/echo"},
			{"chi", "github.com/go-chi/chi"},
			{"buffalo", "github.com/gobuffalo/buffalo"},
		}

		for _, fw := range frameworks {
			if fileContains(goMod, fw.marker) {
				return []Stack{{
					Name: fw.name, Lang: "go", Category: "backend", Path: relPath(dir),
				}}
			}
		}

		// Plain Go
		return []Stack{{
			Name: "go", Lang: "go", Category: "backend", Path: relPath(dir),
		}}
	}
	return nil
}

// ===========================================================================
// Rust (with framework sub-detection + Tauri)
// ===========================================================================

func detectRust(root string) []Stack {
	var stacks []Stack

	for _, dir := range allCandidateDirs(root) {
		cargo := filepath.Join(root, dir, "Cargo.toml")
		if !fileExists(cargo) {
			continue
		}

		// Tauri desktop app
		if fileContains(cargo, "tauri") ||
			fileExists(filepath.Join(root, dir, "tauri.conf.json")) ||
			fileExists(filepath.Join(root, "src-tauri", "tauri.conf.json")) {
			stacks = append(stacks, Stack{
				Name: "tauri", Lang: "rust", Category: "desktop", Path: relPath(dir),
			})
			return stacks
		}

		// Web frameworks
		type rustFw struct {
			name   string
			marker string
		}
		frameworks := []rustFw{
			{"axum", "axum"},
			{"actix", "actix-web"},
			{"rocket", "rocket"},
			{"warp", "warp"},
		}

		for _, fw := range frameworks {
			if fileContains(cargo, fw.marker) {
				stacks = append(stacks, Stack{
					Name: fw.name, Lang: "rust", Category: "backend", Path: relPath(dir),
				})
				return stacks
			}
		}

		// Plain Rust
		stacks = append(stacks, Stack{
			Name: "rust", Lang: "rust", Category: "backend", Path: relPath(dir),
		})
		return stacks
	}

	return nil
}

// ===========================================================================
// Flutter / Dart
// ===========================================================================

func detectFlutter(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "pubspec.yaml")) {
			return []Stack{{
				Name: "flutter", Lang: "dart", Category: "mobile", Path: relPath(dir),
			}}
		}
	}
	return nil
}

// ===========================================================================
// Ruby (Rails, Sinatra)
// ===========================================================================

func detectRuby(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		gemfile := filepath.Join(root, dir, "Gemfile")
		if !fileExists(gemfile) {
			continue
		}
		if fileContains(gemfile, "rails") {
			return []Stack{{Name: "rails", Lang: "ruby", Category: "backend", Path: relPath(dir)}}
		}
		if fileContains(gemfile, "sinatra") {
			return []Stack{{Name: "sinatra", Lang: "ruby", Category: "backend", Path: relPath(dir)}}
		}
		// Jekyll (Ruby SSG)
		if fileContains(gemfile, "jekyll") {
			return []Stack{{Name: "jekyll", Lang: "ruby", Category: "frontend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// PHP (Laravel, Symfony)
// ===========================================================================

func detectPHP(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		composer := filepath.Join(root, dir, "composer.json")
		if !fileExists(composer) {
			continue
		}
		if fileContains(composer, "laravel/framework") {
			return []Stack{{Name: "laravel", Lang: "php", Category: "backend", Path: relPath(dir)}}
		}
		if fileContains(composer, "symfony/framework-bundle") || fileContains(composer, "symfony/symfony") {
			return []Stack{{Name: "symfony", Lang: "php", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// .NET / C#
// ===========================================================================

func detectDotNet(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".csproj") {
				csproj := filepath.Join(root, dir, e.Name())
				name := "dotnet"
				if fileContains(csproj, "Maui") {
					name = "maui"
					return []Stack{{Name: name, Lang: "csharp", Category: "mobile", Path: relPath(dir)}}
				}
				if fileContains(csproj, "Microsoft.AspNetCore") || fileContains(csproj, "Microsoft.NET.Sdk.Web") {
					return []Stack{{Name: name, Lang: "csharp", Category: "backend", Path: relPath(dir)}}
				}
				return []Stack{{Name: name, Lang: "csharp", Category: "backend", Path: relPath(dir)}}
			}
			if strings.HasSuffix(e.Name(), ".sln") {
				return []Stack{{Name: "dotnet", Lang: "csharp", Category: "backend", Path: relPath(dir)}}
			}
		}
	}
	return nil
}

// ===========================================================================
// Elixir (Phoenix)
// ===========================================================================

func detectElixir(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		mixExs := filepath.Join(root, dir, "mix.exs")
		if !fileExists(mixExs) {
			continue
		}
		name := "elixir"
		if fileContains(mixExs, "phoenix") {
			name = "phoenix"
		}
		return []Stack{{Name: name, Lang: "elixir", Category: "backend", Path: relPath(dir)}}
	}
	return nil
}

// ===========================================================================
// Erlang
// ===========================================================================

func detectErlang(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "rebar.config")) || fileExists(filepath.Join(root, dir, "rebar.lock")) {
			return []Stack{{Name: "erlang", Lang: "erlang", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Scala (sbt, Play, Akka)
// ===========================================================================

func detectScala(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		sbt := filepath.Join(root, dir, "build.sbt")
		if !fileExists(sbt) {
			continue
		}
		if fileContains(sbt, "play") {
			return []Stack{{Name: "play", Lang: "scala", Category: "backend", Path: relPath(dir)}}
		}
		if fileContains(sbt, "akka") {
			return []Stack{{Name: "akka", Lang: "scala", Category: "backend", Path: relPath(dir)}}
		}
		return []Stack{{Name: "scala", Lang: "scala", Category: "backend", Path: relPath(dir)}}
	}
	return nil
}

// ===========================================================================
// Clojure
// ===========================================================================

func detectClojure(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "deps.edn")) || fileExists(filepath.Join(root, dir, "project.clj")) {
			return []Stack{{Name: "clojure", Lang: "clojure", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Haskell
// ===========================================================================

func detectHaskell(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "stack.yaml")) {
			return []Stack{{Name: "haskell", Lang: "haskell", Category: "backend", Path: relPath(dir)}}
		}
		// Check for .cabal files
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".cabal") {
				return []Stack{{Name: "haskell", Lang: "haskell", Category: "backend", Path: relPath(dir)}}
			}
		}
	}
	return nil
}

// ===========================================================================
// OCaml
// ===========================================================================

func detectOCaml(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "dune-project")) || fileExists(filepath.Join(root, dir, "dune")) {
			return []Stack{{Name: "ocaml", Lang: "ocaml", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// C / C++
// ===========================================================================

func detectCCpp(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		cmake := filepath.Join(root, dir, "CMakeLists.txt")
		meson := filepath.Join(root, dir, "meson.build")

		if fileExists(cmake) {
			// Try to distinguish C vs C++
			lang := "cpp"
			name := "cpp"
			if fileContains(cmake, "project") && fileContains(cmake, " C)") && !fileContains(cmake, "CXX") {
				lang = "c"
				name = "c"
			}
			return []Stack{{Name: name, Lang: lang, Category: "backend", Path: relPath(dir)}}
		}

		if fileExists(meson) {
			lang := "cpp"
			name := "cpp"
			if fileContains(meson, "'c'") && !fileContains(meson, "'cpp'") {
				lang = "c"
				name = "c"
			}
			return []Stack{{Name: name, Lang: lang, Category: "backend", Path: relPath(dir)}}
		}

		// Plain Makefile with .c/.cpp sources
		makefile := filepath.Join(root, dir, "Makefile")
		if fileExists(makefile) {
			if fileContains(makefile, ".cpp") || fileContains(makefile, "g++") || fileContains(makefile, "CXX") {
				return []Stack{{Name: "cpp", Lang: "cpp", Category: "backend", Path: relPath(dir)}}
			}
			if fileContains(makefile, ".c") || fileContains(makefile, "gcc") {
				return []Stack{{Name: "c", Lang: "c", Category: "backend", Path: relPath(dir)}}
			}
		}
	}
	return nil
}

// ===========================================================================
// Zig
// ===========================================================================

func detectZig(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "build.zig")) {
			return []Stack{{Name: "zig", Lang: "zig", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Nim
// ===========================================================================

func detectNim(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".nimble") {
				return []Stack{{Name: "nim", Lang: "nim", Category: "backend", Path: relPath(dir)}}
			}
		}
	}
	return nil
}

// ===========================================================================
// V (vlang)
// ===========================================================================

func detectVlang(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "v.mod")) {
			return []Stack{{Name: "vlang", Lang: "vlang", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Crystal
// ===========================================================================

func detectCrystal(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "shard.yml")) {
			return []Stack{{Name: "crystal", Lang: "crystal", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Gleam
// ===========================================================================

func detectGleam(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "gleam.toml")) {
			return []Stack{{Name: "gleam", Lang: "gleam", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Julia
// ===========================================================================

func detectJulia(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "Project.toml")) {
			// Verify it's Julia (not Rust or something else)
			if fileContains(filepath.Join(root, dir, "Project.toml"), "uuid") {
				return []Stack{{Name: "julia", Lang: "julia", Category: "data", Path: relPath(dir)}}
			}
		}
	}
	return nil
}

// ===========================================================================
// R
// ===========================================================================

func detectR(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		desc := filepath.Join(root, dir, "DESCRIPTION")
		if fileExists(desc) && fileContains(desc, "Type:") {
			return []Stack{{Name: "r", Lang: "r", Category: "data", Path: relPath(dir)}}
		}
		// Also check for .Rproj files
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".Rproj") {
				return []Stack{{Name: "r", Lang: "r", Category: "data", Path: relPath(dir)}}
			}
		}
	}
	return nil
}

// ===========================================================================
// Lua
// ===========================================================================

func detectLua(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".rockspec") {
				return []Stack{{Name: "lua", Lang: "lua", Category: "backend", Path: relPath(dir)}}
			}
		}
		if fileExists(filepath.Join(root, dir, ".luarocks")) || fileExists(filepath.Join(root, dir, ".luacheckrc")) {
			return []Stack{{Name: "lua", Lang: "lua", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Perl
// ===========================================================================

func detectPerl(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "cpanfile")) ||
			fileExists(filepath.Join(root, dir, "Makefile.PL")) ||
			fileExists(filepath.Join(root, dir, "dist.ini")) {
			return []Stack{{Name: "perl", Lang: "perl", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Groovy (Grails, Gradle Groovy DSL scripts)
// ===========================================================================

func detectGroovy(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		gradle := filepath.Join(root, dir, "build.gradle")
		if !fileExists(gradle) {
			continue
		}
		if fileContains(gradle, "grails") {
			return []Stack{{Name: "grails", Lang: "groovy", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// WordPress
// ===========================================================================

func detectWordPress(root string) []Stack {
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "wp-config.php")) ||
			fileExists(filepath.Join(root, dir, "wp-config-sample.php")) ||
			dirExists(filepath.Join(root, dir, "wp-content")) {
			return []Stack{{Name: "wordpress", Lang: "php", Category: "backend", Path: relPath(dir)}}
		}
	}
	return nil
}

// ===========================================================================
// Infrastructure (Docker, Terraform, Helm, Pulumi, Ansible, GitHub Actions)
// ===========================================================================

func detectInfra(root string) []Stack {
	var stacks []Stack

	// Docker
	if fileExists(filepath.Join(root, "Dockerfile")) ||
		fileExists(filepath.Join(root, "docker-compose.yml")) ||
		fileExists(filepath.Join(root, "docker-compose.yaml")) ||
		fileExists(filepath.Join(root, "compose.yml")) ||
		fileExists(filepath.Join(root, "compose.yaml")) {
		stacks = append(stacks, Stack{
			Name: "docker", Lang: "dockerfile", Category: "infra", Path: ".",
		})
	}

	// Terraform
	for _, dir := range allCandidateDirs(root) {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".tf") {
				stacks = append(stacks, Stack{
					Name: "terraform", Lang: "hcl", Category: "infra", Path: relPath(dir),
				})
				goto doneTerraform
			}
		}
	}
doneTerraform:

	// Helm
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "Chart.yaml")) {
			stacks = append(stacks, Stack{
				Name: "helm", Lang: "yaml", Category: "infra", Path: relPath(dir),
			})
			break
		}
	}

	// Pulumi
	for _, dir := range allCandidateDirs(root) {
		if fileExists(filepath.Join(root, dir, "Pulumi.yaml")) {
			stacks = append(stacks, Stack{
				Name: "pulumi", Lang: "yaml", Category: "infra", Path: relPath(dir),
			})
			break
		}
	}

	// Ansible
	if fileExists(filepath.Join(root, "ansible.cfg")) ||
		fileExists(filepath.Join(root, "playbook.yml")) ||
		fileExists(filepath.Join(root, "site.yml")) ||
		dirExists(filepath.Join(root, "playbooks")) ||
		dirExists(filepath.Join(root, "roles")) {
		stacks = append(stacks, Stack{
			Name: "ansible", Lang: "yaml", Category: "infra", Path: ".",
		})
	}

	// GitHub Actions
	if dirExists(filepath.Join(root, ".github", "workflows")) {
		stacks = append(stacks, Stack{
			Name: "github-actions", Lang: "yaml", Category: "infra", Path: ".github/workflows/",
		})
	}

	// Hugo (static site generator)
	if fileExists(filepath.Join(root, "hugo.toml")) || fileExists(filepath.Join(root, "hugo.yaml")) ||
		(fileExists(filepath.Join(root, "config.toml")) && fileContains(filepath.Join(root, "config.toml"), "baseurl")) {
		stacks = append(stacks, Stack{
			Name: "hugo", Lang: "go", Category: "frontend", Path: ".",
		})
	}

	return stacks
}

// ===========================================================================
// Helpers
// ===========================================================================

func subdirs(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

func allCandidateDirs(root string) []string {
	dirs := subdirs(root)
	dirs = append(dirs, ".")
	return dirs
}

func isSpringGradle(path string) bool {
	return fileContains(path, "spring-boot") ||
		fileContains(path, "spring.boot") ||
		fileContains(path, "springframework.boot")
}

func relPath(dir string) string {
	if dir == "." {
		return "."
	}
	return dir + "/"
}

func hasStackInDir(stacks []Stack, dir string) bool {
	path := relPath(dir)
	for _, s := range stacks {
		if s.Path == path {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileContains(path string, substr string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), strings.ToLower(substr))
}

func findRelDir(root, name string) string {
	if dirExists(filepath.Join(root, name)) {
		return name + "/"
	}
	return "."
}
