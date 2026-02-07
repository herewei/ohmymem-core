## Constraints

<!-- template-entry, tag: [go, version], source: go -->
* **[go, version]** Go version must be 1.21 or higher (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [go, logging], source: go -->
* **[go, logging]** Use `log/slog` for structured logging, never `fmt.Println` or `log` package directly (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [go, error], source: go -->
* **[go, error]** Errors must be wrapped with context using `fmt.Errorf("operation: %w", err)` (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [go, context], source: go -->
* **[go, context]** Functions that may block or do I/O must accept `context.Context` as first parameter (*理由: Template default*)
<!-- entry-end -->

## Decisions

## Patterns

<!-- template-entry, tag: [go, naming], source: go -->
* **[go, naming]** Use MixedCaps for exported names, mixedCaps for unexported (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [go, testing], source: go -->
* **[go, testing]** Use table-driven tests for comprehensive coverage (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [go, interface], source: go -->
* **[go, interface]** Define interfaces where they are used, not where they are implemented (*理由: Template default*)
<!-- entry-end -->

## Anti-Patterns

<!-- template-entry, tag: [go, init], source: go -->
* **[go, init]** Avoid `init()` functions; use explicit initialization for better control and testability (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [go, panic], source: go -->
* **[go, panic]** Never use `panic()` for regular error handling; reserve for truly unrecoverable situations (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [go, global], source: go -->
* **[go, global]** Don't store mutable state in global variables; use dependency injection (*理由: Template default*)
<!-- entry-end -->