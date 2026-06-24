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
2. For normal Mosoo operations, read `references/cli.md`.
3. For application code that calls a published Mosoo Agent, read
   `references/api.md`.
4. For missing first-time setup, read `references/setup.md` and ask the user to
   run Bootstrap.

## Rules

- Do not implement a replacement planner, tool runner, memory system, sandbox,
  model loop, lifecycle manager, or provider integration when the task is to use
  a Mosoo Agent.
- Do not require Cloudflare or Wrangler for basic Mosoo setup.
- Prefer machine-readable CLI output such as `--json` before making environment
  or auth decisions.
