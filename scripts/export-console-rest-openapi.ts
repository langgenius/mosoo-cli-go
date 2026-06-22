import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { createPublicApiPlatformIdSchema } from "../.cache/mosoo/pkgs/contracts/src/http/public-api-openapi.contract.ts";

type HttpMethod = "delete" | "get" | "patch" | "post" | "put";

interface OpenApiParameter {
	description?: string;
	example?: unknown;
	in: "header" | "path" | "query";
	name: string;
	required?: boolean;
	schema: Record<string, unknown>;
}

interface OpenApiOperation {
	description?: string;
	operationId: string;
	parameters?: OpenApiParameter[];
	requestBody?: Record<string, unknown>;
	responses: Record<string, unknown>;
	security?: Array<Record<string, []>>;
	summary: string;
	tags: string[];
}

type OpenApiPaths = Record<string, Partial<Record<HttpMethod, OpenApiOperation>>>;

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "..");
const outputPath = resolve(repositoryRoot, ".cache/mosoo/docs/openapi/console-rest.openapi.json");
const committedOrigin = "https://mosoo.ai";

const EXAMPLE_APP_ID = "01J00000000000000000000001";
const EXAMPLE_FILE_ID = "01J0000000000000000000000J";
const EXAMPLE_TOKEN_ID = "01J0000000000000000000000T";

const ulidSchema = (example: string) => createPublicApiPlatformIdSchema({ example });

function platformIdPathParameter(input: {
	description: string;
	example: string;
	name: string;
}): OpenApiParameter {
	return {
		description: input.description,
		example: input.example,
		in: "path",
		name: input.name,
		required: true,
		schema: {
			...ulidSchema(input.example),
			"x-default": input.example,
		},
	};
}

function jsonResponse(description: string, schema: Record<string, unknown>) {
	return {
		content: {
			"application/json": { schema },
		},
		description,
	};
}

function errorResponse(description: string) {
	return jsonResponse(description, {
		type: "object",
		properties: {
			error: { type: "string" },
		},
	});
}

const bearerSecurity: Array<Record<string, []>> = [{ consoleBearer: [] }];

