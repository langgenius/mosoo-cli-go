# Setup Reference

First-time setup is handled by Bootstrap.

Run:

```sh
curl -fsSL https://install.mosoo.ai/codex | bash
```

For an auditable install flow:

```sh
curl -fsSL https://install.mosoo.ai/codex -o install-mosoo-codex.sh
less install-mosoo-codex.sh
bash install-mosoo-codex.sh
```

By default, Bootstrap is interactive and asks for `y` or `n` before each
high-impact step, such as installing the CLI, installing the Skill, writing
config, running login, or running optional Cloudflare checks.

For automation:

```sh
curl -fsSL https://install.mosoo.ai/codex | bash -s -- --yes
```

For an execution preview:

```sh
curl -fsSL https://install.mosoo.ai/codex | bash -s -- --dry-run
```

Then verify:

```sh
mosoo --version
mosoo doctor --json
```

Bootstrap may install or update the Mosoo CLI, install or update this Mosoo
Skill, guide login, write initial config, and run readiness checks. Public
bootstrap defaults to the Mosoo cloud target at `https://api.mosoo.ai`; pass
`--target local` for a local development API. It should not make Cloudflare or
Wrangler a default prerequisite.

Release installers can set `MOSOO_CLI_VERSION` to make Bootstrap fail when the
installed binary does not report the expected `mosoo --version` build metadata.
