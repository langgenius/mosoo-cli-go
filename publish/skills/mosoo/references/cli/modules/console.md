# Module `console`

## Source

- Backend: `graphql`
- Default hostname: `http://127.0.0.1:8787/api`
- Repository: https://github.com/langgenius/mosoo.git
- Pinned tag: `local-snapshot`
- Schema: `docs/graphql/console.graphql`
- Expose queries: `accessibleAgentList`, `agent`, `agentChannelBindingList`, `agentCostCard`, `agentEditorState`, `agentManifest`, `agentSessionDiagnostics`, `agentSessionList`, `agentSessionRetrieve`, `appCostCard`, `appDeploymentStatus`, `appEnvironmentList`, `appInfo`, `appList`, `appOverview`, `appSkillList`, `availableAgentModels`, `controlPlaneOverview`, `environment`, `exportAgentPackage`, `fileList`, `listSessionResources`, `mcpOAuthFlowStatus`, `mcpRegistry`, `organizationBillingCostCard`, `session`, `sessionList`, `sessionMessages`, `sessionProcessEvents`, `skillDetail`, `threadAgentSessionList`, `threadAgentSessionRetrieve`, `threadSessionMessages`, `threadSessionProcessEvents`, `vendorCredentialList`, `viewer`
- Expose mutations: `addSessionResource`, `archiveAgentSession`, `autoTitleSession`, `connectMcpBearer`, `createAgent`, `createAgentFork`, `createAgentSession`, `createApp`, `createAppMcpServer`, `createDiscordAgentChannelBinding`, `createEnvironmentFork`, `createLarkAgentChannelBinding`, `createSkillFork`, `createSlackAgentChannelBinding`, `createTelegramAgentChannelBinding`, `createVendorCredential`, `deleteAgent`, `deleteAgentChannelBinding`, `deleteAgentSession`, `deleteAppDeployment`, `deleteEnvironment`, `deleteMcpServer`, `deleteOwnedSkill`, `deleteVendorCredential`, `deployApp`, `importAgentPackage`, `onboardingBootstrap`, `pollLarkAgentChannelRegistration`, `pollWeChatAgentChannelPairing`, `prewarmAgentSession`, `publishAgent`, `recreateSandbox`, `removeSessionResource`, `renameApp`, `renameSession`, `resetAgentState`, `restartDriver`, `revokeMcpCredential`, `setAppDefaultEnvironment`, `setDefaultVendorCredential`, `setEnvironmentVariableValue`, `setMcpServerEnabled`, `setSystemAgentModel`, `startAgentRun`, `startLarkAgentChannelRegistration`, `startMcpOAuth`, `startWeChatAgentChannelPairing`, `testVendorCredential`, `unarchiveAgentSession`, `unpublishAgent`, `updateAgentConfig`, `updateProfile`, `updateVendorCredential`
- Group policies: `13`
- Selection policy: max depth `5`
- Resolved SHA: `local-snapshot`

## Agents

### `mosoo console agents accessible-agent-list`

- Summary: List accessible Agents
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Examples:
  - List Agents available to the signed-in user for one App and capture an Agent ID.
    Command: `mosoo console agents accessible-agent-list --app-id <app-id> -o json`
    Body shape: `{"appId":"\u003capp-id\u003e"}`
    Output ID path: `data.accessibleAgentList[0].id`
    Output list path: `data.accessibleAgentList`
    Follow-up commands:
      - `mosoo console agents agent --app-id <app-id> --agent-id <agent-id> -o json`

### `mosoo console agents agent`

- Summary: Agent
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--agent-id` (variable, required): agentId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents agent-editor-state`

- Summary: Agent editor state
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--agent-id` (variable, required): agentId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents create-agent`

- Summary: Create an Agent
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Shortcuts:
  - `mosoo create-agent`
