# Agent Manifest Workflow

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
