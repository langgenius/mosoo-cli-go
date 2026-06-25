# Publish

This directory contains source files for externally fetchable Mosoo CLI distribution artifacts.

- `installers/`: installer scripts exposed through stable install URLs.
  `installers/codex` is the source for `https://install.mosoo.ai/codex`.
- `manifests/`: machine-readable release metadata consumed by installers.
  `manifests/codex.example.json` is the route template for configuring the
  public `install.mosoo.ai` endpoints.
- `skills/`: skill packages distributed by bootstrap or installer flows. The
  Mosoo Skill entrypoint is `skills/mosoo/SKILL.md`; the hand-maintained CLI
  guide lives at `skills/mosoo/references/cli.md`, while generated CLI command
  indexes are rendered into `skills/mosoo/references/cli/`.

Keep repository-internal code generation and export scripts under `scripts/`.
