# CLI Reference

Generated from Lathe's mosoo CLI Skill output during `make build`.

## Runtime State

Run:

```sh
mosoo doctor --json
```

Use the result to decide whether the current task targets local Mosoo runtime or
Mosoo cloud runtime before running API commands.

Use this reference when a user asks you to operate `mosoo`, inspect its API commands, or find the right generated command for an API task.

## Workflow

1. Search for candidates with `mosoo search "<intent>" --json`; use `--limit` when needed. Search is only candidate discovery.
2. Inspect the exact command with `mosoo commands show <path...> --json` before executing an unfamiliar command.
3. If the command detail has `auth.required=true`, run `mosoo auth status --hostname <host>` before execution. Use `http.default_hostname` when present unless the user provides `--hostname` or `$MOSOO_HOST`.
4. Execute only after flags, body, auth, HTTP path, and output hints are clear from `commands show`.

## General Commands

- `mosoo commands --json`: full generated command catalog.
- `mosoo commands --include-hidden --json`: include hidden generated commands.
- `mosoo commands show <path...> --json`: source of truth for one command.
- `mosoo commands schema --json`: catalog schema version for parser compatibility.
- `mosoo search "<intent>" --json`: ranked candidate commands.

## References

- Read `references/cli/catalog.md` for the command discovery protocol and catalog field meanings.
- Read `references/cli/modules/console.md` for the `console` module command index.
- Read `references/cli/modules/console-rest.md` for the `console-rest` module command index.
- Read `references/cli/modules/public-thread-api.md` for the `public-thread-api` module command index.

## Rules

- Do not guess flags or request body shape from command names.
- Do not execute directly from search results; confirm with `commands show` first.
- Prefer `-o json` for machine-readable command output unless the user asks for human-readable output.
- Use `--file`, `--set`, or `--set-str` for JSON request bodies according to `commands show` body requirements.