- Flags:
  - `--input-description` (variable): input.description
  - `--input-kind` (variable, required, one of: pet|cattle): input.kind
  - `--input-model` (variable, required): input.model
  - `--input-name` (variable, required): input.name
  - `--input-prompt` (variable, required): input.prompt
  - `--input-provider` (variable, required): input.provider
  - `--input-runtime-id` (variable, required): input.runtimeId
  - `--input-skill-ids` (variable, required): input.skillIds
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Examples:
  - Create an Agent from a JSON body file
    Command: `mosoo console agents create-agent --file agent-create.json -o json`
    Body shape: `{"input":{"appId":"\u003capp-id\u003e","kind":"pet","model":"\u003cmodel\u003e","name":"Research Agent","prompt":"You research concise answers with citations.","provider":"\u003cprovider\u003e","runtimeId":"\u003cruntime-id\u003e","skillIds":[]}}`
    Output ID path: `data.createAgent.id`
    Follow-up commands:
      - `mosoo console agents agent --app-id <app-id> --agent-id <id> -o json`
      - `mosoo console agents publish-agent --input-app-id <app-id> --input-agent-id <id> -o json`

### `mosoo console agents create-agent-fork`

- Summary: Create an agent fork
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-kind` (variable, one of: pet|cattle): input.kind
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents delete-agent`

- Summary: Delete an agent
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents export-agent-package`

- Summary: Export an agent package
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--agent-id` (variable, required): agentId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents import-agent-package`

- Summary: Import an agent package
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-file-id` (variable, required): input.fileId
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents publish-agent`

- Summary: Publish an Agent
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console agents publish --input-app-id <app-id> --input-agent-id <agent-id> -o json`

### `mosoo console agents recreate-sandbox`

- Summary: Recreate a sandbox
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-affected-fields` (variable): input.affectedFields
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-apply-action-kind` (variable): input.applyActionKind
  - `--input-target-version-id` (variable): input.targetVersion.id
  - `--input-target-version-version-number` (variable): input.targetVersion.versionNumber
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents reset-agent-state`

- Summary: Reset an agent state
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-affected-fields` (variable): input.affectedFields
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-apply-action-kind` (variable): input.applyActionKind
  - `--input-target-version-id` (variable): input.targetVersion.id
  - `--input-target-version-version-number` (variable): input.targetVersion.versionNumber
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents restart-driver`

- Summary: Restart a driver
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-affected-fields` (variable): input.affectedFields
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-apply-action-kind` (variable): input.applyActionKind
  - `--input-target-version-id` (variable): input.targetVersion.id
  - `--input-target-version-version-number` (variable): input.targetVersion.versionNumber
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console agents unpublish-agent`

- Summary: Unpublish an agent
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--agent-id` (variable, required): agentId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## Apps

### `mosoo console apps app-deployment-status`

- Summary: App deployment status
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console apps app-list`

- Summary: List Apps
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--organization-id` (variable, required): organizationId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console apps list --organization-id <organization-id> -o json`

### `mosoo console apps app-overview`

- Summary: Show one App overview
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--agent-limit` (variable): agentLimit
  - `--credential-limit` (variable): credentialLimit
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console apps app-overview --app-id <app-id> -o json`

### `mosoo console apps control-plane-overview`

- Summary: Show control-plane overview
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Shortcuts:
  - `mosoo ls`
- Flags:
  - `--app-limit` (variable): appLimit
  - `--agent-limit` (variable): agentLimit
  - `--credential-limit` (variable): credentialLimit
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
  - Use this before lower-level app-list, accessible-agent-list, or vendor-credential-list when you need a CLI overview.
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console apps overview --app-limit 20 --agent-limit 20 --credential-limit 20 -o json`

### `mosoo console apps create-app`

- Summary: Create an app
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-name` (variable, required): input.name
  - `--input-organization-id` (variable, required): input.organizationId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Examples:
  - Create a minimal App and capture its id for follow-up commands.
    Command: `mosoo console apps create-app --input-organization-id <organization-id> --input-name "CLI Example App" -o json`
    Body shape: `{"input":{"name":"CLI Example App","organizationId":"\u003corganization-id\u003e"}}`
    Output ID path: `data.createApp.id`
    Follow-up commands:
      - `mosoo console apps app-overview --app-id <app-id> -o json`

### `mosoo console apps delete-app-deployment`

- Summary: Delete an app deployment
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console apps deploy-app`

