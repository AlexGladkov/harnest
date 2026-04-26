package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type Stack struct {
	Name     string // e.g. "spring-boot", "vue", "swift", "kmp"
	Lang     string // e.g. "kotlin", "swift", "typescript", "python"
	Category string // "backend", "frontend", "mobile", "shared"
	Path     string // relative path from project root
}

func Detect(root string) []Stack {
	var stacks []Stack

	detectors := []func(string) []Stack{
		detectKotlin,
		detectSwift,
		detectJavaScript,
		detectPython,
		detectGo,
		detectRust,
		detectFlutter,
	}

	for _, d := range detectors {
		stacks = append(stacks, d(root)...)
	}

	return stacks
}

func detectKotlin(root string) []Stack {
	var stacks []Stack

	// Spring Boot backend
	if fileContains(filepath.Join(root, "backend", "build.gradle.kts"), "spring-boot") ||
		fileContains(filepath.Join(root, "backend", "build.gradle.kts"), "spring.boot") ||
		fileContains(filepath.Join(root, "build.gradle.kts"), "spring-boot") ||
		fileContains(filepath.Join(root, "build.gradle.kts"), "spring.boot") {
		stacks = append(stacks, Stack{
			Name:     "spring-boot",
			Lang:     "kotlin",
			Category: "backend",
			Path:     findRelDir(root, "backend"),
		})
	}

	// Compose Multiplatform / KMP
	if dirExists(filepath.Join(root, "composeApp")) ||
		fileContains(filepath.Join(root, "build.gradle.kts"), "compose") {
		stacks = append(stacks, Stack{
			Name:     "compose-multiplatform",
			Lang:     "kotlin",
			Category: "shared",
			Path:     findRelDir(root, "composeApp"),
		})
	}

	// Android-only
	if dirExists(filepath.Join(root, "app")) &&
		fileExists(filepath.Join(root, "app", "build.gradle.kts")) &&
		!dirExists(filepath.Join(root, "composeApp")) {
		stacks = append(stacks, Stack{
			Name:     "android",
			Lang:     "kotlin",
			Category: "mobile",
			Path:     "app/",
		})
	}

	return stacks
}

func detectSwift(root string) []Stack {
	var stacks []Stack

	if dirExists(filepath.Join(root, "iosApp")) {
		stacks = append(stacks, Stack{
			Name:     "ios-native",
			Lang:     "swift",
			Category: "mobile",
			Path:     "iosApp/",
		})
	}

	if fileExists(filepath.Join(root, "Package.swift")) {
		stacks = append(stacks, Stack{
			Name:     "swift-package",
			Lang:     "swift",
			Category: "backend",
			Path:     ".",
		})
	}

	return stacks
}

func detectJavaScript(root string) []Stack {
	var stacks []Stack

	// Vue
	for _, dir := range []string{"vue-frontend", "frontend", "web", "."} {
		pkgPath := filepath.Join(root, dir, "package.json")
		if fileContains(pkgPath, "\"vue\"") {
			stacks = append(stacks, Stack{
				Name:     "vue",
				Lang:     "typescript",
				Category: "frontend",
				Path:     dir + "/",
			})
			break
		}
	}

	// React / Next.js
	for _, dir := range []string{"frontend", "web", "client", "."} {
		pkgPath := filepath.Join(root, dir, "package.json")
		if fileContains(pkgPath, "\"next\"") {
			stacks = append(stacks, Stack{
				Name:     "nextjs",
				Lang:     "typescript",
				Category: "frontend",
				Path:     dir + "/",
			})
			break
		} else if fileContains(pkgPath, "\"react\"") && !fileContains(pkgPath, "\"next\"") {
			stacks = append(stacks, Stack{
				Name:     "react",
				Lang:     "typescript",
				Category: "frontend",
				Path:     dir + "/",
			})
			break
		}
	}

	// Angular
	if fileExists(filepath.Join(root, "angular.json")) {
		stacks = append(stacks, Stack{
			Name:     "angular",
			Lang:     "typescript",
			Category: "frontend",
			Path:     ".",
		})
	}

	// Node.js backend
	for _, dir := range []string{"server", "api", "backend"} {
		pkgPath := filepath.Join(root, dir, "package.json")
		if fileContains(pkgPath, "\"express\"") || fileContains(pkgPath, "\"fastify\"") ||
			fileContains(pkgPath, "\"nestjs\"") || fileContains(pkgPath, "\"@nestjs/core\"") {
			stacks = append(stacks, Stack{
				Name:     "node",
				Lang:     "typescript",
				Category: "backend",
				Path:     dir + "/",
			})
			break
		}
	}

	return stacks
}

func detectPython(root string) []Stack {
	var stacks []Stack

	pyproject := filepath.Join(root, "pyproject.toml")
	requirements := filepath.Join(root, "requirements.txt")

	if fileContains(pyproject, "fastapi") || fileContains(requirements, "fastapi") {
		stacks = append(stacks, Stack{
			Name:     "fastapi",
			Lang:     "python",
			Category: "backend",
			Path:     ".",
		})
	} else if fileContains(pyproject, "django") || fileContains(requirements, "django") {
		stacks = append(stacks, Stack{
			Name:     "django",
			Lang:     "python",
			Category: "backend",
			Path:     ".",
		})
	} else if fileContains(pyproject, "flask") || fileContains(requirements, "flask") {
		stacks = append(stacks, Stack{
			Name:     "flask",
			Lang:     "python",
			Category: "backend",
			Path:     ".",
		})
	}

	return stacks
}

func detectGo(root string) []Stack {
	if fileExists(filepath.Join(root, "go.mod")) {
		return []Stack{{
			Name:     "go",
			Lang:     "go",
			Category: "backend",
			Path:     ".",
		}}
	}
	return nil
}

func detectRust(root string) []Stack {
	if fileExists(filepath.Join(root, "Cargo.toml")) {
		return []Stack{{
			Name:     "rust",
			Lang:     "rust",
			Category: "backend",
			Path:     ".",
		}}
	}
	return nil
}

func detectFlutter(root string) []Stack {
	if fileExists(filepath.Join(root, "pubspec.yaml")) {
		return []Stack{{
			Name:     "flutter",
			Lang:     "dart",
			Category: "mobile",
			Path:     ".",
		}}
	}
	return nil
}

// helpers

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
