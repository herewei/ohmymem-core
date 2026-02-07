# End to End Test Manual

This document describes how to run and manually verify the end-to-end (E2E) tests under `tests/e2e/`.

## Scope

The E2E tests exercise the `ohmymem init` command against temporary project directories created from fixtures in `tests/e2e/testdata/`. The tests validate file creation, symlink behavior, and key content in the generated `.ohmymem/memory.md` file.

## Prerequisites

- macOS or Linux shell
- Go toolchain installed
- Build output `ohmymem` present at repository root

Build the binary:

```bash
make build-darwin
```

Or build directly:

```bash
go build -o ohmymem .
```

## How The E2E Tests Work

- The test helper locates the `ohmymem` binary in the repository root and executes it with arguments.
- Each test runs in a temporary directory created under the OS temp folder.
- If a fixture is specified, `tests/e2e/testdata/<fixture>` is copied into the temp directory.
- After each test, the temp directory is removed.

## Fixtures

Located in `tests/e2e/testdata/`.

- `already_initialized/` includes `.ohmymem/memory.md` and `AGENTS.md` to simulate a pre-initialized project.
- `go_project/` includes a minimal `go.mod` to trigger Go detection.
- `empty_project/` exists but is not currently used by any test.

## Running The Automated E2E Tests

Run all E2E tests:

```bash
go test ./tests/e2e -v
```

Run a single test:

```bash
go test ./tests/e2e -run TestInit_EmptyDirectory -v
```

## Manual Test Checklist

Use this checklist to manually confirm behavior without `go test`.

### 1) Init In Empty Directory

Setup:

```bash
TMP_DIR=$(mktemp -d)
/absolute/path/to/ohmymem init --yes
```

Run in the empty directory:

```bash
(cd "$TMP_DIR" && /absolute/path/to/ohmymem init --yes)
```

Expected:

- Exit code `0`
- Stdout contains `Initialization complete`
- `.ohmymem/memory.md` exists
- `AGENTS.md` exists
- `.cursorrules` is a symlink pointing to `AGENTS.md`
- `.ohmymem/memory.md` contains `schema_version`
- `.ohmymem/memory.md` contains `## Constraints`
- `.ohmymem/memory.md` contains `## Decisions`
- `.ohmymem/memory.md` contains `## Patterns`
- `.ohmymem/memory.md` contains `## Anti-Patterns`

### 2) Init In Go Project

Setup:

```bash
TMP_DIR=$(mktemp -d)
cp tests/e2e/testdata/go_project/go.mod "$TMP_DIR/go.mod"
```

Run:

```bash
(cd "$TMP_DIR" && /absolute/path/to/ohmymem init --yes)
```

Expected:

- Exit code `0`
- Stdout contains `Go` or `go`
- `.ohmymem/memory.md` created and contains `schema_version`

### 3) Init When Already Initialized (No Force)

Setup:

```bash
TMP_DIR=$(mktemp -d)
cp -R tests/e2e/testdata/already_initialized/. "$TMP_DIR/"
```

Run:

```bash
(cd "$TMP_DIR" && /absolute/path/to/ohmymem init --yes)
```

Expected:

- Exit code `1`
- Stderr contains `already initialized`
- Existing `.ohmymem/memory.md` is unchanged

### 4) Init With Force Overwrite

Setup:

```bash
TMP_DIR=$(mktemp -d)
cp -R tests/e2e/testdata/already_initialized/. "$TMP_DIR/"
```

Record original contents:

```bash
cat "$TMP_DIR/.ohmymem/memory.md"
```

Run:

```bash
(cd "$TMP_DIR" && /absolute/path/to/ohmymem init --force --yes)
```

Expected:

- Exit code `0`
- `.ohmymem/memory.md` content changes
- New content includes `schema_version`

### 5) Init With Custom Repo

Setup:

```bash
TMP_DIR=$(mktemp -d)
CUSTOM_REPO="$(pwd)/tests/e2e/testdata"
```

Run:

```bash
(cd "$TMP_DIR" && /absolute/path/to/ohmymem init --yes --repo "$CUSTOM_REPO")
```

Expected:

- Exit code `0` or non-zero depending on `--repo` implementation
- If non-zero, the error should clearly describe why the repo could not be used

## Notes

- The E2E tests expect the `ohmymem` binary to be located at the repository root.
- All tests use temporary directories; they do not mutate repository state.
- If you see `permission denied` errors, verify the binary is executable: `chmod +x ohmymem`.
