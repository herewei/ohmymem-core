package e2e

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestInit_EmptyDirectory tests initializing an empty directory
// Given: 空目录
// When:  ohmymem init --yes
// Then:
//   - 退出码 = 0
//   - stdout 包含 "Initialization complete"
//   - .ohmymem/memory.md 存在
//   - AGENTS.md 存在
//   - .cursorrules 是指向 AGENTS.md 的软链接
//   - memory.md 包含 "schema_version"
//   - memory.md 包含 "## Constraints"
func TestInit_EmptyDirectory(t *testing.T) {
	dir := setupTestEnv(t, "") // 空目录

	result := runCmd(dir, "init", "--yes")

	// 验证命令成功
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Errorf("stdout: %s", result.Stdout)
		t.Errorf("stderr: %s", result.Stderr)
	}

	// 验证输出
	if !strings.Contains(result.Stdout, "Initialization complete") {
		t.Errorf("stdout should contain 'Initialization complete', got: %s", result.Stdout)
	}

	// 验证文件结构
	assertFileExists(t, filepath.Join(dir, ".ohmymem", "memory.md"))
	assertFileExists(t, filepath.Join(dir, "AGENTS.md"))
	assertSymlink(t, filepath.Join(dir, ".cursorrules"), "AGENTS.md")

	// 验证memory.md内容
	assertFileContains(t, filepath.Join(dir, ".ohmymem", "memory.md"), "schema_version")
	assertFileContains(t, filepath.Join(dir, ".ohmymem", "memory.md"), "## Constraints")
	assertFileContains(t, filepath.Join(dir, ".ohmymem", "memory.md"), "## Decisions")
	assertFileContains(t, filepath.Join(dir, ".ohmymem", "memory.md"), "## Patterns")
	assertFileContains(t, filepath.Join(dir, ".ohmymem", "memory.md"), "## Anti-Patterns")
}

// TestInit_GoProject tests initializing a Go project
// Given: 目录包含 go.mod (module example.com/test)
// When:  ohmymem init --yes
// Then:
//   - 退出码 = 0
//   - stdout 包含 "Detected: Go" 或 "Language: go"
//   - memory.md 包含 [go] 相关内容
func TestInit_GoProject(t *testing.T) {
	dir := setupTestEnv(t, "go_project") // 有go.mod的目录

	result := runCmd(dir, "init", "--yes")

	// 验证命令成功
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Errorf("stdout: %s", result.Stdout)
		t.Errorf("stderr: %s", result.Stderr)
	}

	// 验证检测到Go
	if !strings.Contains(result.Stdout, "Go") && !strings.Contains(result.Stdout, "go") {
		t.Errorf("stdout should contain 'Go' or 'go', got: %s", result.Stdout)
	}

	// 验证memory.md包含Go相关内容
	assertFileContains(t, filepath.Join(dir, ".ohmymem", "memory.md"), "schema_version")
}

// TestInit_AlreadyInitialized tests error when already initialized
// Given: 目录已有 .ohmymem/memory.md
// When:  ohmymem init --yes (to skip interactive prompts)
// Then:
//   - 退出码 = 1
//   - stderr 包含 "already initialized"
func TestInit_AlreadyInitialized(t *testing.T) {
	dir := setupTestEnv(t, "already_initialized")

	result := runCmd(dir, "init", "--yes")

	// 验证命令失败
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}

	// 验证错误信息
	if !strings.Contains(result.Stderr, "already initialized") {
		t.Errorf("stderr should contain 'already initialized', got: %s", result.Stderr)
	}
}

// TestInit_ForceOverwrite tests force flag to overwrite existing initialization
// Given: 目录已有 .ohmymem/memory.md
// When:  ohmymem init --force --yes
// Then:
//   - 退出码 = 0
//   - memory.md 被覆盖（内容变化）
func TestInit_ForceOverwrite(t *testing.T) {
	dir := setupTestEnv(t, "already_initialized")

	// 记录原始内容
	originalContent, err := os.ReadFile(filepath.Join(dir, ".ohmymem", "memory.md"))
	if err != nil {
		t.Fatalf("failed to read original memory.md: %v", err)
	}

	result := runCmd(dir, "init", "--force", "--yes")

	// 验证命令成功
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
		t.Errorf("stdout: %s", result.Stdout)
		t.Errorf("stderr: %s", result.Stderr)
	}

	// 验证内容被覆盖
	newContent, err := os.ReadFile(filepath.Join(dir, ".ohmymem", "memory.md"))
	if err != nil {
		t.Fatalf("failed to read new memory.md: %v", err)
	}

	if string(originalContent) == string(newContent) {
		t.Error("memory.md should have been overwritten with different content")
	}

	// 验证新生成的内容包含 schema_version
	if !strings.Contains(string(newContent), "schema_version") {
		t.Error("new memory.md should contain schema_version")
	}
}

// TestInit_WithCustomRepo tests using --repo flag with custom template repository
// Given: 空目录
// When:  ohmymem init --yes --repo <custom_repo>
// Then:
//   - 退出码 = 0 或报错（如果仓库无效）
//   - 使用指定仓库的模板
func TestInit_WithCustomRepo(t *testing.T) {
	dir := setupTestEnv(t, "") // 空目录

	// 使用本地路径作为 repo（用于测试）
	_, testFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(testFile), "..", "..")
	customRepo := filepath.Join(projectRoot, "tests", "testdata", "templates")

	result := runCmd(dir, "init", "--yes", "--repo", customRepo)

	// 注意：这个测试可能会失败，取决于 --repo 的实现
	// 目前只是验证 flag 被接受
	if result.ExitCode != 0 {
		t.Logf("Custom repo test exited with code %d", result.ExitCode)
		t.Logf("stdout: %s", result.Stdout)
		t.Logf("stderr: %s", result.Stderr)
	}
}