- Summary: Deploy app
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-config-path` (variable): input.configPath
  - `--input-repo-url` (variable, required): input.repoUrl
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console apps rename-app`

- Summary: Rename an app
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-name` (variable, required): input.name
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## Channels

### `mosoo console channels agent-channel-binding-list`

- Summary: List agent channel bindings
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--agent-id` (variable, required): agentId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels create-discord-agent-channel-binding`

- Summary: Create a discord agent channel binding
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-application-id` (variable, required): input.applicationId
  - `--input-bot-token` (variable, required): input.botToken
  - `--input-app-id` (variable, required): input.appId
  - `--input-relay-secret` (variable, required): input.relaySecret
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels create-lark-agent-channel-binding`

- Summary: Create a lark agent channel binding
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-lark-app-id` (variable, required): input.larkAppId
  - `--input-app-secret` (variable, required): input.appSecret
  - `--input-connection-mode` (variable, required, one of: webhook|websocket): input.connectionMode
  - `--input-domain` (variable, required, one of: feishu|lark): input.domain
  - `--input-encrypt-key` (variable): input.encryptKey
  - `--input-app-id` (variable, required): input.appId
  - `--input-verification-token` (variable): input.verificationToken
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels create-slack-agent-channel-binding`

- Summary: Create a slack agent channel binding
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-app-level-token` (variable): input.appLevelToken
  - `--input-bot-token` (variable, required): input.botToken
  - `--input-app-id` (variable, required): input.appId
  - `--input-signing-secret` (variable, required): input.signingSecret
  - `--input-thread-replies-require-mention` (variable): input.threadRepliesRequireMention
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels create-telegram-agent-channel-binding`

- Summary: Create a telegram agent channel binding
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-bot-token` (variable, required): input.botToken
  - `--input-app-id` (variable, required): input.appId
  - `--input-webhook-secret` (variable, required): input.webhookSecret
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels delete-agent-channel-binding`

- Summary: Delete an agent channel binding
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-binding-id` (variable, required): input.bindingId
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels poll-lark-agent-channel-registration`

- Summary: Poll lark agent channel registration
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-device-code` (variable, required): input.deviceCode
  - `--input-domain` (variable, required, one of: feishu|lark): input.domain
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels poll-we-chat-agent-channel-pairing`

- Summary: Poll we chat agent channel pairing
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-app-id` (variable, required): input.appId
  - `--input-qr-token` (variable, required): input.qrToken
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels start-lark-agent-channel-registration`

- Summary: Start lark agent channel registration
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-domain` (variable, required, one of: feishu|lark): input.domain
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console channels start-we-chat-agent-channel-pairing`

- Summary: Start we chat agent channel pairing
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## Common

### `mosoo console common app-info`

- Summary: App info
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags: none
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## Cost

### `mosoo console cost agent-cost-card`

- Summary: Agent cost card
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--agent-id` (variable, required): agentId
  - `--range` (variable, required, one of: LAST_7_DAYS|LAST_30_DAYS|MONTH_TO_DATE|LAST_90_DAYS): range
  - `--run-purposes` (variable): runPurposes
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console cost app-cost-card`

- Summary: App cost card
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--range` (variable, required, one of: LAST_7_DAYS|LAST_30_DAYS|MONTH_TO_DATE|LAST_90_DAYS): range
  - `--run-purposes` (variable): runPurposes
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console cost organization-billing-cost-card`

- Summary: Organization billing cost card
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--organization-id` (variable, required): organizationId
  - `--range` (variable, required, one of: LAST_7_DAYS|LAST_30_DAYS|MONTH_TO_DATE|LAST_90_DAYS): range
  - `--run-purposes` (variable): runPurposes
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## Credentials

