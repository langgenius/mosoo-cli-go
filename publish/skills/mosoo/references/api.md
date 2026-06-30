> ## Documentation Index
> Fetch the complete documentation index at: https://docs.mosoo.ai/llms.txt
> Use this file to discover all available pages before exploring further.

# Mosoo API for coding agents

> Machine-oriented guide for calling published Mosoo Agents through the public API.

# Mosoo API for coding agents

Use this document when integrating an existing, published Mosoo Agent into application code through the public API.

## Integration scope

This document only covers API calls between application code and an existing Mosoo Agent: sending inputs, attaching files, reading outputs, and handling public API errors.

It does not cover:

* Creating, configuring, or publishing Agents.
* Managing Mosoo Apps.
* Managing Agent lifecycle, versions, readiness, runtime settings, model providers, tools, or channels.
* Any other Mosoo product surface outside the public API used to pass input to an Agent and read output from it.

This document is also not the source for secrets or live resource IDs. Do not invent API tokens, `agentId`, `threadId`, `fileId`, or `runId` values. Use values supplied by the user, environment variables, Mosoo UI, or Mosoo CLI when available.

This document is not a replacement for the raw OpenAPI document when generating clients or validating every schema detail. Use the raw OpenAPI for code generation and strict schema validation.

## Integration model

Mosoo is the Agent runtime. The published Agent runs inside a Mosoo-managed sandbox, and your application backend and product code should call that sandboxed Agent through the public API.

Build the app-side integration layer around Mosoo Threads, events, files, and public error handling. Your application can still own backend APIs, jobs, business logic, user flows, and data storage around the sandboxed Agent. For a Mosoo Agent integration, do not implement a parallel Agent layer, sandbox, model loop, planner, tool runner, memory system, lifecycle manager, or provider integration in your application.

Primary contract:

* Human docs: `/`, `/quickstart`, `/auth-and-access`
* Raw OpenAPI: `/mosoo-openapi.en.generated.json`
* API version: `v1`
* Base URL used in examples: `https://try.mosoo.ai/api/v1`

Do not infer request fields that are not listed here or in the raw OpenAPI. The public API rejects unsupported fields.

## What Mosoo exposes

Mosoo exposes published Agents through Thread-based HTTP APIs. If required credentials or resource IDs are missing, get them from the environment, the user, Mosoo UI, or Mosoo CLI when it is available. Do not guess or synthesize Mosoo IDs.

| Concept   | Meaning for callers                                                                                                                                   |
| --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| Agent     | A configured Mosoo Agent. It must be published and have API access enabled before the API can call it.                                                |
| `agentId` | The published Agent ID shown in the Agent API Access panel                                                                                            |
| API token | Bearer credential used by your code. In the current single-user model, it authenticates the Mosoo user who owns the local Agents, Threads, and files. |
| Thread    | The API conversation container. It records messages, files, run status, and public event history.                                                     |
| Run       | One execution pass of the Agent on a Thread. A user message can create or resume a Run.                                                               |
| Event     | A public timeline entry for Thread inputs, Agent output deltas, tool updates, file changes, status changes, and usage updates.                        |
| File      | A Thread attachment uploaded by the caller or an Agent artifact produced during execution.                                                            |

Resource IDs in v1 are bare ULIDs, not prefixed IDs.

## Authentication and access

All API requests require an API token:

```http theme={null}
Authorization: Bearer mst_...
```

JSON requests also require:

```http theme={null}
Content-Type: application/json
```

A valid request must pass all of these checks:

1. The Agent exists.
2. The Agent is published and has API access enabled.
3. The API token is valid and not revoked.
4. The requested Thread or file exists in the same single-user Mosoo workspace.

Mosoo currently exposes this API as a single-user integration surface. The API token authenticates requests for that user; it does not switch users, create a multi-user access context, or let a request override model provider, channel, tool, or runtime settings. Configure and publish the Agent in Mosoo before calling the API.

## Responsibility boundary

| Your application owns                                                                                                                                                                                    | Mosoo owns                                                                                                                                                          |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| App UI, app routing, app backend APIs, jobs, app-side users, business logic, data storage, `thread.id` persistence, and caller-owned correlation IDs such as `client_external_ref` or `clientRequestId`. | Sandboxed published Agent execution, model/provider configuration, tool execution, Agent runtime behavior, Agent memory/runtime state, and public event generation. |

## Quick start workflow

Use this workflow when implementing the quickstart in application code. It is based on the human quickstart, but it starts after the Agent already exists, is published, and has API access enabled.

| Step                      | Call                                                                                                                       | Persist or read                                                                                | Stop condition                                                                   |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| 1. Load credentials       | No API call. Read `MOSOO_API_TOKEN` and `MOSOO_AGENT_ID` from the user, environment, Mosoo UI, or Mosoo CLI.               | Keep the token in secret storage. Keep `agentId` as configuration.                             | Missing token or `agentId`; do not invent either value.                          |
| 2. Create a Thread        | `POST /agents/{agentId}/threads` with the first user input.                                                                | Persist `thread.id`. Optionally persist `client_external_ref` for your app record.             | A Thread exists and the first Run has been queued.                               |
| 3. Continue the Thread    | `POST /threads/{threadId}/events` with `user_message`, `permission_decision`, or `user_interrupt` events.                  | Persist caller-owned references such as `clientRequestId` if your app needs correlation.       | The follow-up input has been accepted by the API.                                |
| 4. Read output            | `GET /threads/{threadId}/events` for the event log, or `GET /threads/{threadId}/events/stream` for streaming.              | Read public Agent output, run status, usage updates, tool status, and file events from events. | The app has enough public events to render the Agent response or current status. |
| 5. Attach files if needed | For small base64 payloads, `POST /threads/{threadId}/files`. For binary or larger uploads, use the file upload flow below. | Persist returned `file.id` when the app needs to reference or remove the file later.           | The file is attached before sending the message that depends on it.              |

