# CLI Host Context Examples

Use this reference when a command example needs to be runnable against a known
Mosoo runtime. Start by resolving the active target:

```sh
mosoo doctor --json
```

`doctor` reports the selected target, the base URL, and the per-surface hosts.
Console GraphQL and console REST commands use the `/api` surface. Public Thread
API commands use the `/api/v1` surface.

## Target And Base URL

Use `--target local` or `--target cloud` when one of the built-in targets is the
intended runtime:

```sh
mosoo --target local -o json console apps app-list \
  --organization-id <organization-id>
```

Use `--target custom --base-url <service-root>` for a non-default deployment.
Pass the service root. The CLI derives `/api` and `/api/v1` from it:

```sh
mosoo --target custom --base-url http://127.0.0.1:8787 -o json \
  console apps app-list --organization-id <organization-id>
```

## Exact Surface Host

Use `--hostname <surface-host>` only when overriding one exact API surface. Pass
the full surface host, not the service root:

```sh
mosoo --hostname http://127.0.0.1:8787/api -o json \
  console apps app-list --organization-id <organization-id>

mosoo --hostname http://127.0.0.1:8787/api/v1 -o json \
  public-thread-api threads create --agent-id <agent-id> --file body.json
```

## Environment Override

`MOSOO_HOST` behaves like `--hostname`, but for the shell environment. Keep it
scoped to one command when possible so it does not leak into later commands for
another surface:

```sh
MOSOO_HOST=http://127.0.0.1:8787/api \
  mosoo -o json console apps app-list --organization-id <organization-id>
```

## Preflight Checks

Before running a copied command, inspect the command metadata and auth state for
the surface that will be used:

```sh
mosoo commands show console apps app-list --json
mosoo auth status --hostname http://127.0.0.1:8787/api
```