### `mosoo console credentials available-agent-models`

- Summary: Available agent models
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--runtime-id` (variable, required): runtimeId
  - `--current-model-id` (variable): currentModelId
  - `--current-vendor-id` (variable): currentVendorId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console credentials create-vendor-credential`

- Summary: Add a provider key
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Shortcuts:
  - `mosoo add-key`
- Flags:
  - `--input-api-base` (variable): input.apiBase
  - `--input-api-key` (variable, required): input.apiKey
  - `--input-models` (variable): input.models
  - `--input-name` (variable, required): input.name
  - `--input-app-id` (variable, required): input.appId
  - `--input-vendor-id` (variable, required): input.vendorId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
  - For preset providers such as openai or anthropic, omit input.models unless Mosoo asks for explicit model configuration.
  - Use --input-api-key-env, --input-api-key-file, or --input-api-key-stdin so provider keys do not appear in shell history.
  - Current Lathe required variable flags must be present; safe input modes satisfy the required input.apiKey variable.
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Examples:
  - Create an OpenAI provider credential from an environment variable
    Command: `mosoo console credentials create-vendor-credential \ --input-app-id <app-id> \ --input-vendor-id openai \ --input-name "OpenAI" \ --input-api-key-env OPENAI_API_KEY \ -o json`
    Body shape: `{"input":{"apiKey":"$OPENAI_API_KEY","appId":"\u003capp-id\u003e","name":"OpenAI","vendorId":"openai"}}`
    Output ID path: `data.createVendorCredential.id`
    Follow-up commands:
      - `mosoo console credentials vendor-credential-list --app-id <app-id> -o json`

### `mosoo console credentials delete-vendor-credential`

- Summary: Delete a vendor credential
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-id` (variable, required): input.id
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console credentials set-default-vendor-credential`

- Summary: Set default vendor credential
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-id` (variable, required): input.id
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console credentials test-vendor-credential`

- Summary: Test a provider key
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-api-base` (variable): input.apiBase
  - `--input-api-key` (variable, required): input.apiKey
  - `--input-model-id` (variable): input.modelId
  - `--input-app-id` (variable, required): input.appId
  - `--input-vendor-id` (variable, required): input.vendorId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
  - Prefer --input-api-key-env, --input-api-key-file, or --input-api-key-stdin so provider secrets do not appear in shell history.
  - Current Lathe required variable flags must be present; --set and --set-str can supplement body fields but do not replace required flags.
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Examples:
  - Smoke test an OpenAI provider key from the environment before saving it.
    Command: `mosoo console credentials test-vendor-credential \ --input-app-id <app-id> \ --input-vendor-id openai \ --input-model-id gpt-4o-mini \ --input-api-key-env OPENAI_API_KEY \ -o json`
    Body shape: `{"input":{"apiKey":"\u003cread from OPENAI_API_KEY\u003e","appId":"\u003capp-id\u003e","modelId":"gpt-4o-mini","vendorId":"openai"}}`
    Follow-up commands:
      - `mosoo console credentials create-vendor-credential --input-app-id <app-id> --input-vendor-id openai --input-name "OpenAI" --input-api-key-env OPENAI_API_KEY -o json`

### `mosoo console credentials update-vendor-credential`

- Summary: Update a vendor credential
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-api-base` (variable): input.apiBase
  - `--input-api-key` (variable): input.apiKey
  - `--input-id` (variable, required): input.id
  - `--input-models` (variable): input.models
  - `--input-name` (variable): input.name
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console credentials vendor-credential-list`

- Summary: List provider keys
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console credentials list --app-id <app-id> -o json`

## Environments

### `mosoo console environments app-environment-list`

