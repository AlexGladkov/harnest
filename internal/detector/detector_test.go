package detector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubdirs_SkipsVendored(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"src", "node_modules/foo", "vendor/bar", "build/x", "target/y"} {
		if err := os.MkdirAll(filepath.Join(root, p), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	got := subdirs(root)
	for _, d := range got {
		for _, skip := range []string{"node_modules", "vendor", "build", "target"} {
			if strings.HasPrefix(d, skip) {
				t.Errorf("subdirs returned vendored dir %q — should be skipped", d)
			}
		}
	}

	var hasSrc bool
	for _, d := range got {
		if d == "src" {
			hasSrc = true
		}
	}
	if !hasSrc {
		t.Errorf("expected \"src\" in subdirs, got %v", got)
	}
}

func TestDetectDotNet_Sln(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "test.sln"), "")

	stacks := detectDotNet(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "dotnet" || stacks[0].Lang != "csharp" {
		t.Errorf("expected dotnet/csharp, got %s/%s", stacks[0].Name, stacks[0].Lang)
	}
}

func TestDetectDotNet_Slnx(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "test.slnx"), "")

	stacks := detectDotNet(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "dotnet" || stacks[0].Lang != "csharp" {
		t.Errorf("expected dotnet/csharp, got %s/%s", stacks[0].Name, stacks[0].Lang)
	}
}

func TestDetectDotNet_Csproj(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.csproj"), `<Project Sdk="Microsoft.NET.Sdk.Web"></Project>`)

	stacks := detectDotNet(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "dotnet" || stacks[0].Lang != "csharp" {
		t.Errorf("expected dotnet/csharp, got %s/%s", stacks[0].Name, stacks[0].Lang)
	}
	if stacks[0].Category != "backend" {
		t.Errorf("expected backend, got %s", stacks[0].Category)
	}
}

func TestDetectDotNet_Maui(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.csproj"), `<Project Sdk="Maui"></Project>`)

	stacks := detectDotNet(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "maui" || stacks[0].Category != "mobile" {
		t.Errorf("expected maui/mobile, got %s/%s", stacks[0].Name, stacks[0].Category)
	}
}

func TestDetectDotNet_None(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "readme.md"), "")

	stacks := detectDotNet(root)
	if len(stacks) != 0 {
		t.Errorf("expected 0 stacks, got %d", len(stacks))
	}
}

func TestDetectDotNet_Subdir(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	mustMkdir(t, srcDir)
	writeFile(t, filepath.Join(srcDir, "app.slnx"), "")

	stacks := detectDotNet(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack in subdir, got %d", len(stacks))
	}
	if stacks[0].Path != "src/" {
		t.Errorf("expected src/, got %s", stacks[0].Path)
	}
}

func TestDetectDotNet_DeepNested(t *testing.T) {
	// .slnx at depth 3 should still be detected
	root := t.TempDir()
	nested := filepath.Join(root, "src", "server", "api")
	mustMkdir(t, nested)
	writeFile(t, filepath.Join(nested, "project.slnx"), "")

	stacks := detectDotNet(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack at depth 3, got %d", len(stacks))
	}
	if stacks[0].Name != "dotnet" {
		t.Errorf("expected dotnet, got %s", stacks[0].Name)
	}
}

func TestDetectDotNet_MultiProject(t *testing.T) {
	// Multiple .slnx files in different directories → multiple stacks
	root := t.TempDir()
	proj1 := filepath.Join(root, "src", "Api")
	proj2 := filepath.Join(root, "src", "Worker")
	mustMkdir(t, proj1)
	mustMkdir(t, proj2)
	writeFile(t, filepath.Join(proj1, "Api.slnx"), "")
	writeFile(t, filepath.Join(proj2, "Worker.slnx"), "")

	stacks := detectDotNet(root)
	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}
	if stacks[0].Name != "dotnet" || stacks[1].Name != "dotnet" {
		t.Errorf("expected dotnet stacks, got %s, %s", stacks[0].Name, stacks[1].Name)
	}
	if stacks[0].Path == stacks[1].Path {
		t.Error("paths should differ")
	}
}

func TestDetectDotNet_CsprojPriority(t *testing.T) {
	// .csproj takes priority over .sln/.slnx
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.sln"), "")
	writeFile(t, filepath.Join(root, "app.csproj"), `<Project Sdk="Microsoft.NET.Sdk.Web"></Project>`)

	stacks := detectDotNet(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	// csproj is checked first, so it should be detected
	if stacks[0].Name != "dotnet" || stacks[0].Category != "backend" {
		t.Errorf("csproj should take priority, got %s/%s", stacks[0].Name, stacks[0].Category)
	}
}

// =============================================================================
// Kotlin
// =============================================================================

func TestDetectKotlin_SpringBoot(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "build.gradle.kts"), `
plugins { id("org.springframework.boot") }
dependencies { implementation("org.springframework.boot:spring-boot-starter-web") }
`)
	stacks := detectKotlin(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "spring-boot" || stacks[0].Lang != "kotlin" {
		t.Errorf("expected spring-boot/kotlin, got %s/%s", stacks[0].Name, stacks[0].Lang)
	}
}