const paths: OpenApiPaths = {
	"/access-tokens": {
		get: {
			operationId: "AccessTokens_List",
			summary: "List personal access tokens",
			description:
				"Lists active personal access tokens for the authenticated account. Today this route accepts a viewer session cookie; Bearer PAT support is planned.",
			tags: ["Access Tokens"],
			security: bearerSecurity,
			responses: {
				"200": jsonResponse("Token list.", {
					type: "object",
					properties: {
						tokens: { type: "array", items: { type: "object" } },
					},
				}),
				"401": errorResponse("Authentication required."),
			},
		},
		post: {
			operationId: "AccessTokens_Create",
			summary: "Create a personal access token",
			description:
				"Creates a new Mosoo access token (mst_...) for CLI and API use. Today this route accepts a viewer session cookie; Bearer PAT support is planned.",
			tags: ["Access Tokens"],
			security: bearerSecurity,
			requestBody: {
				required: true,
				content: {
					"application/json": {
						schema: {
							type: "object",
							required: ["label"],
							properties: {
								label: {
									type: "string",
									description: "Human-readable label for the token.",
									maxLength: 80,
								},
							},
						},
					},
				},
			},
			responses: {
				"201": jsonResponse("Created token including one-time token value.", { type: "object" }),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/access-tokens/{tokenId}": {
		delete: {
			operationId: "AccessTokens_Revoke",
			summary: "Revoke a personal access token",
			tags: ["Access Tokens"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "Personal access token ID.",
					example: EXAMPLE_TOKEN_ID,
					name: "tokenId",
				}),
			],
			responses: {
				"200": jsonResponse("Revoke acknowledgement.", {
					type: "object",
					properties: { ok: { type: "boolean" } },
				}),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/files": {
		get: {
			operationId: "Files_List",
			summary: "List files for an app or session",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				{
					description: "App ID that owns the files.",
					example: EXAMPLE_APP_ID,
					in: "query",
					name: "appId",
					required: true,
					schema: ulidSchema(EXAMPLE_APP_ID),
				},
				{
					description: "Optional session ID filter.",
					in: "query",
					name: "sessionId",
					schema: ulidSchema(EXAMPLE_FILE_ID),
				},
				{
					description: "Optional session file kind filter.",
					in: "query",
					name: "sessionKind",
					schema: { enum: ["artifact", "attachment", "all"], type: "string" },
				},
			],
			responses: {
				"200": jsonResponse("File listing.", {
					type: "object",
					properties: {
						files: { type: "array", items: { type: "object" } },
					},
				}),
				"401": errorResponse("Authentication required."),
			},
		},
		post: {
			operationId: "Files_CreateUpload",
			summary: "Create a file upload",
			description: "Starts a session-scoped file upload and returns upload instructions.",
			tags: ["Files"],
			security: bearerSecurity,
			requestBody: {
				required: true,
				content: {
					"application/json": {
						schema: {
							type: "object",
							required: ["name", "target"],
							properties: {
								name: { type: "string" },
								mimeType: { type: "string" },
								size: { type: "integer", minimum: 1 },
								target: {
									type: "object",
									required: ["kind"],
									properties: {
										kind: { enum: ["session"], type: "string" },
										appId: ulidSchema(EXAMPLE_APP_ID),
										sessionId: ulidSchema(EXAMPLE_FILE_ID),
										sessionKind: { enum: ["artifact", "attachment"], type: "string" },
									},
								},
							},
						},
					},
				},
			},
			responses: {
				"200": jsonResponse("Upload session created.", { type: "object" }),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/files/{fileId}": {
		delete: {
			operationId: "Files_Delete",
			summary: "Delete a file",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "File ID.",
					example: EXAMPLE_FILE_ID,
					name: "fileId",
				}),
				{
					description: "Optional optimistic concurrency token.",
					in: "header",
					name: "If-Match",
					schema: { type: "string" },
				},
			],
			responses: {
				"200": jsonResponse("Delete acknowledgement.", {
					type: "object",
					properties: { ok: { type: "boolean" } },
				}),
				"401": errorResponse("Authentication required."),
			},
		},
		patch: {
			operationId: "Files_Update",
			summary: "Update file metadata",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "File ID.",
					example: EXAMPLE_FILE_ID,
					name: "fileId",
				}),
			],
			requestBody: {
				required: true,
				content: {
					"application/json": {
						schema: {
							type: "object",
							properties: {
								name: { type: "string" },
								path: { type: "string" },
							},
						},
					},
				},
			},
			responses: {
				"200": jsonResponse("Updated file record.", { type: "object" }),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/files/{fileId}/content": {
		get: {
			operationId: "Files_DownloadContent",
			summary: "Download file content",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "File ID.",
					example: EXAMPLE_FILE_ID,
					name: "fileId",
				}),
				{
					description: "Content disposition hint.",
					in: "query",
					name: "disposition",
					schema: { enum: ["attachment", "inline"], type: "string" },
				},
			],
			responses: {
				"200": {
					content: {
						"application/octet-stream": { schema: { type: "string", format: "binary" } },
					},
					description: "File bytes.",
				},
				"401": errorResponse("Authentication required."),
			},
		},
		put: {
			operationId: "Files_UploadContent",
			summary: "Upload file content in a single request",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "File ID.",
					example: EXAMPLE_FILE_ID,
					name: "fileId",
				}),
			],
			requestBody: {
				required: true,
				content: {
					"application/octet-stream": { schema: { type: "string", format: "binary" } },
				},
			},
			responses: {
				"200": jsonResponse("Upload acknowledgement.", {
					type: "object",
					properties: { ok: { type: "boolean" } },
				}),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/files/{fileId}/upload": {
		delete: {
			operationId: "Files_AbortUpload",
			summary: "Abort a pending upload",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "File ID.",
					example: EXAMPLE_FILE_ID,
					name: "fileId",
				}),
			],
			responses: {
				"200": jsonResponse("Abort acknowledgement.", {
					type: "object",
					properties: { ok: { type: "boolean" } },
				}),
				"401": errorResponse("Authentication required."),
			},
		},
		get: {
			operationId: "Files_GetUpload",
			summary: "Get upload session status",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "File ID.",
					example: EXAMPLE_FILE_ID,
					name: "fileId",
				}),
			],
			responses: {
				"200": jsonResponse("Upload session.", { type: "object" }),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/files/{fileId}/complete": {
		post: {
			operationId: "Files_CompleteUpload",
			summary: "Complete a multipart upload",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "File ID.",
					example: EXAMPLE_FILE_ID,
					name: "fileId",
				}),
			],
			requestBody: {
				required: true,
				content: {
					"application/json": {
						schema: {
							type: "object",
							properties: {
								parts: {
									type: "array",
									items: {
										type: "object",
										properties: {
											etag: { type: "string" },
											partNumber: { type: "integer" },
										},
									},
								},
							},
						},
					},
				},
			},
			responses: {
				"200": jsonResponse("Completed upload.", { type: "object" }),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/files/{fileId}/parts/{partNumber}": {
		put: {
			operationId: "Files_UploadPart",
			summary: "Upload one multipart part",
			tags: ["Files"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "File ID.",
					example: EXAMPLE_FILE_ID,
					name: "fileId",
				}),
				{
					description: "1-based multipart part number.",
					in: "path",
					name: "partNumber",
					required: true,
					schema: { minimum: 1, type: "integer" },
				},
			],
			requestBody: {
				required: true,
				content: {
					"application/octet-stream": { schema: { type: "string", format: "binary" } },
				},
			},
			responses: {
				"200": jsonResponse("Uploaded part metadata.", { type: "object" }),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/skill/inspect": {
		post: {
			operationId: "Skill_Inspect",
			summary: "Inspect a skill package from upload or GitHub URL",
			tags: ["Skills"],
			security: bearerSecurity,
			requestBody: {
				required: true,
				content: {
					"multipart/form-data": {
						schema: {
							type: "object",
							properties: {
								file: { type: "string", format: "binary" },
								githubUrl: { type: "string" },
							},
						},
					},
				},
			},
			responses: {
				"200": jsonResponse("Inspection result.", { type: "object" }),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/skill/package": {
		post: {
			operationId: "Skill_Package",
			summary: "Create or update a skill package upload",
			tags: ["Skills"],
			security: bearerSecurity,
			requestBody: {
				required: true,
				content: {
					"multipart/form-data": {
						schema: {
							type: "object",
							required: ["appId"],
							properties: {
								appId: ulidSchema(EXAMPLE_APP_ID),
								skillId: ulidSchema(EXAMPLE_FILE_ID),
								file: { type: "string", format: "binary" },
								githubUrl: { type: "string" },
							},
						},
					},
				},
			},
			responses: {
				"200": jsonResponse("Skill package write result.", { type: "object" }),
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/skill/{skillId}/package": {
		get: {
			operationId: "Skill_DownloadPackage",
			summary: "Download a skill package archive",
			tags: ["Skills"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "Skill ID.",
					example: EXAMPLE_FILE_ID,
					name: "skillId",
				}),
				{
					description: "App ID that owns the skill.",
					example: EXAMPLE_APP_ID,
					in: "query",
					name: "appId",
					required: true,
					schema: ulidSchema(EXAMPLE_APP_ID),
				},
			],
			responses: {
				"200": {
					content: {
						"application/zip": { schema: { type: "string", format: "binary" } },
					},
					description: "Skill package zip archive.",
				},
				"401": errorResponse("Authentication required."),
			},
		},
	},
	"/skill/{skillId}/source": {
		get: {
			operationId: "Skill_DownloadSource",
			summary: "Download skill source markdown",
			tags: ["Skills"],
			security: bearerSecurity,
			parameters: [
				platformIdPathParameter({
					description: "Skill ID.",
					example: EXAMPLE_FILE_ID,
					name: "skillId",
				}),
				{
					description: "App ID that owns the skill.",
					example: EXAMPLE_APP_ID,
					in: "query",
					name: "appId",
					required: true,
					schema: ulidSchema(EXAMPLE_APP_ID),
				},
			],
			responses: {
				"200": {
					content: {
						"text/markdown": { schema: { type: "string" } },
					},
					description: "Skill source markdown.",
				},
				"401": errorResponse("Authentication required."),
			},
		},
	},
};

const document = {
	openapi: "3.1.0",
	info: {
		title: "Mosoo Console REST API",
		description:
			"Viewer-authenticated REST helpers for files, access tokens, and skill packages. Routes are mounted under /api.",
		version: "v1",
	},
	servers: [{ url: `${committedOrigin}/api` }],
	security: bearerSecurity,
	components: {
		securitySchemes: {
			consoleBearer: {
				type: "http",
				scheme: "bearer",
				bearerFormat: "Mosoo Access Token or viewer session",
				description:
					"Send Authorization: Bearer mst_... for PAT-authenticated calls. Some routes still require a viewer session cookie until PAT support lands on these paths.",
			},
		},
	},
	paths,
};

await mkdir(dirname(outputPath), { recursive: true });
await writeFile(outputPath, `${JSON.stringify(document, null, 2)}\n`, "utf8");
console.log(`wrote ${outputPath} (${Object.keys(paths).length} paths)`);
