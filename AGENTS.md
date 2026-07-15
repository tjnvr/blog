## Workflow

* Only explore files that are **explicitly mentioned** in the task or conversation. Do not browse or search the rest of the codebase unless specifically instructed.

* If you are unsure about an API, function, library, or behavior, write small test scripts, run them, print the output, and use the results to make informed decisions before proceeding.

## Validating changes

* After making changes, validate them by running `task --force validate generate`.

* Iterate until the command passes end to end: fix every reported failure and re-run it. Do not consider a change complete until `task --force validate generate` is green.
