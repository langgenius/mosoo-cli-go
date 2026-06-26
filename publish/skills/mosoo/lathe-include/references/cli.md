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

## Public API Tokens

`MOSOO_API_TOKEN` is a server-side credential for application backends or
Workers that call a published Agent through the Public API. Do not expose it in
browser or frontend code.

Users can create multiple Mosoo API tokens and assign each token an
application-level purpose or logical scope in their own app backend. For
example, an app can keep one token for a production Agent integration, another
token for smoke tests, and its own metadata that decides which app users or
workflows may use each token.

Mosoo validates the token. The calling app is responsible for selecting the
right token, storing any app-level scope metadata, and enforcing business rules
before calling Mosoo. For multi-user apps, keep tenant and user mapping in the
app backend; a single token does not switch Mosoo identity based on request
payload fields.

When writing app env files, store token values only in backend or Worker
environment files and redact token values in logs, examples, and command
output.

Use `mosoo agent env export` or `mosoo agent env write --file <path>` to prepare `MOSOO_API_BASE`, `MOSOO_AGENT_ID`, and `MOSOO_API_TOKEN` for backend or Worker workflows; when `MOSOO_API_TOKEN` is unset, the helper uses the token from `mosoo auth login` for the selected Public API host.

## Agent App Provisioning Workflow

For App and Agent setup, run the generated commands in order and save each
returned ID before moving to the next step:

```sh
mosoo console apps app-list --organization-id <organization-id> -o json
mosoo console apps create-app --input-organization-id <organization-id> --input-name <app-name> -o json
mosoo console agents create-agent --file create-agent.json -o json
mosoo console agents publish-agent --input-app-id <app-id> --input-agent-id <agent-id> -o json
```

Use `mosoo commands show <path...> --json` before each command to confirm body
shape and required flags. Prefer `--file` for large Agent create bodies. If a
step fails or times out, inspect state with `console apps app-list`, `console
agents accessible-agent-list`, or `console agents agent` before retrying.

## Public Thread File Upload Workflow

For Public Thread file uploads, run the generated commands in order and save the
returned `fileId` before moving to the next step:

```sh
mosoo console-rest files create-upload --file upload.json -o json
mosoo console-rest files upload-content --file-id <file-id> --file content-body.json -o json
mosoo console-rest files complete-upload --file-id <file-id> --file complete.json -o json
mosoo public-thread-api files add --thread-id <thread-id> --set fileId=<file-id> -o json
```

Use `mosoo commands show <path...> --json` before each command to confirm body
shape and host selection. If a step fails or times out, inspect state with
`console-rest files get-upload` before retrying. Use `console-rest files
abort-upload` only for a pending upload that should not be completed.

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

When changing prompts, models, providers, tools, runtime, or environment
settings, do not reconstruct an update payload from memory or guessed defaults.
Round-trip the current manifest, edit only the requested fields, and preserve
unchanged values such as `environmentId`, runtime, provider, model, skill IDs,
MCP server IDs, and `providerOptions`.

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