func TestDetectKotlin_MultiModule(t *testing.T) {
	root := t.TempDir()
	m1 := filepath.Join(root, "service-a")
	m2 := filepath.Join(root, "service-b")
	mustMkdir(t, m1)
	mustMkdir(t, m2)
	writeFile(t, filepath.Join(m1, "build.gradle.kts"), `
plugins { id("org.springframework.boot") }
`)
	writeFile(t, filepath.Join(m2, "build.gradle.kts"), `
plugins { id("org.springframework.boot") }
`)
	stacks := detectKotlin(root)
	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}
	for _, s := range stacks {
		if s.Name != "spring-boot" {
			t.Errorf("expected spring-boot, got %s", s.Name)
		}
	}
}

func TestDetectKotlin_Ktor(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "build.gradle.kts"), `
dependencies { implementation("io.ktor:ktor-server-core") }
`)
	stacks := detectKotlin(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "ktor" {
		t.Errorf("expected ktor, got %s", stacks[0].Name)
	}
}

func TestDetectKotlin_None(t *testing.T) {
	root := t.TempDir()
	stacks := detectKotlin(root)
	if len(stacks) != 0 {
		t.Errorf("expected 0 stacks, got %d", len(stacks))
	}
}

func TestDetectKotlin_ComposeMultiplatform(t *testing.T) {
	root := t.TempDir()
	composeDir := filepath.Join(root, "composeApp")
	mustMkdir(t, composeDir)
	writeFile(t, filepath.Join(composeDir, "build.gradle.kts"), `
plugins { kotlin("multiplatform") }
`)
	stacks := detectKotlin(root)
	found := false
	for _, s := range stacks {
		if s.Name == "compose-multiplatform" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected compose-multiplatform in %v", stacks)
	}
}

// =============================================================================
// Java
// =============================================================================

func TestDetectJava_SpringBoot(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pom.xml"), `
<project>
  <parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
  </parent>
</project>
`)
	stacks := detectJava(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "spring-boot-java" || stacks[0].Lang != "java" {
		t.Errorf("expected spring-boot-java/java, got %s/%s", stacks[0].Name, stacks[0].Lang)
	}
}

func TestDetectJava_MultiModule(t *testing.T) {
	root := t.TempDir()
	m1 := filepath.Join(root, "mod-a", "src", "main", "java")
	m2 := filepath.Join(root, "mod-b", "src", "main", "java")
	mustMkdir(t, m1)
	mustMkdir(t, m2)
	writeFile(t, filepath.Join(root, "mod-a", "pom.xml"), "<project></project>")
	writeFile(t, filepath.Join(root, "mod-b", "pom.xml"), "<project></project>")

	stacks := detectJava(root)
	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d: %v", len(stacks), stacks)
	}
	for _, s := range stacks {
		if s.Name != "java" {
			t.Errorf("expected java, got %s", s.Name)
		}
	}
}

func TestDetectJava_PlainGradle(t *testing.T) {
	root := t.TempDir()
	javaSrc := filepath.Join(root, "src", "main", "java")
	mustMkdir(t, javaSrc)
	writeFile(t, filepath.Join(root, "build.gradle"), "")

	stacks := detectJava(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "java" {
		t.Errorf("expected java, got %s", stacks[0].Name)
	}
}

func TestDetectJava_None(t *testing.T) {
	root := t.TempDir()
	stacks := detectJava(root)
	if len(stacks) != 0 {
		t.Errorf("expected 0 stacks, got %d", len(stacks))
	}
}

func TestDetectJava_Quarkus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "pom.xml"), `
<project>
  <groupId>io.quarkus</groupId>
</project>
`)
	stacks := detectJava(root)
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "quarkus" {
		t.Errorf("expected quarkus, got %s", stacks[0].Name)
	}
}

// =============================================================================
// DotNet additional corner cases
// =============================================================================

func TestDetectDotNet_MixedSlnAndCsprojDirs(t *testing.T) {
	// One dir has .slnx, another has .csproj — both should be detected
	root := t.TempDir()
	dir1 := filepath.Join(root, "backend", "Api")
	dir2 := filepath.Join(root, "backend", "Worker")
	mustMkdir(t, dir1)
	mustMkdir(t, dir2)
	writeFile(t, filepath.Join(dir1, "Api.slnx"), "")
	writeFile(t, filepath.Join(dir2, "Worker.csproj"), `<Project Sdk="Microsoft.NET.Sdk"></Project>`)

	stacks := detectDotNet(root)
	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}
}

func TestDetectDotNet_ManyMicroservices(t *testing.T) {
	// 5 services — all should be found
	root := t.TempDir()
	for _, svc := range []string{"auth", "api", "worker", "notifier", "gateway"} {
		dir := filepath.Join(root, "services", svc)
		mustMkdir(t, dir)
		writeFile(t, filepath.Join(dir, svc+".slnx"), "")
	}
	stacks := detectDotNet(root)
	if len(stacks) != 5 {
		t.Fatalf("expected 5 stacks, got %d", len(stacks))
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
}
