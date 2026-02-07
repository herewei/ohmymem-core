## Constraints

<!-- template-entry, tag: [echo, context], source: go-echo -->
* **[echo, context]** Use Echo's `echo.Context` for HTTP operations, access `context.Context` via `c.Request().Context()` (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [echo, middleware], source: go-echo -->
* **[echo, middleware]** Middleware must call `next(c)` to continue the chain, or return error to halt (*理由: Template default*)
<!-- entry-end -->

## Decisions

## Patterns

<!-- template-entry, tag: [echo, routing], source: go-echo -->
* **[echo, routing]** Group routes by resource using `e.Group("/resource")` for better organization (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [echo, binding], source: go-echo -->
* **[echo, binding]** Use Echo's `c.Bind()` for request body parsing with struct tags for validation (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [echo, response], source: go-echo -->
* **[echo, response]** Use `c.JSON()` for JSON responses with proper status codes (*理由: Template default*)
<!-- entry-end -->

## Anti-Patterns

<!-- template-entry, tag: [echo, raw], source: go-echo -->
* **[echo, raw]** Avoid using raw `http.ResponseWriter`; use Echo's context methods instead (*理由: Template default*)
<!-- entry-end -->