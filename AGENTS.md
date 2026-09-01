## Validating changes

* After making changes, run `task --force go-validate generate`. Iterate until it passes end to end: fix every reported failure and re-run it. Do not consider a change complete until it is green.
* While iterating on one package, `go test ./internal/<pkg>/... -run TestName` is faster than the full suite.

## Project structure

* `content/markdown/`: page sources, **human-written only**, not agent-written or agent-inspired. No help needed beyond spelling fixes.
* `internal/`: Go source: core packages vs. `internal/generator/*` (see Core packages).
* `styles/`: Tailwind CSS input.
* `target/build/`: generated output (gitignored).

## Investigation

* Only explore files that are **explicitly mentioned** in the task or conversation. Do not browse or search the rest of the codebase unless specifically instructed.
* If you are unsure about an API, function, library, or behavior, write small test scripts, run them, print the output, and use the results to make informed decisions before proceeding.

## Gotchas

* Tailwind logical margin/padding utilities (`my-*`, `mx-*`) can lose the CSS cascade to Typography-plugin rules that use physical properties. Use `mt-*`/`mb-*`/`ml-*`/`mr-*` instead.
* goldmark has no `html.WithUnsafe()`: raw HTML, including `<!-- key: value -->` metadata comments, is stripped to `<!-- raw HTML omitted -->` in rendered output.

## Core packages

* Every `internal/` package except `internal/generator/*`. They hold **only proven, future-proof, documented code**. Extracted from `internal/generator/*` once code earns that status, never preemptively.
* The surface they expose is deliberate and carefully human-reviewed. **Always ask for human approval before migrating anything here.**

## Documentation

* Do not add comments by default. Only add Go doc comments (package docs, exported identifier docs) in core packages. Do not add doc comments in `internal/generator/*` (page and site generation logic).
* Every generator-owned feature **usable from within markdown content** must be documented in `README.md`, **idiomatic and as concise as possible**: tables over prose, no redundant explanation.
