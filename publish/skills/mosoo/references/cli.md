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

## Agent Config Updates

Agent config updates are full-manifest updates. The service intentionally
expects the complete Agent manifest/YAML on each update so coding agents keep a
single consistent source of truth.

Before changing an Agent prompt, model, provider, tools, runtime, or
environment:

1. Pull the current Agent manifest with `mosoo console agents agent-manifest`.
2. Save the manifest/YAML locally and edit only the fields requested by the
   user.
3. Preserve unchanged fields, including `environmentId`, runtime, provider,
   model, skill IDs, MCP server IDs, and `providerOptions`.
4. Submit the complete updated config; do not send a partial patch or a payload
   reconstructed from guessed defaults.
5. Pull the manifest again after the update and compare the changed fields.

If the current manifest cannot be pulled, stop and report the blocker instead
of inferring required fields from command names, old examples, or memory.

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
- Do not update Agent config by guessing required fields. Always round-trip the
  current Agent manifest/YAML and preserve unchanged values.
