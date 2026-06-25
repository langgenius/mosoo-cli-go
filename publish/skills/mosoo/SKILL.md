---
name: mosoo
description: >
  Use when a coding agent needs to work with Mosoo setup, local or cloud runtime
  state, Mosoo CLI operations, or app integration with a published Mosoo Agent.
---

# Mosoo

Treat Mosoo as the Agent runtime unless the user explicitly asks to build a
separate agent runtime.

## Workflow

1. Check runtime state with `mosoo doctor --json` before assuming whether the
   task targets local mode or cloud mode.
2. For application code that calls an already published Mosoo Agent, read
   `references/api.md`.
3. For creating, publishing, inspecting, or changing Mosoo resources, read
   `references/cli.md`, then follow its command-index links when command
   details are needed.
4. For missing first-time setup, read `references/setup.md` and ask the user to
   run Bootstrap.

## Routing

- Existing published Agent integration: do not create or publish anything; use
  `references/api.md` and app backend code.
- New app, Agent creation, publishing, credential setup, or Console/API
  inspection: use `references/cli.md`, then run `mosoo search ... --json` and
  `mosoo commands show <path...> --json` before executing generated commands.
- Agent configuration changes: follow the manifest round-trip workflow in
  `references/cli.md`; pull the current Agent manifest/YAML first, edit it
  locally, and submit the complete updated config.
- App env file only: derive `MOSOO_API_BASE`, `MOSOO_AGENT_ID`, and
  `MOSOO_API_TOKEN` from the published Agent/API contract; do not create
  Mosoo resources unless the user asked for that.
- Published Agent verification: use the public Thread API contract in
  `references/api.md` or the generated public-thread-api commands in
  `references/cli.md`.

## Rules

- Do not implement a replacement planner, tool runner, memory system, sandbox,
  model loop, lifecycle manager, or provider integration when the task is to use
  a Mosoo Agent.
- Do not require Cloudflare or Wrangler for basic Mosoo setup.
- Prefer machine-readable CLI output such as `--json` before making environment
  or auth decisions.
- Do not construct Agent config update payloads from memory or guessed fields.
  Preserve the existing manifest values unless the user explicitly asks to
  change them.