Implementation rules:

* Build the app-side integration layer around the published Mosoo Agent running in Mosoo's sandbox; do not implement a local Agent runtime or replacement sandbox.
* Do not call model providers directly for this Mosoo Agent integration.
* Do not send provider credentials, model configuration, channel credentials, tool configuration, or Agent configuration through this public API.
* Use `Idempotency-Key` for Thread creation and event submission.
* Treat Thread events as the integration contract and `GET /threads/{threadId}/events` as the stable source for results.
* Keep Agent creation, App management, and Agent lifecycle operations outside this workflow.
* On failure, branch on `error.code` and follow the error handling table below.

## Minimal call sequence

Set credentials:

```bash theme={null}
export MOSOO_API_TOKEN="mst_..."
export MOSOO_AGENT_ID="01J00000000000000000000001"
```

Create a Thread and queue the initial Run:

```bash theme={null}
curl -X POST "https://try.mosoo.ai/api/v1/agents/$MOSOO_AGENT_ID/threads" \
  -H "Authorization: Bearer $MOSOO_API_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: ticket-182-create-thread" \
  -d '{
    "client_external_ref": "ticket-182",
    "input": {
      "type": "user.message",
      "content": [
        {
          "type": "text",
          "text": "Triage this customer escalation and suggest next steps."
        }
      ]
    }
  }'
```

Store `thread.id`:

```bash theme={null}
export MOSOO_THREAD_ID="01J00000000000000000000002"
```

Send a follow-up message:

```bash theme={null}
curl -X POST "https://try.mosoo.ai/api/v1/threads/$MOSOO_THREAD_ID/events" \
  -H "Authorization: Bearer $MOSOO_API_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: ticket-182-follow-up-1" \
  -d '{
    "events": [
      {
        "type": "user_message",
        "clientRequestId": "ticket-182-message-1",
        "text": "Give me the three highest-priority next actions."
      }
    ]
  }'
```

Read the public event log:

```bash theme={null}
curl "https://try.mosoo.ai/api/v1/threads/$MOSOO_THREAD_ID/events?limit=100" \
  -H "Authorization: Bearer $MOSOO_API_TOKEN"
```

## File upload flow

Use this flow when a Thread needs a file attachment. Public uploads are currently single PUT uploads up to 67108864 bytes.

Create a Thread-scoped upload session:

```bash theme={null}
curl -X POST "https://try.mosoo.ai/api/v1/threads/$MOSOO_THREAD_ID/files/uploads" \
  -H "Authorization: Bearer $MOSOO_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file": {
      "name": "brief.txt",
      "contentType": "text/plain",
      "size": 19
    }
  }'
```

The response is a `FileUploadSummary`. Store `fileId`.

```bash theme={null}
export MOSOO_FILE_ID="01J0000000000000000000000J"
```

Upload bytes:

```bash theme={null}
curl -X PUT "https://try.mosoo.ai/api/v1/files/$MOSOO_FILE_ID/content" \
  -H "Authorization: Bearer $MOSOO_API_TOKEN" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @brief.txt
```

Complete the upload:

```bash theme={null}
curl -X POST "https://try.mosoo.ai/api/v1/files/$MOSOO_FILE_ID/complete" \
  -H "Authorization: Bearer $MOSOO_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

Claim the ready draft file into the Thread:

```bash theme={null}
curl -X POST "https://try.mosoo.ai/api/v1/threads/$MOSOO_THREAD_ID/files" \
  -H "Authorization: Bearer $MOSOO_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "fileId": "01J0000000000000000000000J"
  }'
```

Use the file in a later user message with `attachmentIds`:

```bash theme={null}
curl -X POST "https://try.mosoo.ai/api/v1/threads/$MOSOO_THREAD_ID/events" \
  -H "Authorization: Bearer $MOSOO_API_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: ticket-182-file-question-1" \
  -d '{
    "events": [
      {
        "type": "user_message",
        "attachmentIds": ["01J0000000000000000000000J"],
        "text": "Summarize the attached file."
      }
    ]
  }'
