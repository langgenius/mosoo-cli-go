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

High-frequency root commands such as `mosoo ls`, `mosoo run`, `mosoo add-key`,
and `mosoo create-agent` are Lathe-generated `shortcuts` for canonical generated
operations. Treat them as generated command entries, and confirm their exact
flags and body shape with `mosoo commands show <shortcut> --json`.

For a new App, Agent creation, publishing, credential setup, or Console/API
inspection, search the generated catalog first. For app environment files only,
derive `MOSOO_API_BASE`, `MOSOO_AGENT_ID`, and `MOSOO_API_TOKEN` from the
published Agent/API contract instead of creating new resources.

Use this reference when a user asks you to operate `mosoo`, inspect its API commands, or find the right generated command for an API task.

## Common Workflow Recipes

Use this section as the entry point for end-to-end Mosoo CLI tasks. It defines
workflow order and handoff values only; keep detailed command flags and request
shapes in the owning workflow sections below.

For a backend or Worker integration with a published Agent:

1. Resolve runtime and hosts with `Runtime State` and `Host Context`.
2. Provision or select the App and Agent with `Agent App Provisioning Workflow`.
3. Prepare backend environment values with `Public API Tokens`.
4. Upload files only when the thread needs attachments; use `Public Thread File Upload Workflow` and carry forward the returned `fileId`.
5. Create or continue the thread, wait for completion, and inspect output with `Public Thread Wait, Final Output, And Transcript Workflow`.
6. Edit Agent configuration only through `Agent Manifest Workflow`.

Carry these handoff values between workflow sections: `appId`, `agentId`,
`threadId`, `fileId`, env file path, and manifest file path. If any value is
missing, return to the section that produces it instead of guessing.

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

Use this workflow when a task starts from App and Agent setup instead of an
already published Agent. It is a product workflow assembled from generated
commands plus the env and Public Thread workflows below.

First create or reuse the App, create the Agent, and publish it. Run the
generated commands in order and save each returned `appId` and `agentId` before
moving to the next step:

```sh
mosoo console apps app-list --organization-id <organization-id> -o json
mosoo console apps create-app --input-organization-id <organization-id> --input-name <app-name> -o json
mosoo add-key --input-app-id <app-id> --input-vendor-id openai --input-name OpenAI --input-api-key-env OPENAI_API_KEY -o json
mosoo create-agent --file create-agent.json -o json
mosoo console agents publish-agent --input-app-id <app-id> --input-agent-id <agent-id> -o json
```

After publish, continue through the related workflow sections instead of
duplicating their commands here:

1. Write backend or Worker env values with `Public API Tokens`.
2. Upload attachments only when the first thread requires files; use `Public Thread File Upload Workflow`.
3. Run a smoke test by creating a Public Thread and waiting for final output with `Public Thread Wait, Final Output, And Transcript Workflow`.
4. If Agent configuration needs a follow-up change, round-trip it through `Agent Manifest Workflow`.

Use `mosoo commands show <path...> --json` before each generated command to
confirm body shape and required flags. Prefer `--file` for large Agent create
bodies. If a step fails or times out, inspect state with `console apps app-list`,
`console agents accessible-agent-list`, or `console agents agent` before
retrying; do not recreate resources until the current remote state is known.

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

## Public Thread Wait, Final Output, And Transcript Workflow

```sh
mosoo public-thread-api threads create --agent-id <agent-id> --file body.json --wait -o json
mosoo public-thread-api threads create --agent-id <agent-id> --file body.json --final-output
mosoo public-thread-api events wait --thread-id <thread-id> --final-output
mosoo public-thread-api threads transcript --thread-id <thread-id>
```

## Workflow

1. Search for candidates with `mosoo search "<intent>" --json`; use `--limit` when needed. Search is only candidate discovery.
2. Inspect the exact command with `mosoo commands show <path...> --json` before executing an unfamiliar command.
3. If the command detail has `auth.required=true`, run `mosoo auth status --hostname <host>` before execution. Use `http.default_hostname` when present unless the user provides `--hostname` or `$MOSOO_HOST`.
4. Execute only after flags, body, auth, HTTP path, and output hints are clear from `commands show`.

## Browser Login

Use `mosoo auth login --auth-type oauth --hostname <host> --provider google`
when the user needs browser-based Mosoo login for CLI access. The browser
session may come from Google OAuth or Mosoo's email login, but the CLI stores the
issued Mosoo API token as `auth_type: bearer` after authorization.

## Host Context

Use `mosoo doctor --json` first when the target is not explicit. It reports the
resolved target, base URL, and per-surface hosts. Console GraphQL and console
REST commands use the `/api` surface. Public Thread API commands use the
`/api/v1` surface.

Use `--target local` or `--target cloud` for built-in targets. Use `--target
custom --base-url <service-root>` for a non-default deployment so the CLI derives
the correct surface hosts. Use `--hostname <surface-host>` or `MOSOO_HOST` only
when overriding one exact surface host.

For runnable examples covering `--target`, `--base-url`, `--hostname`, and
`MOSOO_HOST`, read `references/cli/host-context.md`.

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
