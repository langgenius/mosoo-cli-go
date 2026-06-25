# AGENTS.md

## Repository Rules

- Do not proactively use browser-use or browser automation unless the user asks for it.
- Use subagents when they can safely parallelize discovery or verification without losing repository context.
- Branch names, commit messages, and PR titles must follow Conventional Commits semantics.
- Commit messages must at least use `type(scope): subject`, for example `feat(cli): add workflow recipes`.
- Use these commit types: `fix`, `feat`, `docs`, `test`, `refactor`, and `chore`.
- Use `!` only for intentional breaking changes.
- Keep PR scope focused, self-assign issues and PRs, include verification results, and explicitly call out generated files, GraphQL codegen, or DB baseline updates when present.

## Lathe And Generated Files

This repository is a codegen wrapper around Mosoo API specs and Lathe. Do not blindly edit Lathe-generated outputs to fix behavior, docs, examples, or command metadata.

- Treat `internal/generated/**`, `cmd/mosoo/cli.yaml`, `specs/sources.yaml`, and `publish/skills/mosoo/references/cli/catalog.md` plus `publish/skills/mosoo/references/cli/modules/*.md` as generated outputs.
- Prefer changing the source of generation, then regenerate with the existing pipeline.
- For command help, examples, notes, prerequisites, and known errors, use the overlay pipeline: update `scripts/render-overlays.ts` and/or the relevant `overlays/*.yaml` source path, then run the build pipeline.
- `publish/skills/mosoo/references/cli.md` is generated from `publish/skills/mosoo/lathe-include/references/cli.md`. Put high-level CLI workflow guidance in the include file, then regenerate.
- If a required change cannot be expressed through specs, overlays, `cli.yaml`, or existing scripts, first inspect Lathe's supported extension points before editing generated files directly.
- Generated diffs are acceptable only when they are the intentional output of updated specs, overlays, scripts, or Lathe configuration. Do not include incidental generated churn in unrelated PRs.

## Lathe Usage In This Repo

This repo pins Lathe through `go.mod` and builds a repo-local binary at `.cache/bin/lathe`. Do not depend on a globally installed `lathe` from `PATH`.

- Build the local Lathe tool with `make tools`.
- The normal repo flow is `make build`, not a bare `lathe bootstrap`.
- `make build` exports Mosoo specs, renders `specs/sources.yaml`, renders `overlays/*.yaml`, runs Lathe codegen, renders published CLI references, and builds `bin/mosoo`.
- The Lathe command used by the build is:

```sh
.cache/bin/lathe codegen \
  -sources specs/sources.yaml \
  -cache .cache \
  -overlay overlays
```

`cli.yaml` sets `skill.root: .cache/lathe-skill`, so Lathe writes the generated full Skill to `.cache/lathe-skill/mosoo`. `scripts/render-publish-skill.ts` then copies `references/cli.md` into `publish/skills/mosoo/references/cli.md` and generated `references/catalog.md` plus `references/modules/*.md` into `publish/skills/mosoo/references/cli/`.

Lathe's relevant commands and flags:

- `lathe specsync`: sync pinned upstream API specs into the local cache.
- `lathe codegen`: generate runtime command specs and optional Skill files.
- `lathe bootstrap`: equivalent to `lathe specsync` plus `lathe codegen`; avoid using it directly here unless you have checked it matches the Makefile pipeline.
- `lathe version`: print version information.
- `lathe codegen -manifest <path>`: choose the `cli.yaml` path, default `cli.yaml`.
- `lathe codegen -sources <path>`: choose the `sources.yaml` path, default `specs/sources.yaml`.
- `lathe codegen -cache <dir>`: choose the cache root, default `$LATHE_SPECS_CACHE` or `.cache`.
- `lathe codegen -overlay <dir>`: load `<module>.yaml` overlays from a directory.
- `lathe codegen -skill-root <dir>`: write generated Skill output there; empty disables Skill generation.
- `lathe codegen -skill-include <dir>`: merge repo-local Skill resources into generated Skill files.

Overlay files must be named for the module key in `specs/sources.yaml`, such as `overlays/console.yaml`, `overlays/consolerest.yaml`, and `overlays/threads.yaml`. Overlays rewrite generated command names, help text, aliases, examples, groups, params, hidden state, ignored commands, notes, prerequisites, and known errors at codegen time. The runtime does not read overlays.

## Build And Verification

- Use `make build` to regenerate specs, overlays, Lathe output, published CLI references, and `bin/mosoo`.
- Use `go test ./...` for Go behavior changes.
- Before finalizing a change, inspect `git status --short` and separate intentional source changes from generated outputs.
