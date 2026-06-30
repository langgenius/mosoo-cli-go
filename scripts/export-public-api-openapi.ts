import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { createPublicApiOpenApiDocument } from "../.cache/mosoo/apps/api/src/adapters/http/routes/public-api-openapi.ts";

type HttpMethod = "delete" | "get" | "post";

interface OperationMetadata {
	operationId: string;
	tags: string[];
}

const operationMetadata: Record<string, Partial<Record<HttpMethod, OperationMetadata>>> = {
	"/agents/{agentId}/threads": {
		get: { operationId: "Threads_ListForAgent", tags: ["Threads"] },
		post: { operationId: "Threads_Create", tags: ["Threads"] },
	},
	"/threads/{threadId}": {
		delete: { operationId: "Threads_Delete", tags: ["Threads"] },
		get: { operationId: "Threads_Retrieve", tags: ["Threads"] },
	},
	"/threads/{threadId}/archive": {
		post: { operationId: "Threads_Archive", tags: ["Threads"] },
	},
	"/threads/{threadId}/events": {
		get: { operationId: "ThreadEvents_ListEvents", tags: ["Events"] },
		post: { operationId: "ThreadEvents_Send", tags: ["Events"] },
	},
	"/threads/{threadId}/events/stream": {
		get: { operationId: "ThreadEvents_Stream", tags: ["Events"] },
	},
	"/threads/{threadId}/files": {
		get: { operationId: "ThreadFiles_ListFiles", tags: ["Files"] },
		post: { operationId: "ThreadFiles_Add", tags: ["Files"] },
	},
	"/threads/{threadId}/files/{fileId}": {
		delete: { operationId: "ThreadFiles_Remove", tags: ["Files"] },
	},
	"/threads/{threadId}/unarchive": {
		post: { operationId: "Threads_Unarchive", tags: ["Threads"] },
	},
};

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "..");
const outputPath = resolve(repositoryRoot, ".cache/mosoo/docs/openapi/public-thread-api.openapi.json");
const committedOrigin = "https://try.mosoo.ai";

const document = createPublicApiOpenApiDocument(committedOrigin);
for (const [path, methods] of Object.entries(operationMetadata)) {
	const pathItem = document.paths[path];
	if (pathItem === undefined) {
		throw new Error(`OpenAPI path missing: ${path}`);
	}

	for (const [method, metadata] of Object.entries(methods) as [HttpMethod, OperationMetadata][]) {
		const operation = pathItem[method];
		if (operation === undefined) {
			throw new Error(`OpenAPI operation missing: ${method.toUpperCase()} ${path}`);
		}

		Object.assign(operation, metadata);
	}
}
await mkdir(dirname(outputPath), { recursive: true });
await writeFile(outputPath, `${JSON.stringify(document, null, 2)}\n`, "utf8");
console.log(`wrote ${outputPath}`);
