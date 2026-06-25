# CLI Reference

Generated from Lathe's mosoo CLI Skill output during `make build`.

## Runtime State

Run:

```sh
mosoo doctor --json
```

Use the result to decide whether the current task targets local Mosoo runtime or
Mosoo cloud runtime before running API commands.

## Command Selection

Use generated CLI commands for Mosoo resource operations, and use
`references/api.md` for application code that calls an already published Agent.
Do not invent a wrapper command when the generated catalog already exposes the
operation.

For a new App, Agent creation, publishing, credential setup, or Console/API
inspection, search the generated catalog first. For app environment files only,
derive `MOSOO_API_BASE`, `MOSOO_AGENT_ID`, and `MOSOO_API_TOKEN` from the
published Agent/API contract instead of creating new resources.

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

## Agent Manifest Workflow

Prefer the product workflow commands for editable Agent manifest YAML:

```sh
mosoo agent manifest probe --app-id <app-id> --agent-id <agent-id> --out agent.yaml
mosoo agent manifest diff --app-id <app-id> --agent-id <agent-id> --file agent.yaml
mosoo agent manifest apply --app-id <app-id> --agent-id <agent-id> --file agent.yaml --dry-run
mosoo agent manifest apply --app-id <app-id> --agent-id <agent-id> --file agent.yaml
```

`probe` reads the current remote manifest and writes YAML for editing or version
control. It also has a `pull` alias. `diff` performs a local field-level diff
between the local YAML target state and the current remote state.

`apply` always fetches the current remote manifest before writing, treats the
local YAML as the intended patch, preserves remote fields omitted from the YAML,
and then calls the raw `updateAgentConfig` operation. Use `--dry-run` first to
show the field-level changes without writing.

The raw generated `console agents manifest` and `console agents update-config`
commands are hidden from normal discovery. Use them only for low-level API
inspection with `mosoo commands --include-hidden --json`.

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