```

When creating a Thread, `files[].file_id` can reference ready draft file handles uploaded by the same caller. For Thread-scoped uploads after Thread creation, use the upload flow above and then call `POST /threads/{threadId}/files` with `fileId`.

## Idempotency

`Idempotency-Key` is supported on:

* `POST /agents/{agentId}/threads`
* `POST /threads/{threadId}/events`

Rules:

* The key is scoped to the API token, method, route, and request body.
* Reusing the same key with the same request replays the stored response.
* Reusing the same key with a different request returns `409 idempotency_conflict`.
* Reusing the key while the first request is still processing returns `409 idempotency_conflict`.
* Keys must be non-empty and 128 characters or fewer.
* Conflict responses can include `Retry-After`.

Generate stable keys from your own operation IDs, such as `ticket-182-create-thread` or `job-2026-06-23-run-1-message-3`.

## Reading results

The stable read surface is the Thread event log.

Use `GET /threads/{threadId}/events` for polling and snapshots. Use `GET /threads/{threadId}/events/stream` for long-running consumer UX.

Do not expect event APIs to expose raw runtime payloads, private transcripts, or internal diagnostics. Each event entry has public fields:

| Field        | Meaning                                                                                                 |
| ------------ | ------------------------------------------------------------------------------------------------------- |
| `id`         | Event ID, bare ULID, monotonically increasing in chronological order.                                   |
| `type`       | Public event type such as `run.started`, `agent.message.delta`, `tool.use.started`, or `usage.updated`. |
| `status`     | `available`, `error`, or `unsupported`.                                                                 |
| `content`    | Public event content or a reference to the associated payload.                                          |
| `occurredAt` | RFC 3339 timestamp.                                                                                     |
| `durationMs` | Duration in milliseconds when applicable, otherwise `null`.                                             |
| `tokens`     | Token count when applicable, otherwise `null`.                                                          |

Important event types:

* `user.message`
* `agent.message.delta`
* `agent.thinking.delta`
* `tool.confirmation.required`
* `tool.use.started`
* `tool.use.completed`
* `file.changed`
* `session_files.updated`
* `run.started`
* `run.completed`
* `run.failed`
* `session.status`
* `usage.updated`

## Status values

Thread status:

| Status         | Meaning                                       |
| -------------- | --------------------------------------------- |
| `IDLE`         | No active Run. A user message can queue work. |
| `RUNNING`      | A Run is executing.                           |
| `RESCHEDULING` | The Thread is between runs.                   |
| `TERMINATED`   | The Thread has ended.                         |

Run status:

| Status          | Meaning                                                         |
| --------------- | --------------------------------------------------------------- |
| `queued`        | Run exists but has not started.                                 |
| `booting`       | Runtime is preparing.                                           |
| `running`       | Run is executing.                                               |
| `waiting_input` | Run is blocked on caller/user input, often permission decision. |
| `completed`     | Terminal success.                                               |
| `failed`        | Terminal failure.                                               |
| `cancelled`     | Terminal cancellation.                                          |
| `expired`       | Terminal timeout or expiry.                                     |

## Error handling

All non-2xx public API JSON errors use this envelope:

```json theme={null}
{
  "error": {
    "code": "invalid_request",
    "message": "Request body must be an object."
  }
}
```

Branch on `error.code`, not `error.message`. The message is for developers and should not be displayed directly to end users.

| HTTP | `error.code`           | Meaning                                                                                 | Caller action                                                                                |
| ---- | ---------------------- | --------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| 400  | `invalid_request`      | Request shape, field value, body size, limit, or unsupported field is invalid.          | Fix the request. Do not retry unchanged.                                                     |
| 400  | `invalid_json`         | Body is not valid JSON.                                                                 | Fix serialization or `Content-Type`. Do not retry unchanged.                                 |
| 401  | `unauthenticated`      | Missing, invalid, or revoked API token.                                                 | Re-read `Authorization`; rotate or recreate the API token.                                   |
| 403  | `forbidden`            | API token is valid but the operation is not allowed for this Agent, Thread, or file.    | Check that the request uses the expected token and a resource from the same Mosoo workspace. |
| 404  | `not_found`            | Agent, Thread, or file is not visible in the current Mosoo workspace or does not exist. | Check the ID and use resource IDs returned by the API.                                       |
| 409  | `agent_not_published`  | Agent exists but is not published as an active API service.                             | Ask the user to publish the Agent and enable API access.                                     |
| 409  | `service_inactive`     | Published API service has no live published version.                                    | Ask the user to republish or repair the Agent in Mosoo.                                      |
| 409  | `readiness_blocked`    | Agent is not ready to run.                                                              | Do not blindly retry; ask the user to fix Agent readiness or configuration in Mosoo.         |
| 409  | `idempotency_conflict` | Same `Idempotency-Key` is processing or was used for a different request.               | If still processing, wait for `Retry-After`; if body differs, use a new key.                 |
| 429  | `rate_limited`         | API token exceeded public API rate limits.                                              | Back off and retry after `Retry-After`.                                                      |
| 500  | `internal_error`       | Mosoo failed internally.                                                                | Retry briefly with backoff; persist failure details if it repeats.                           |

Retry policy:

* Safe to retry with the same `Idempotency-Key`: network timeouts, 5xx from create-thread or send-events, and `idempotency_conflict` caused by in-flight processing.
* Safe to retry after delay: `rate_limited`, using `Retry-After`.
* Do not retry unchanged: `invalid_request`, `invalid_json`, `unauthenticated`, `forbidden`, `not_found`, `agent_not_published`, `service_inactive`, `readiness_blocked`.

## API contract

This section is generated from `/mosoo-openapi.en.generated.json`. Do not edit it manually; run `npm run openapi:sync`.

All endpoints below are relative to `/api/v1`.

| Method   | Path                                 | Purpose                                                             |
| -------- | ------------------------------------ | ------------------------------------------------------------------- |
| `POST`   | `/files/{fileId}/complete`           | Complete Thread file upload                                         |
| `GET`    | `/files/{fileId}/content`            | Download Thread file content                                        |
| `PUT`    | `/files/{fileId}/content`            | Upload Thread file content                                          |
| `GET`    | `/agents/{agentId}/threads`          | List Threads for an Agent API Endpoint                              |
| `POST`   | `/agents/{agentId}/threads`          | Create a Thread for an Agent API Endpoint                           |
| `GET`    | `/threads/{threadId}`                | Retrieve Thread summary                                             |
| `DELETE` | `/threads/{threadId}`                | Delete a Thread                                                     |
| `POST`   | `/threads/{threadId}/archive`        | Archive a Thread                                                    |
| `GET`    | `/threads/{threadId}/events`         | List Thread events                                                  |
| `POST`   | `/threads/{threadId}/events`         | Send user messages, permission decisions, or interrupts to a Thread |
| `GET`    | `/threads/{threadId}/events/stream`  | Stream Thread events                                                |
| `GET`    | `/threads/{threadId}/files`          | List Thread files                                                   |
| `POST`   | `/threads/{threadId}/files`          | Add a Thread file                                                   |
| `POST`   | `/threads/{threadId}/files/uploads`  | Create a Thread file upload                                         |
| `DELETE` | `/threads/{threadId}/files/{fileId}` | Remove a Thread file                                                |
| `POST`   | `/threads/{threadId}/unarchive`      | Unarchive a Thread                                                  |

Common error response envelope:

```json theme={null}
{
  "error": {
    "code": "invalid_request",
    "message": "Request body must be an object."
  }
}
```

### `POST /files/{fileId}/complete`

Purpose: Completes a pending single PUT Thread file upload. The file must have been created through POST /threads/{threadId}/files/uploads, uploaded through PUT /files/{fileId}/content, and belong to a public Thread visible to the API token.

Path params:

* `fileId` required, `string(ulid)`. File ID returned by add or list Thread files. v1 IDs are bare ULIDs.

Request body:

* `application/json`: `CompleteFileUploadRequest`

Example:

```json theme={null}
{}
```

Success responses:

* `200`: Completed files API upload. (`application/json` -> `CompleteFileUploadResponse`)

Example `200`:

```json theme={null}
{
  "file": {
    "createdAt": "2026-05-19T00:02:00.000Z",
    "createdBy": "01J00000000000000000000002",
    "etag": "etag-1",
    "expiresAt": null,
    "id": "01J0000000000000000000000J",
    "mimeType": "text/plain",
    "name": "brief.txt",
    "path": "session-files/01J0000000000000000000000J/brief.txt",
    "sessionKind": "attachment",
    "size": 19,
    "status": "ready",
    "updatedAt": "2026-05-19T00:03:00.000Z",
    "version": 1
  }
}
```

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `GET /files/{fileId}/content`

Purpose: Downloads bytes for a ready Thread attachment or Agent artifact. The file must belong to a public Thread visible to the API token.

Path params:

* `fileId` required, `string(ulid)`. File ID returned by add or list Thread files. v1 IDs are bare ULIDs.

Query params:

* `disposition` optional, `"attachment" | "inline"`. Controls the Content-Disposition response header. Use attachment for downloads or inline for previewable content.

Success responses:

* `200`: Thread file content. (`application/octet-stream` -> `string(binary)`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `PUT /files/{fileId}/content`

Purpose: Uploads raw bytes for a pending single PUT Thread file upload. The file must have been created through POST /threads/{threadId}/files/uploads and belong to a public Thread visible to the API token.

Path params:

* `fileId` required, `string(ulid)`. File ID returned by add or list Thread files. v1 IDs are bare ULIDs.

Request body:

* `application/octet-stream`: `string(binary)`

Success responses:

* `200`: Uploaded. (`application/json` -> `object`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `GET /agents/{agentId}/threads`

Purpose: Returns Threads created by the authenticated API token.

Path params:

* `agentId` required, `string(ulid)`. Agent API Endpoint ID from the Agent's API Access panel. v1 IDs are bare ULIDs.

Query params:

* `archived` optional, `boolean`. Filter by archived state: true returns only archived Threads, false only active ones. Omit to return all Threads.

Success responses:

* `200`: Thread list. (`application/json` -> `object`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `POST /agents/{agentId}/threads`

Purpose: Creates a Thread and the backing AgentSession. If input is present, Mosoo also queues the initial Run. If input is omitted, the Thread is immediately visible with IDLE status and no run. API token requests are attributed to the current Mosoo user.

Path params:

* `agentId` required, `string(ulid)`. Agent API Endpoint ID from the Agent's API Access panel. v1 IDs are bare ULIDs.

Headers:

* `Idempotency-Key` optional, `string`. Optional key for retry-safe create-thread and send-events calls. Reusing the same key with the same request returns the original response. Reusing the key while the original request is still processing returns 409.

Request body:

* `application/json`: `CreateThreadRequest`

Example `emptyThread`:

```json theme={null}
{
  "client_external_ref": "draft-empty-thread"
}
```

Example `accessTokenWithFile`:

```json theme={null}
{
  "client_external_ref": "linear-ENG-123",
  "files": [
    {
      "file_id": "01J0000000000000000000000J"
    }
  ],
  "input": {
    "content": [
      {
        "text": "Summarize the attached launch plan and list follow-ups.",
        "type": "text"
      }
    ],
    "type": "user.message"
  }
}
```

Example `accessTokenBasic`:

```json theme={null}
{
  "client_external_ref": "demo-thread-001",
  "input": {
    "content": [
      {
        "text": "Say hello from the API.",
        "type": "text"
      }
    ],
    "type": "user.message"
  }
}
```

Example `cattleAgentSameShape`:

```json theme={null}
{
  "input": {
    "content": [
      {
        "text": "Run this one-off Public Thread API request.",
        "type": "text"
      }
    ],
    "type": "user.message"
  }
}
```

Success responses:

* `201`: Created Thread. (`application/json` -> `CreateThreadResponse`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `GET /threads/{threadId}`

Purpose: Returns the current Thread summary, its most recent Run, and links.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Success responses:

* `200`: Thread summary. (`application/json` -> `RetrieveThreadResponse`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `DELETE /threads/{threadId}`

Purpose: Permanently deletes the Thread and its backing AgentSession.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Success responses:

* `200`: Deleted. (`application/json` -> `object`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `POST /threads/{threadId}/archive`

Purpose: Archives the Thread so it is hidden from default Thread lists.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Success responses:

* `200`: Archived. (`application/json` -> `object`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `GET /threads/{threadId}/events`

Purpose: Returns the latest public event log entries for this Thread in chronological order. This is the stable snapshot read surface for CLI and API consumers; it does not expose raw runtime payloads, transcript, or diagnostics.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Query params:

* `limit` optional, `integer`. Maximum number of latest Thread events to return.

Success responses:

* `200`: Thread event list. (`application/json` -> `ThreadEventListResponse`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `POST /threads/{threadId}/events`

Purpose: Applies a batch of events to the Thread: send user messages, answer pending permission requests, or interrupt the current Run. A user message queues a new Run when the Thread is idle.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Headers:

* `Idempotency-Key` optional, `string`. Optional key for retry-safe create-thread and send-events calls. Reusing the same key with the same request returns the original response. Reusing the key while the original request is still processing returns 409.

Request body:

* `application/json`: `SendEventsRequest`

Example:

```json theme={null}
{
  "events": [
    {
      "text": "Say hello from the API.",
      "type": "user_message"
    }
  ]
}
```

Success responses:

* `200`: Accepted event batch. (`application/json` -> `SendEventsResponse`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `GET /threads/{threadId}/events/stream`

Purpose: Streams public Thread event log entries as Server-Sent Events. Each `thread.event` data payload uses the same ThreadEventLogEntry shape as GET /threads/{threadId}/events. The stream is for long-running consumer UX and does not expose raw runtime payloads, internal diagnostics, or private transcripts.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Query params:

* `limit` optional, `integer`. Maximum number of latest Thread events to return.

Success responses:

* `200`: Thread event stream. (`text/event-stream` -> `string`)

Example `200`:

```json theme={null}
": connected\n\nevent: thread.event\nid: 01J00000000000000000000010\ndata: {\"id\":\"01J00000000000000000000010\",\"type\":\"run.started\",\"status\":\"completed\",\"content\":\"01J0000000000000000000000A\",\"occurredAt\":\"2026-05-19T00:00:01.000Z\",\"durationMs\":null,\"tokens\":null}\n\n"
```

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `GET /threads/{threadId}/files`

Purpose: Lists files attached to the Thread, including caller attachments and Agent artifacts.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Success responses:

* `200`: Thread file list. (`application/json` -> `ThreadFileListResponse`)

Example `200`:

```json theme={null}
{
  "files": [
    {
      "committed": true,
      "createdAt": "2026-05-19T00:02:00.000Z",
      "id": "01J0000000000000000000000J",
      "kind": "attachment",
      "mimeType": "text/plain",
      "name": "brief.txt",
      "size": 19
    }
  ]
}
```

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `POST /threads/{threadId}/files`

Purpose: Claims a ready draft file handle into this Thread. Upload file bytes through the files data plane first, then pass the resulting fileId here.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Request body:

* `application/json`: `CreateThreadFileRequest`

Example:

```json theme={null}
{
  "fileId": "01J0000000000000000000000J"
}
```

Success responses:

* `201`: Created Thread file. (`application/json` -> `ThreadFileResponse`)

Example `201`:

```json theme={null}
{
  "file": {
    "committed": true,
    "createdAt": "2026-05-19T00:02:00.000Z",
    "id": "01J0000000000000000000000J",
    "kind": "attachment",
    "mimeType": "text/plain",
    "name": "brief.txt",
    "size": 19
  }
}
```

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `POST /threads/{threadId}/files/uploads`

Purpose: Creates a files API upload session scoped to this Thread. The request only supplies file metadata; Mosoo resolves the backing App and Session from the Thread and creates a session attachment upload. MVP public uploads must be 67108864 bytes or fewer and use single PUT.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Request body:

* `application/json`: `CreateThreadFileUploadRequest`

Example:

```json theme={null}
{
  "file": {
    "contentType": "text/plain",
    "name": "brief.txt",
    "size": 19
  }
}
```

Success responses:

* `201`: Created files API upload session. (`application/json` -> `FileUploadSummary`)

Example `201`:

```json theme={null}
{
  "contentType": "text/plain",
  "expectedSize": 19,
  "expiresAt": "2026-05-20T00:02:00.000Z",
  "fileId": "01J0000000000000000000000J",
  "partSize": null,
  "path": "session-files/01J0000000000000000000000J/brief.txt",
  "status": "pending",
  "strategy": "single_put"
}
```

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `DELETE /threads/{threadId}/files/{fileId}`

Purpose: Detaches a file from the Thread.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.
* `fileId` required, `string(ulid)`. File ID returned by add or list Thread files. v1 IDs are bare ULIDs.

Success responses:

* `200`: Removed. (`application/json` -> `object`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

### `POST /threads/{threadId}/unarchive`

Purpose: Restores a previously archived Thread to active Thread lists.

Path params:

* `threadId` required, `string(ulid)`. Thread ID returned by create thread. v1 IDs are bare ULIDs.

Success responses:

* `200`: Unarchived. (`application/json` -> `object`)

Error responses:

* `400` `InvalidRequest`: The request shape or query value is invalid.
* `401` `Unauthenticated`: A valid API token is required.
* `403` `Forbidden`: This operation is not allowed for this Agent.
* `404` `NotFound`: The resource was not found in the current Mosoo workspace.
* `409` `Conflict`: The Agent/session state rejects this action, or an Idempotency-Key is already processing or was reused for a different request.
* `429` `RateLimited`: The API token exceeded the public API request budget for the current window.
* `500` `InternalError`: The request failed unexpectedly.

## Schema quick reference

### `ThreadEventInput`

A single event posted to a Thread. Exactly one variant applies: send a user message, answer a pending permission request, or interrupt a running Run.

Variants:

1. `object`: Send a new user message into the Thread, optionally with file attachments.
   * `attachmentIds` optional, `string(ulid)[]`. File IDs (bare ULIDs) to attach to this message. Each file must already be uploaded and claimed into this Thread.
   * `clientRequestId` optional, `string | null`. Optional caller-supplied correlation ID echoed back on the matching event result so you can pair responses with the message you sent.
   * `text` required, `string`. The user message text. Must not be empty.
   * `type` required, `"user_message"`. Discriminator selecting the send-user-message variant.

2. `object`: Answer a permission request the Agent raised while waiting for input (for example a tool confirmation).
   * `decision` required, `"allow_once" | "reject_once"`. Whether to allow or reject the requested action for this single occurrence. `allow_once` permits it now; `reject_once` denies it now.
   * `requestId` required, `string`. ID of the permission request being answered, taken from the corresponding `tool.confirmation.required` event.
   * `type` required, `"permission_decision"`. Discriminator selecting the permission-decision variant.

3. `object`: Interrupt a Run that is currently executing on the Thread.
   * `runId` optional, `string(ulid) | null`. Run ID (bare ULID) to interrupt. Omit or send null to interrupt the Thread's current Run.
   * `type` required, `"user_interrupt"`. Discriminator selecting the interrupt variant.

### `SendEventsRequest`

Request body for posting a batch of events to a Thread.

Fields:

* `events` required, `ThreadEventInput[]`. Ordered list of events to apply to the Thread. At least one is required.

### `CreateThreadFileRequest`

Request body for claiming a previously uploaded draft file handle into a Thread.

Fields:

* `fileId` required, `string(ulid)`. ID of a ready draft file (bare ULID) previously uploaded through the files data plane by the same API token.

### `CreateThreadFileUploadRequest`

Request body for creating a files API upload session scoped to this Thread.

Fields:

* `file` required, `object`. Metadata for the file bytes that will be uploaded.
  * `contentType` required, `string`. MIME type for the file bytes.
  * `name` required, `string`. Original file name to record for this Thread attachment.
  * `size` required, `integer`. File size in bytes. Public Thread upload creation currently accepts single-put uploads only.

### `FileUploadSummary`

Files API upload session summary returned after creating an upload target.

Fields:

* `contentType` required, `string`. Normalized MIME type the upload expects.
* `expectedSize` required, `integer`. Expected file size in bytes for integrity validation.
* `expiresAt` required, `string(date-time)`. Timestamp (RFC 3339) when this upload session expires.
* `fileId` required, `string(ulid)`. File ID (bare ULID) for this upload session.
* `partSize` required, `integer | null`. Multipart part size in bytes, or null when the upload uses single PUT.
* `path` required, `string`. Materialized file path exposed by the files API upload summary.
* `status` required, `"aborted" | "completed" | "completing" | "expired" | "failed" | "pending" | "uploading"`. Current files API upload session status.
* `strategy` required, `"multipart" | "single_put"`. Files API upload transfer strategy selected for the file size.

### `CompleteFileUploadPart`

A completed multipart upload part. Public Thread MVP uploads use single PUT and do not send parts.

Fields:

* `etag` required, `string`. ETag returned by the uploaded multipart part.
* `partNumber` required, `integer`. One-based multipart part number for completion ordering.

### `CompleteFileUploadRequest`

Request body for completing a files API upload session. Public Thread MVP single PUT uploads send an empty object.

Fields:

* `parts` optional, `CompleteFileUploadPart[]`. Multipart completion parts. Omit this field for public MVP single PUT uploads.

### `FileEntry`

Files API file entry returned after an upload completes.

Fields:

* `createdAt` required, `string(date-time)`. Timestamp (RFC 3339) at which the file record was created.
* `createdBy` required, `string(ulid)`. Account ID (bare ULID) that created the file record.
* `etag` required, `string | null`. Storage ETag for the finalized object, or null when unavailable.
* `expiresAt` required, `string | null(date-time)`. Expiration timestamp for temporary files, or null for durable Thread files.
* `id` required, `string(ulid)`. File ID (bare ULID) for the finalized file.
* `mimeType` required, `string | null`. Stored MIME type for the file bytes, or null when unknown.
* `name` required, `string`. Original file name recorded for this file.
* `path` required, `string`. Files API record path for the finalized file.
* `sessionKind` required, `"artifact" | "attachment" | null`. Session file kind for Thread-scoped files, or null for non-session files.
* `size` required, `integer`. Finalized file size in bytes.
* `status` required, `"deleting" | "failed" | "pending" | "ready"`. Current files API file lifecycle status.
* `updatedAt` required, `string(date-time)`. Timestamp (RFC 3339) at which the file record was last updated.
* `version` required, `integer`. Monotonic file version number within its path.

### `CompleteFileUploadResponse`

Result of completing a files API upload session.

Fields:

* `file` required, `FileEntry`. Completed files API file entry.

### `ErrorResponse`

Standard error envelope returned for any non-2xx public API response.

Fields:

* `error` required, `object`. Details about why the request failed.
  * `code` required, `"agent_not_published" | "forbidden" | "idempotency_conflict" | "internal_error" | "invalid_json" | "invalid_request" | "not_found" | "rate_limited" | "readiness_blocked" | "service_inactive" | "unauthenticated"`. Stable, machine-readable error code you can branch on.
  * `message` required, `string`. Human-readable explanation of the error. Not intended for end users.

### `SendEventsResponse`

Result of accepting a batch of events posted to a Thread.

Fields:

* `acceptedAt` required, `string(date-time)`. Timestamp (RFC 3339) at which the event batch was accepted for processing.
* `events` required, `ThreadEventResult[]`. Per-event outcomes, in the same order as the submitted events.
* `thread` required, `ThreadSummary`. The Thread state after applying the batch.
* `warnings` required, `UserWarning[]`. Non-fatal warnings raised while accepting the batch (for example a partially honored request). Empty when there are none.

### `ThreadEventLogEntry`

A single public event log entry for a Thread. This is the stable read surface and never exposes raw runtime payloads, transcripts, or diagnostics.

Fields:

* `content` required, `string`. Public content of the event — typically a reference to the associated payload (such as a message ID) rather than the raw runtime data.
* `durationMs` required, `integer | null`. Wall-clock duration of the event in milliseconds, when applicable (for example a completed Run). Null when not measured.
* `id` required, `string(ulid)`. Unique event ID (bare ULID), monotonically increasing in chronological order.
* `occurredAt` required, `string(date-time)`. Timestamp (RFC 3339) at which the event occurred.
* `status` required, `"available" | "error" | "unsupported"`. Delivery status of the event: `available` when the event is fully populated, `error` when it failed, `unsupported` when this event type cannot be rendered on the public surface.
* `tokens` required, `integer | null`. Token count associated with the event when applicable (for example model usage). Null when not measured.
* `type` required, `"agent.message.delta" | "agent.thinking.delta" | "file.changed" | "run.completed" | "run.failed" | "run.started" | "session.status" | "session_files.updated" | "tool.confirmation.required" | "tool.use.completed" | "tool.use.started" | "usage.updated" | "user.message"`. Event type, such as `run.started`, `run.completed`, `agent.message.delta`, or `tool.use.started`.

### `ThreadEventListResponse`

A page of the latest Thread event log entries in chronological order.

Fields:

* `events` required, `ThreadEventLogEntry[]`. The returned event log entries, oldest first within the requested window.
* `truncated` required, `boolean`. True when older events exist beyond the returned window because the requested limit was reached.

### `ThreadEventResult`

Outcome of a single submitted event.

Fields:

* `clientRequestId` required, `string | null`. The `clientRequestId` echoed from the submitted user message, or null when none was provided or the event type does not carry one.
* `run` required, `RunSummary | null`. The Run created or affected by this event, or null when the event did not start or change a Run.
* `type` required, `"permission_decision" | "user_interrupt" | "user_message"`. The kind of event this result corresponds to.

### `CreateThreadRequest`

Request body for creating a Thread. All fields are optional: omit `input` to create an empty IDLE Thread, or include it to queue the initial Run.

Fields:

* `client_external_ref` optional, `string`. Optional client-owned reference (for example an external ticket key) stored on the Thread for correlation. Not unique and not validated by Mosoo.
* `files` optional, `object[]`. Draft file handles uploaded by the same API token.
* `input` optional, `object`. Initial user message that seeds the Thread and queues the first Run. Omit to create an empty Thread with no run.
  * `content` required, `object[]`. Ordered content parts that make up the initial message.
  * `type` required, `"user.message"`. Discriminator for the initial input. Always `user.message`.

### `CreateThreadResponse`

Result of creating a Thread.

Fields:

* `links` required, `ThreadLinks`. Convenience links for the created Thread.
* `run` required, `RunSummary | null`. The initial Run queued when `input` was provided, or null when an empty Thread was created.
* `thread` required, `ThreadSummary`. The created Thread.

### `RetrieveThreadResponse`

Current state of a Thread.

Fields:

* `links` required, `ThreadLinks`. Convenience links for the Thread.
* `run` required, `RunSummary | null`. The Thread's most recent Run, or null when no Run has been created yet.
* `thread` required, `ThreadSummary`. The Thread summary.

### `ThreadFile`

A file associated with a Thread.

Fields:

* `committed` required, `boolean`. True once the file is durably attached to the Thread; false while it is still a draft handle.
* `createdAt` required, `string(date-time)`. Timestamp (RFC 3339) at which the file was created.
* `id` required, `string(ulid)`. Unique file ID (bare ULID).
* `kind` required, `"attachment" | "artifact"`. Files added through the public API are attachments; artifacts are files produced by the Agent.
* `mimeType` required, `string | null`. Detected MIME type of the file, or null when unknown.
* `name` required, `string`. Original file name.
* `size` required, `integer`. File size in bytes.

### `ThreadFileListResponse`

List of files attached to a Thread.

Fields:

* `files` required, `ThreadFile[]`. The Thread's files.

### `ThreadFileResponse`

A single Thread file.

Fields:

* `file` required, `ThreadFile`. The Thread file metadata, including its identifier, name, MIME type, and origin.

### `ThreadAttributedUser`

The account a Thread is attributed to (the current Mosoo user).

Fields:

* `id` required, `string(ulid)`. Account ID (bare ULID) the Thread is attributed to.

### `ThreadCaller`

The credential that created the Thread.

Fields:

* `id` required, `string(ulid)`. ID (bare ULID) of the caller that created the Thread.
* `kind` required, `"access_token"`. Caller credential type. `access_token` identifies an API token.

### `ThreadLinks`

Convenience links for a Thread.

Fields:

* `thread` required, `string`. Absolute API URL of the Thread resource.

### `ThreadSummary`

Summary of a Thread on a Agent API Endpoint.

Fields:

* `agent_id` required, `string(ulid)`. ID (bare ULID) of the Agent API Endpoint this Thread belongs to.
* `attributed_user` required, `ThreadAttributedUser | null`. The account the Thread is attributed to, or null when not attributed to an account.
* `client_external_ref` required, `string | null`. The client-owned reference supplied at creation, or null when none was provided.
* `created_at` required, `string(date-time)`. Timestamp (RFC 3339) at which the Thread was created.
* `created_by` required, `ThreadCaller`. The credential that created the Thread.
* `id` required, `string(ulid)`. Unique Thread ID (bare ULID).
* `kind` required, `"pet" | "cattle"`. Agent kind backing this Thread (for example a persistent or one-off Agent API Endpoint).
* `last_run_id` required, `string(ulid) | null`. ID (bare ULID) of the most recent Run, or null when no Run exists yet.
* `source` required, `"api"`. Origin of the Thread. Always `api` for Threads created via this API.
* `status` required, `"IDLE" | "RUNNING" | "RESCHEDULING" | "TERMINATED"`. Lifecycle status of the Thread: `IDLE` (no active run), `RUNNING` (a Run is executing), `RESCHEDULING` (between runs), or `TERMINATED` (ended).
* `title` required, `string | null`. Human-readable Thread title, or null when one has not been derived yet.
* `updated_at` required, `string(date-time)`. Timestamp (RFC 3339) of the most recent change to the Thread.

### `RunSummary`

Summary of a single Agent Run on a Thread.

Fields:

* `completedAt` required, `string | null(date-time)`. Timestamp (RFC 3339) at which the Run reached a terminal state, or null while it has not finished.
* `createdAt` required, `string(date-time)`. Timestamp (RFC 3339) at which the Run was created.
* `id` required, `string(ulid)`. Unique Run ID (bare ULID).
* `startedAt` required, `string | null(date-time)`. Timestamp (RFC 3339) at which the Run began executing, or null while it is still queued.
* `status` required, `"queued" | "booting" | "running" | "waiting_input" | "completed" | "failed" | "cancelled" | "expired"`. Current Run status. `queued` and `booting` precede execution; `running` and `waiting_input` are active; `completed`, `failed`, `cancelled`, and `expired` are terminal.
* `trigger` required, `"user_prompt" | "retry" | "resume" | "system"`. What started the Run: `user_prompt` (a user message), `retry`, `resume`, or `system`.
* `updatedAt` required, `string(date-time)`. Timestamp (RFC 3339) of the most recent change to the Run.

### `UserWarning`

A non-fatal warning surfaced to the caller.

Fields:

* `code` required, `string`. Stable, machine-readable warning code.
* `message` required, `string`. Human-readable explanation of the warning.

## Implementation guardrails for coding agents

* Always send `Authorization: Bearer $MOSOO_API_TOKEN`.
* Always treat `agentId`, `threadId`, `fileId`, and `runId` as bare ULIDs.
* Use `Idempotency-Key` for create-thread and send-events retries.
* Use event log APIs as the source for public results.
* Do not assume raw model output, private runtime payloads, or internal diagnostics are exposed.
* Do not send provider credentials, model configuration, channel credentials, or Agent configuration through this public API.
* Do not retry invalid requests unchanged.
* Do not treat `403` as an authentication problem; `403` means the token was understood but the operation is not allowed for that resource.
* Do not treat a published Agent as ready unless create-thread succeeds or the user confirms Agent readiness in Mosoo.
* Prefer stable caller-owned references such as `client_external_ref` and `clientRequestId` for correlation.
