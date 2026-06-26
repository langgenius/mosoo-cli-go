# CLI Reference

Generated from Lathe's mosoo CLI Skill output during `make build`.

Use this reference when a user asks you to operate `mosoo`, inspect its API
commands, or find the right generated command for an API task.

## Runtime State

Run `mosoo doctor --json` before assuming whether the task targets local Mosoo
runtime or Mosoo cloud runtime.

Console GraphQL and console REST commands use the `/api` surface. Public Thread
API commands use the `/api/v1` surface. For target and hostname examples, read
`references/cli/host-context.md`.

## Command Selection

Use generated CLI commands for Mosoo resource operations, and use
`references/api.md` for application code that calls an already published Agent.
Do not invent a wrapper command when the generated catalog already exposes the
operation.

For a new App, Agent creation, publishing, credential setup, or Console/API
inspection, search the generated catalog first. For app environment files only,
derive `MOSOO_API_BASE`, `MOSOO_AGENT_ID`, and `MOSOO_API_TOKEN` from the
published Agent/API contract instead of creating new resources.

## Workflow Map

Use these workflow files only when the task needs that operation:

- `references/cli/workflows/app-provisioning.md`: create or select an App,
  create an Agent, publish it, and carry `appId` / `agentId` forward.
- `references/cli/workflows/public-api-tokens.md`: prepare
  `MOSOO_API_BASE`, `MOSOO_AGENT_ID`, and `MOSOO_API_TOKEN` for a backend or
  Worker integration.
- `references/cli/workflows/thread-files.md`: upload and attach files to a
  Public Thread.
- `references/cli/workflows/thread-run.md`: create or continue a Public Thread,
  wait for final output, and inspect transcripts.
- `references/cli/workflows/agent-manifest.md`: pull, diff, and apply Agent
  manifest YAML safely.

For a backend or Worker integration with a published Agent, the usual order is:

1. Resolve runtime and hosts with `Runtime State` and `Host Context`.
2. Provision or select the App and Agent with `app-provisioning.md`.
3. Prepare backend environment values with `public-api-tokens.md`.
4. Upload files only when the thread needs attachments; use `thread-files.md`.
5. Create or continue the thread with `thread-run.md`.
6. Edit Agent configuration only through `agent-manifest.md`.

Carry these handoff values between workflow files: `appId`, `agentId`,
`threadId`, `fileId`, env file path, and manifest file path. If any value is
missing, return to the workflow that produces it instead of guessing.

## Discovery Protocol

1. Search for candidates with `mosoo search "<intent>" --json`; use `--limit`
   when needed. Search is only candidate discovery.
2. Inspect the exact command with `mosoo commands show <path...> --json` before
   executing an unfamiliar command.
3. If the command detail has `auth.required=true`, run
   `mosoo auth status --hostname <host>` before execution. Use
   `http.default_hostname` when present unless the user provides `--hostname` or
   `$MOSOO_HOST`.
4. Execute only after flags, body, auth, HTTP path, and output hints are clear
   from `commands show`.

## General Commands

- `mosoo commands --json`: full generated command catalog.
- `mosoo commands --include-hidden --json`: include hidden generated commands.
- `mosoo commands show <path...> --json`: source of truth for one command.
- `mosoo commands schema --json`: catalog schema version for parser
  compatibility.
- `mosoo search "<intent>" --json`: ranked candidate commands.

## References

- Read `references/cli/catalog.md` for the command discovery protocol and
  catalog field meanings.
- Read `references/cli/modules/console.md` for the `console` module command
  index.
- Read `references/cli/modules/console-rest.md` for the `console-rest` module
  command index.
- Read `references/cli/modules/public-thread-api.md` for the
  `public-thread-api` module command index.

## Rules

- Do not guess flags or request body shape from command names.
- Do not execute directly from search results; confirm with `commands show`
  first.
- Prefer `-o json` for machine-readable command output unless the user asks for
  human-readable output.
- Use `--file`, `--set`, or `--set-str` for JSON request bodies according to
  `commands show` body requirements.
