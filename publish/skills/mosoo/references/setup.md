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

Then verify:

```sh
mosoo doctor --json
```

Bootstrap may install or update the Mosoo CLI, install or update this Mosoo
Skill, guide login, write initial config, and run readiness checks. It should
not make Cloudflare or Wrangler a default prerequisite.