- Summary: List app environments
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console environments create-environment-fork`

- Summary: Create an environment fork
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-environment-id` (variable, required): input.environmentId
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console environments delete-environment`

- Summary: Delete an environment
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-environment-id` (variable, required): input.environmentId
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console environments environment`

- Summary: Environment
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--environment-id` (variable, required): environmentId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console environments set-app-default-environment`

- Summary: Set app default environment
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-environment-id` (variable, required): input.environmentId
  - `--input-app-id` (variable, required): input.appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console environments set-environment-variable-value`

- Summary: Set environment variable value
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-environment-id` (variable, required): input.environmentId
  - `--input-key` (variable, required): input.key
  - `--input-app-id` (variable, required): input.appId
  - `--input-value` (variable, required): input.value
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## Files

### `mosoo console files file-list`

- Summary: List files
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-scope-id` (variable): input.scopeId
  - `--input-scope-kind` (variable, one of: account|agent_package|app_draft|library|session): input.scopeKind
  - `--input-session-id` (variable): input.sessionId
  - `--input-session-kind` (variable, one of: artifact|attachment): input.sessionKind
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## MCP

### `mosoo console mcp connect-mcp-bearer`

- Summary: Connect a mcp bearer
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-server-id` (variable, required): input.serverId
  - `--input-subject-label` (variable): input.subjectLabel
  - `--input-token` (variable, required): input.token
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console mcp create-app-mcp-server`

- Summary: Create an app mcp server
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-auth-type` (variable, required, one of: oauth|bearer): input.authType
  - `--input-description` (variable): input.description
  - `--input-icon-url` (variable): input.iconUrl
  - `--input-name` (variable, required): input.name
  - `--input-oauth-client-id` (variable): input.oauthClientId
  - `--input-oauth-client-secret` (variable): input.oauthClientSecret
  - `--input-app-id` (variable, required): input.appId
  - `--input-url` (variable, required): input.url
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console mcp delete-mcp-server`

- Summary: Delete a mcp server
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--server-id` (variable, required): serverId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console mcp mcp-o-auth-flow-status`

- Summary: console_mcpOAuthFlowStatus
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--flow-id` (variable, required): flowId

### `mosoo console mcp mcp-registry`

- Summary: Mcp registry
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console mcp revoke-mcp-credential`

- Summary: Revoke a mcp credential
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--server-id` (variable, required): serverId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console mcp set-mcp-server-enabled`

- Summary: Set mcp server enabled
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--server-id` (variable, required): serverId
  - `--enabled` (variable, required): enabled
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console mcp start-mcp-o-auth`

- Summary: console_startMcpOAuth
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-return-url` (variable): input.returnUrl
  - `--input-server-id` (variable, required): input.serverId

## Onboarding

### `mosoo console onboarding onboarding-bootstrap`

- Summary: Onboarding bootstrap
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-name` (variable): input.name
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## Sessions

### `mosoo console sessions add-session-resource`

- Summary: Add a session resource
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-file-content-type` (variable, required): input.file.contentType
  - `--input-file-name` (variable, required): input.file.name
  - `--input-file-size` (variable, required): input.file.size
  - `--input-app-id` (variable, required): input.appId
  - `--input-session-id` (variable, required): input.sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions agent-session-diagnostics`

- Summary: Agent session diagnostics
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions agent-session-list`

- Summary: List agent sessions
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--agent-id` (variable, required): agentId
  - `--archived` (variable): archived
  - `--before-cursor` (variable): beforeCursor
  - `--limit` (variable): limit
  - `--participant-only` (variable): participantOnly
  - `--type` (variable, one of: api_channel|preview|ui): type
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console sessions agent-session-list`

### `mosoo console sessions agent-session-retrieve`

- Summary: Agent session retrieve
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions archive-agent-session`

- Summary: Archive an agent session
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions auto-title-session`

- Summary: Auto a title session
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-session-id` (variable, required): input.sessionId
  - `--input-title` (variable, required): input.title
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions create-agent-session`

- Summary: Create an agent session
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-agent-id` (variable, required): input.agentId
  - `--input-app-id` (variable, required): input.appId
  - `--input-type` (variable, one of: api_channel|preview|ui): input.type
  - `--input-wait-for-runtime-ready` (variable): input.waitForRuntimeReady
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions delete-agent-session`

