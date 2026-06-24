# Publish

This directory contains source files for externally fetchable Mosoo CLI distribution artifacts.

- `installers/`: installer scripts exposed through stable install URLs.
- `manifests/`: machine-readable release metadata consumed by installers.
- `skills/`: skill packages distributed by bootstrap or installer flows. The
  Mosoo Skill entrypoint is `skills/mosoo/SKILL.md`; generated CLI reference
  material is rendered into `skills/mosoo/references/cli.md` and
  `skills/mosoo/references/cli/`.

Keep repository-internal code generation and export scripts under `scripts/`.
