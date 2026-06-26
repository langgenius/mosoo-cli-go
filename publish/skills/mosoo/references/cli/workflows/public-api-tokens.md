# Public API Tokens Workflow

Use this workflow when preparing backend or Worker environment values for a
published Agent integration.

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

Use `mosoo agent env export` or `mosoo agent env write --file <path>` to prepare
`MOSOO_API_BASE`, `MOSOO_AGENT_ID`, and `MOSOO_API_TOKEN` for backend or Worker
workflows. When `MOSOO_API_TOKEN` is unset, the helper uses the token from
`mosoo auth login` for the selected Public API host.
