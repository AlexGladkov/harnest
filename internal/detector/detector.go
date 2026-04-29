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

	// Spring Boot backend — scan root + all subdirs for build.gradle.kts with spring markers
	for _, dir := range allCandidateDirs(root) {
		gradle := filepath.Join(root, dir, "build.gradle.kts")
		if isSpringGradle(gradle) {
			path := dir + "/"
			if dir == "." {
				path = "."
			}
			stacks = append(stacks, Stack{
				Name:     "spring-boot",
				Lang:     "kotlin",
				Category: "backend",
				Path:     path,
			})
			break
		}
	}


	// Compose Multiplatform / KMP — look for composeApp dir anywhere at top level
	composeFound := false
	for _, dir := range subdirs(root) {
		if strings.HasPrefix(strings.ToLower(dir), "compose") && dirExists(filepath.Join(root, dir)) {
			if fileExists(filepath.Join(root, dir, "build.gradle.kts")) {
				stacks = append(stacks, Stack{
					Name:     "compose-multiplatform",
					Lang:     "kotlin",
					Category: "shared",
					Path:     dir + "/",
				})
				composeFound = true
				break
			}
		}
	}
	if !composeFound && fileContains(filepath.Join(root, "build.gradle.kts"), "compose") {
		stacks = append(stacks, Stack{
			Name:     "compose-multiplatform",
			Lang:     "kotlin",
			Category: "shared",
			Path:     ".",
		})
		composeFound = true
	}

	// Android-only
	if dirExists(filepath.Join(root, "app")) &&
		fileExists(filepath.Join(root, "app", "build.gradle.kts")) &&
		!composeFound {
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

	candidates := allCandidateDirs(root)

	// Vue — scan all subdirs
	for _, dir := range candidates {
		pkgPath := filepath.Join(root, dir, "package.json")
		if fileContains(pkgPath, "\"vue\"") {
			path := dir + "/"
			if dir == "." {
				path = "."
			}
			stacks = append(stacks, Stack{
				Name:     "vue",
				Lang:     "typescript",
				Category: "frontend",
				Path:     path,
			})
			break
		}
	}

	// React / Next.js — scan all subdirs (skip if Vue already found in same dir)
	if len(stacks) == 0 {
		for _, dir := range candidates {
			pkgPath := filepath.Join(root, dir, "package.json")
			if fileContains(pkgPath, "\"next\"") {
				path := dir + "/"
				if dir == "." {
					path = "."
				}
				stacks = append(stacks, Stack{
					Name:     "nextjs",
					Lang:     "typescript",
					Category: "frontend",
					Path:     path,
				})
				break
			} else if fileContains(pkgPath, "\"react\"") && !fileContains(pkgPath, "\"next\"") {
				path := dir + "/"
				if dir == "." {
					path = "."
				}
				stacks = append(stacks, Stack{
					Name:     "react",
					Lang:     "typescript",
					Category: "frontend",
					Path:     path,
				})
				break
			}
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

	// Node.js backend — scan all subdirs
	for _, dir := range candidates {
		pkgPath := filepath.Join(root, dir, "package.json")
		if fileContains(pkgPath, "\"express\"") || fileContains(pkgPath, "\"fastify\"") ||
			fileContains(pkgPath, "\"nestjs\"") || fileContains(pkgPath, "\"@nestjs/core\"") {
			path := dir + "/"
			if dir == "." {
				path = "."
			}
			stacks = append(stacks, Stack{
				Name:     "node",
				Lang:     "typescript",
				Category: "backend",
				Path:     path,
			})
			break
		}
	}

	return stacks
}

func detectPython(root string) []Stack {
	var stacks []Stack

	for _, dir := range allCandidateDirs(root) {
		pyproject := filepath.Join(root, dir, "pyproject.toml")
		requirements := filepath.Join(root, dir, "requirements.txt")

		path := dir + "/"
		if dir == "." {
			path = "."
		}

		if fileContains(pyproject, "fastapi") || fileContains(requirements, "fastapi") {
			stacks = append(stacks, Stack{Name: "fastapi", Lang: "python", Category: "backend", Path: path})
			return stacks
		} else if fileContains(pyproject, "django") || fileContains(requirements, "django") {
			stacks = append(stacks, Stack{Name: "django", Lang: "python", Category: "backend", Path: path})
			return stacks
		} else if fileContains(pyproject, "flask") || fileContains(requirements, "flask") {
			stacks = append(stacks, Stack{Name: "flask", Lang: "python", Category: "backend", Path: path})
			return stacks
		}
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

// subdirs returns names of all first-level directories in root (excluding hidden dirs).
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

// allCandidateDirs returns all first-level subdirs plus "." (root) at the end.
// This allows detectors to scan every possible location.
func allCandidateDirs(root string) []string {
	dirs := subdirs(root)
	dirs = append(dirs, ".")
	return dirs
}

// isSpringGradle checks if a build.gradle.kts contains Spring Boot markers.
func isSpringGradle(path string) bool {
	return fileContains(path, "spring-boot") ||
		fileContains(path, "spring.boot") ||
		fileContains(path, "springframework.boot")
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
