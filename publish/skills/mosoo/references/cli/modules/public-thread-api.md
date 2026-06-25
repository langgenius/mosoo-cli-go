# Module `public-thread-api`

## Source

- Backend: `openapi3`
- Default hostname: `http://127.0.0.1:8787/api/v1`
- Repository: https://github.com/langgenius/mosoo.git
- Pinned tag: `local-snapshot`
- Files: `docs/openapi/public-thread-api.openapi.json`
- Resolved SHA: `local-snapshot`

## Events

### `mosoo public-thread-api events list-events`

- Summary: List thread events
- HTTP: `GET /threads/{threadId}/events`
- Auth: required
- Body: none
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
  - `--limit` (query, default `100`): Maximum number of latest Thread events to return.
- Output: list path `events`; columns `type`, `id`, `content`, `durationMs`, `occurredAt`, `runId`; response media `application/json`; pagination `cursor`
- Known errors:
  - HTTP 401: Invalid personal access token.
  - HTTP 404: Thread not found for this caller.
- Example: `mosoo public-thread-api events list-events --thread-id <thread-id>`

### `mosoo public-thread-api events send`

- Summary: Send events to a thread
- HTTP: `POST /threads/{threadId}/events`
- Auth: required
- Body: required; media type `application/json`
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
  - `--idempotency-key` (header): Optional key for retry-safe create-thread and send-events calls. Reusing the same key with the same request returns the original response. Reusing the key while the original request is still processing returns 409.
- Output: list path `events`; columns `type`, `clientRequestId`, `run`; response media `application/json`
- Known errors:
  - HTTP 401: Invalid personal access token.
  - HTTP 409: Idempotency key reused while the original request is still processing.
- Example: `mosoo public-thread-api events send --thread-id <thread-id> --file events.json`

### `mosoo public-thread-api events stream`

- Summary: Stream thread events (SSE)
- HTTP: `GET /threads/{threadId}/events/stream`
- Auth: required
- Body: none
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
  - `--limit` (query, default `100`): Maximum number of latest Thread events to return.
- Output: response media `text/event-stream`; pagination `cursor`; streaming `sse`
- Example: `mosoo public-thread-api events stream --thread-id <thread-id> -o raw`

## Files

### `mosoo public-thread-api files add`

- Summary: Attach a file to a thread
- HTTP: `POST /threads/{threadId}/files`
- Auth: required
- Body: required; media type `application/json`
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
- Example: `mosoo public-thread-api files add --thread-id <thread-id> --set fileId=<file-id>`

### `mosoo public-thread-api files list-files`

- Summary: List thread files
- HTTP: `GET /threads/{threadId}/files`
- Auth: required
- Body: none
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
- Output: list path `files`; columns `name`, `kind`, `id`, `committed`, `createdAt`, `mimeType`; response media `application/json`
- Example: `mosoo public-thread-api files list-files --thread-id <thread-id>`

### `mosoo public-thread-api files remove`

- Summary: Remove a thread file
- HTTP: `DELETE /threads/{threadId}/files/{fileId}`
- Auth: required
- Body: none
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
  - `--file-id` (path, required, ulid): File ID returned by add or list Thread files. v1 IDs are bare ULIDs.
- Output: response media `application/json`
- Example: `mosoo public-thread-api files remove --thread-id <thread-id> --file-id <file-id>`

## Threads

### `mosoo public-thread-api threads archive`

- Summary: Archive a thread
- HTTP: `POST /threads/{threadId}/archive`
- Auth: required
- Body: none
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
- Output: response media `application/json`
- Example: `mosoo public-thread-api threads archive --thread-id <thread-id>`

### `mosoo public-thread-api threads create`

- Summary: Create a thread for an agent
- HTTP: `POST /agents/{agentId}/threads`
- Auth: required
- Body: optional; media type `application/json`
- Flags:
  - `--agent-id` (path, required, ulid): Agent API Endpoint ID from the Agent's API Access panel. v1 IDs are bare ULIDs.
  - `--idempotency-key` (header): Optional key for retry-safe create-thread and send-events calls. Reusing the same key with the same request returns the original response. Reusing the key while the original request is still processing returns 409.
- Known errors:
  - HTTP 404: Agent not found or not accessible to this token.
- Example: `mosoo public-thread-api threads create --agent-id <agent-id> --file body.json`

### `mosoo public-thread-api threads delete`

- Summary: Delete a thread
- HTTP: `DELETE /threads/{threadId}`
- Auth: required
- Body: none
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
- Output: response media `application/json`
- Example: `mosoo public-thread-api threads delete --thread-id <thread-id>`

### `mosoo public-thread-api threads list-for-agent`

- Summary: List threads for an agent
- HTTP: `GET /agents/{agentId}/threads`
- Auth: required
- Body: none
- Flags:
  - `--agent-id` (path, required, ulid): Agent API Endpoint ID from the Agent's API Access panel. v1 IDs are bare ULIDs.
  - `--archived` (query): Filter by archived state: true returns only archived Threads, false only active ones. Omit to return all Threads.
- Output: list path `threads`; columns `kind`, `id`, `agent_id`, `attributed_user`, `client_external_ref`, `created_at`; response media `application/json`
- Example: `mosoo public-thread-api threads list-for-agent --agent-id <agent-id>`

### `mosoo public-thread-api threads retrieve`

- Summary: Retrieve a thread
- HTTP: `GET /threads/{threadId}`
- Auth: required
- Body: none
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
- Output: response media `application/json`
- Example: `mosoo public-thread-api threads retrieve --thread-id <thread-id>`

### `mosoo public-thread-api threads unarchive`

- Summary: Unarchive a thread
- HTTP: `POST /threads/{threadId}/unarchive`
- Auth: required
- Body: none
- Flags:
  - `--thread-id` (path, required, ulid): Thread ID returned by create thread. v1 IDs are bare ULIDs.
- Output: response media `application/json`
- Example: `mosoo public-thread-api threads unarchive --thread-id <thread-id>`
