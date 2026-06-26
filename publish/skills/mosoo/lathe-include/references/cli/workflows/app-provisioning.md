# Agent App Provisioning Workflow

Use this workflow when a task starts from App and Agent setup instead of an
already published Agent. It is a product workflow assembled from generated
commands plus the env and Public Thread workflows.

First select the App boundary. For a new application, create a new App. Reuse
an existing App only when the user explicitly supplied its `appId`, or when its
name and purpose exactly match the current application after live verification.
Existing credentials or a working model are not sufficient reasons to reuse an
App.

After the App boundary is selected, create the Agent and publish it. Run the
generated commands in order and save each returned `appId` and `agentId` before
moving to the next step:

```sh
mosoo console apps app-list --organization-id <organization-id> -o json
mosoo console apps create-app --input-organization-id <organization-id> --input-name <app-name> -o json
mosoo console agents create-agent --file create-agent.json -o json
mosoo console agents publish-agent --input-app-id <app-id> --input-agent-id <agent-id> -o json
```

After publish, continue through the related workflow files instead of
duplicating their commands here:

1. Write backend or Worker env values with `public-api-tokens.md`.
2. Upload attachments only when the first thread requires files; use
   `thread-files.md`.
3. Run a smoke test by creating a Public Thread and waiting for final output
   with `thread-run.md`.
4. If Agent configuration needs a follow-up change, round-trip it through
   `agent-manifest.md`.

Use `mosoo commands show <path...> --json` before each generated command to
confirm body shape and required flags. Prefer `--file` for large Agent create
bodies. If a step fails or times out, inspect state with `console apps app-list`,
`console agents accessible-agent-list`, or `console agents agent` before
retrying; do not recreate resources until the current remote state is known.
