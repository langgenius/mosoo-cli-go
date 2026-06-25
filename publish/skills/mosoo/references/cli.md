# CLI

This file is the hand-maintained CLI routing guide for Mosoo. `make build`
refreshes generated command indexes under `references/cli/` without overwriting
this guide.

## When To Use CLI

Use `mosoo` for Mosoo resource operations: setup checks, auth state, App and
Agent provisioning, publishing, credential setup, Console/API inspection, and
public Thread API verification.

For application code that only calls an already published Agent, prefer
`references/api.md` instead of creating or changing Mosoo resources.

## Required Flow

1. Run `mosoo doctor --json` before assuming whether the user is on local mode,
   cloud mode, or a custom target.
2. Start with high-level Mosoo CLI workflows when they fit the request.
3. For generated API commands, search first:

   ```sh
   mosoo search "<intent>" --json
   ```

4. Inspect the exact command before running it:

   ```sh
   mosoo commands show <path...> --json
   ```

5. Execute only after flags, request body shape, auth requirements, hostname,
   and output format are clear from `commands show`.

## Generated References

- `references/cli/catalog.md`: generated catalog schema and command selection
  notes.
- `references/cli/modules/console.md`: generated Console GraphQL command index.
- `references/cli/modules/console-rest.md`: generated Console REST command
  index.
- `references/cli/modules/public-thread-api.md`: generated public Thread API
  command index.

## Rules

- Do not guess flags or request body shape from command names.
- Do not execute directly from search results; confirm with `commands show`.
- Prefer machine-readable output such as `--json` or `-o json`.
- Use `--file`, `--set`, or `--set-str` according to `commands show` body
  requirements.