- Summary: Delete an agent session
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions list-session-resources`

- Summary: List session resources
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions prewarm-agent-session`

- Summary: Prewarm an agent session
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions remove-session-resource`

- Summary: Remove a session resource
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-resource-id` (variable, required): input.resourceId
  - `--input-session-id` (variable, required): input.sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions rename-session`

- Summary: Rename a session
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-session-id` (variable, required): input.sessionId
  - `--input-title` (variable, required): input.title
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions session`

- Summary: Session
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions session-list`

- Summary: List sessions
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--archived` (variable): archived
  - `--before-cursor` (variable): beforeCursor
  - `--limit` (variable): limit
  - `--app-id` (variable, required): appId
  - `--type` (variable, one of: api_channel|preview|ui): type
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console sessions session-list`

### `mosoo console sessions session-messages`

- Summary: Session messages
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions session-process-events`

- Summary: Session process events
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--limit` (variable): limit
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions start-agent-run`

- Summary: Start an Agent run
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Shortcuts:
  - `mosoo run`
- Flags:
  - `--input-agent-id` (variable): input.agentId
  - `--input-app-id` (variable, required): input.appId
  - `--input-client-request-id` (variable): input.clientRequestId
  - `--input-prompt` (variable, required): input.prompt
  - `--input-session-id` (variable): input.sessionId
  - `--input-type` (variable, one of: api_channel|preview|ui): input.type
  - `--input-wait-for-runtime-ready` (variable): input.waitForRuntimeReady
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
  - This is the generated main path for mosoo run. Use the returned appId/sessionId with thread-session-process-events to poll output.
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example:

```
mosoo console sessions run \
  --input-app-id <app-id> \
  --input-agent-id <agent-id> \
  --input-prompt "Summarize this repository" \
  -o json
```

### `mosoo console sessions thread-agent-session-list`

- Summary: List thread agent sessions
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--archived` (variable): archived
  - `--before-cursor` (variable): beforeCursor
  - `--limit` (variable): limit
  - `--app-id` (variable, required): appId
  - `--type` (variable, one of: api_channel|preview|ui): type
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions thread-agent-session-retrieve`

- Summary: Thread agent session retrieve
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions thread-session-messages`

- Summary: Thread session messages
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console sessions thread-session-process-events`

- Summary: Poll Thread session events
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--limit` (variable): limit
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console sessions events --app-id <app-id> --session-id <session-id> -o json`

### `mosoo console sessions unarchive-agent-session`

- Summary: Unarchive an agent session
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--session-id` (variable, required): sessionId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## Skills

### `mosoo console skills app-skill-list`

- Summary: List app skills
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console skills create-skill-fork`

- Summary: Create a skill fork
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-app-id` (variable, required): input.appId
  - `--input-skill-id` (variable, required): input.skillId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console skills delete-owned-skill`

- Summary: Delete an owned skill
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--skill-id` (variable, required): skillId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console skills skill-detail`

- Summary: Skill detail
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--app-id` (variable, required): appId
  - `--skill-id` (variable, required): skillId
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

## User

### `mosoo console user set-system-agent-model`

- Summary: Set system agent model
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-model-id` (variable, required): input.modelId
  - `--input-vendor` (variable, required): input.vendor
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console user update-profile`

- Summary: Update a profile
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags:
  - `--input-image-url` (variable): input.imageUrl
  - `--input-name` (variable, required): input.name
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.

### `mosoo console user viewer`

- Summary: Show the signed-in viewer
- HTTP: `POST /graphql`
- Auth: required
- Body: required; templated body, set inputs under `variables` with --set/--set-str/--file
- Flags: none
- Notes:
  - Uses POST /graphql on the console default hostname (/api).
- Known errors:
  - HTTP 401: Missing, invalid, or revoked personal access token.
- Example: `mosoo console user viewer`
