## Constraints

<!-- template-entry, tag: [api, naming], source: rest-api -->
* **[api, naming]** Use plural nouns for resource endpoints: `/users`, `/orders`, not `/user`, `/order` (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [api, methods], source: rest-api -->
* **[api, methods]** Use correct HTTP methods: GET (read), POST (create), PUT (full update), PATCH (partial), DELETE (*理由: Template default*)
<!-- entry-end -->

## Decisions

## Patterns

<!-- template-entry, tag: [api, status], source: rest-api -->
* **[api, status]** Return appropriate status codes: 200 (OK), 201 (Created), 204 (No Content), 400 (Bad Request), 404 (Not Found), 500 (Server Error) (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [api, pagination], source: rest-api -->
* **[api, pagination]** Use cursor-based or offset pagination for list endpoints with `limit` and `offset`/`cursor` params (*理由: Template default*)
<!-- entry-end -->

## Anti-Patterns

<!-- template-entry, tag: [api, verbs], source: rest-api -->
* **[api, verbs]** Avoid verbs in URLs like `/getUsers` or `/createOrder`; use HTTP methods instead (*理由: Template default*)
<!-- entry-end -->