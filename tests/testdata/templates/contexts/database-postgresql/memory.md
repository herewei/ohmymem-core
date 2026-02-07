## Constraints

<!-- template-entry, tag: [database, time], source: database-postgresql -->
* **[database, time]** All time fields must be stored in UTC using `TIMESTAMP WITH TIME ZONE` (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [database, security], source: database-postgresql -->
* **[database, security]** Use parameterized queries only, never string concatenation for SQL (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [database, transactions], source: database-postgresql -->
* **[database, transactions]** Wrap multi-step database operations in transactions (*理由: Template default*)
<!-- entry-end -->

## Decisions

## Patterns

<!-- template-entry, tag: [database, repository], source: database-postgresql -->
* **[database, repository]** Use Repository pattern for data access layer abstraction (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [database, naming], source: database-postgresql -->
* **[database, naming]** Use snake_case for table and column names (*理由: Template default*)
<!-- entry-end -->

## Anti-Patterns

<!-- template-entry, tag: [database, orm], source: database-postgresql -->
* **[database, orm]** Avoid ORM magic queries; prefer explicit SQL for clarity and performance (*理由: Template default*)
<!-- entry-end -->

<!-- template-entry, tag: [database, select-star], source: database-postgresql -->
* **[database, select-star]** Avoid `SELECT *`; explicitly list required columns (*理由: Template default*)
<!-- entry-end -->