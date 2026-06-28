import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { collectConsoleGraphQLOperations } from "./console-graphql-sources.ts";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const overlayDir = resolve(dirname(fileURLToPath(import.meta.url)), "..", "overlays");

type OverlayCommand = {
	use?: string;
	aliases?: string[];
	shortcuts?: OverlayShortcut[];
	short?: string;
	long?: string;
	example?: string;
	examples?: OverlayExample[];
	hidden?: boolean;
	ignore?: boolean;
	notes?: string[];
	prerequisites?: string[];
	known_errors?: { status: number; cause: string }[];
};

type OverlayShortcut = {
	use: string;
	params?: Record<string, string>;
};

type OverlayExample = {
	summary: string;
	command: string;
	body_shape?: Record<string, unknown>;
	output_hints?: {
		id_path?: string;
		list_path?: string;
	};
	follow_up_commands?: string[];
};

const exampleFields = new Set([
	"viewer",
	"appList",
	"appOverview",
	"controlPlaneOverview",
	"createApp",
	"createAgent",
	"createVendorCredential",
	"publishAgent",
	"startAgentRun",
	"testVendorCredential",
	"threadSessionProcessEvents",
	"vendorCredentialList",
	"agentSessionList",
	"sessionList",
]);
const noArticleVerbs = new Set(["set", "poll", "start", "test"]);

const consoleCommandOverrides: Record<string, OverlayCommand> = {
	"agent-manifest": {
		use: "manifest",
		aliases: ["agent-manifest"],
		short: "Get an agent manifest",
		long:
			"Read the current remote Agent manifest through the raw Console GraphQL API. Prefer `mosoo agent manifest probe`, `mosoo agent manifest diff`, and `mosoo agent manifest apply` for editable YAML workflows.",
		example: "mosoo console agents manifest --app-id <app-id> --agent-id <agent-id> -o json",
		hidden: true,
		notes: [
			"Raw API command. The product workflow command is `mosoo agent manifest probe`.",
			"Uses POST /graphql on the console default hostname (/api).",
			"Pull this before changing Agent config; treat the returned manifest/YAML as the source of truth.",
		],
	},
	"update-agent-config": {
		ignore: true,
		use: "update-config",
		aliases: ["update-agent-config"],
		short: "Update an agent config",
		long:
			"Update Agent config through the raw Console GraphQL API. Prefer `mosoo agent manifest apply`, which fetches remote state, preserves omitted fields, shows a field-level diff, and supports `--dry-run`.",
		example: "mosoo console agents update-config --file body.json -o json",
		hidden: true,
		notes: [
			"Raw API command. The product workflow command is `mosoo agent manifest apply`.",
			"Uses POST /graphql on the console default hostname (/api).",
			"Agent config updates are full-manifest updates: pull agent-manifest first, preserve unchanged environment/runtime/provider/tool fields, and submit the complete updated config.",
		],
	},
	accessibleAgentList: {
		aliases: ["list"],
		short: "List accessible Agents",
		long: "List Agents available to the signed-in user for one App.",
		example: "mosoo console agents list --app-id <app-id> -o json",
		examples: [
			{
				summary: "List Agents available to the signed-in user for one App and capture an Agent ID.",
				command: "mosoo console agents accessible-agent-list --app-id <app-id> -o json",
				body_shape: {
					appId: "<app-id>",
				},
				output_hints: {
					list_path: "data.accessibleAgentList",
					id_path: "data.accessibleAgentList[0].id",
				},
				follow_up_commands: [
					"mosoo console agents agent --app-id <app-id> --agent-id <agent-id> -o json",
				],
			},
		],
	},
	appList: {
		aliases: ["list"],
		short: "List Apps",
		long: "List Apps for an Organization. Prefer control-plane-overview for the current-user CLI overview path.",
		example: "mosoo console apps list --organization-id <organization-id> -o json",
	},
	appOverview: {
		short: "Show one App overview",
		long: "Show one App's console overview, including limited Agent and provider credential metadata.",
		example: "mosoo console apps app-overview --app-id <app-id> -o json",
	},
	controlPlaneOverview: {
		aliases: ["overview"],
		shortcuts: [{ use: "ls" }],
		short: "Show control-plane overview",
		long: "Show the current user's control-plane overview for generated CLI list flows. This is the main ls/overview path.",
		example: "mosoo console apps overview --app-limit 20 --agent-limit 20 --credential-limit 20 -o json",
		notes: [
			"Use this before lower-level app-list, accessible-agent-list, or vendor-credential-list when you need a CLI overview.",
		],
	},
	createApp: {
		example:
			'mosoo console apps create-app --input-organization-id <organization-id> --input-name "CLI Example App" -o json',
		examples: [
			{
				summary: "Create a minimal App and capture its id for follow-up commands.",
				command:
					'mosoo console apps create-app --input-organization-id <organization-id> --input-name "CLI Example App" -o json',
				body_shape: {
					input: {
						organizationId: "<organization-id>",
						name: "CLI Example App",
					},
				},
				output_hints: {
					id_path: "data.createApp.id",
				},
				follow_up_commands: ["mosoo console apps app-overview --app-id <app-id> -o json"],
			},
		],
	},
	createAgent: {
		aliases: ["create"],
		shortcuts: [{ use: "create-agent" }],
		short: "Create an Agent",
		long: "Create an Agent draft from a structured input. Publish it with publish-agent after provider credentials are configured.",
		example: [
			"cat > agent-create.json <<'JSON'",
			"{",
			"  \"input\": {",
			"    \"appId\": \"<app-id>\",",
			"    \"name\": \"Research Agent\",",
			"    \"kind\": \"<pet-or-cattle>\",",
			"    \"runtimeId\": \"<runtime-id>\",",
			"    \"provider\": \"<provider>\",",
			"    \"model\": \"<model>\",",
			"    \"prompt\": \"You research concise answers with citations.\",",
			"    \"skillIds\": []",
			"  }",
			"}",
			"JSON",
			"mosoo console agents create --file agent-create.json -o json",
		].join("\n"),
		examples: [
			{
				summary: "Create an Agent from a JSON body file",
				command: "mosoo console agents create-agent --file agent-create.json -o json",
				body_shape: {
					input: {
						appId: "<app-id>",
						name: "Research Agent",
						kind: "pet",
						runtimeId: "<runtime-id>",
						provider: "<provider>",
						model: "<model>",
						prompt: "You research concise answers with citations.",
						skillIds: [],
					},
				},
				output_hints: {
					id_path: "data.createAgent.id",
				},
				follow_up_commands: [
					"mosoo console agents agent --app-id <app-id> --agent-id <id> -o json",
					"mosoo console agents publish-agent --input-app-id <app-id> --input-agent-id <id> -o json",
				],
			},
		],
	},
	createVendorCredential: {
		aliases: ["create"],
		shortcuts: [{ use: "add-key" }],
		short: "Add a provider key",
		long: "Create a provider credential for an App while keeping API keys out of shell history.",
		example: [
			"mosoo console credentials create-vendor-credential \\",
			"  --input-app-id <app-id> \\",
			"  --input-vendor-id openai \\",
			"  --input-name \"OpenAI\" \\",
			"  --input-api-key-env OPENAI_API_KEY \\",
			"  -o json",
		].join("\n"),
		examples: [
			{
				summary: "Create an OpenAI provider credential from an environment variable",
				command: [
					"mosoo console credentials create-vendor-credential \\",
					"  --input-app-id <app-id> \\",
					"  --input-vendor-id openai \\",
					"  --input-name \"OpenAI\" \\",
					"  --input-api-key-env OPENAI_API_KEY \\",
					"  -o json",
				].join("\n"),
				body_shape: {
					input: {
						appId: "<app-id>",
						vendorId: "openai",
						name: "OpenAI",
						apiKey: "$OPENAI_API_KEY",
					},
				},
				output_hints: {
					id_path: "data.createVendorCredential.id",
				},
				follow_up_commands: [
					"mosoo console credentials vendor-credential-list --app-id <app-id> -o json",
				],
			},
		],
		notes: [
			"For preset providers such as openai or anthropic, omit input.models unless Mosoo asks for explicit model configuration.",
			"Use --input-api-key-env, --input-api-key-file, or --input-api-key-stdin so provider keys do not appear in shell history.",
			"Current Lathe required variable flags must be present; safe input modes satisfy the required input.apiKey variable.",
		],
	},
	publishAgent: {
		aliases: ["publish"],
		short: "Publish an Agent",
		long: "Publish an Agent after its draft configuration and provider credentials are ready.",
		example: "mosoo console agents publish --input-app-id <app-id> --input-agent-id <agent-id> -o json",
	},
	startAgentRun: {
		aliases: ["run"],
		shortcuts: [{ use: "run" }],
		short: "Start an Agent run",
		long: "Create or continue a Thread, queue one prompt Run, and return event-surface metadata for polling.",
		example: [
			"mosoo console sessions run \\",
			"  --input-app-id <app-id> \\",
			"  --input-agent-id <agent-id> \\",
			"  --input-prompt \"Summarize this repository\" \\",
			"  -o json",
		].join("\n"),
		notes: [
			"This is the generated main path for mosoo run. Use the returned appId/sessionId with thread-session-process-events to poll output.",
		],
	},
	testVendorCredential: {
		aliases: ["test"],
		short: "Test a provider key",
		long: "Test provider credential material before or after saving it.",
		example: [
			"mosoo console credentials test-vendor-credential \\",
			"  --input-app-id <app-id> \\",
			"  --input-vendor-id openai \\",
			"  --input-model-id gpt-4o-mini \\",
			"  --input-api-key-env OPENAI_API_KEY \\",
			"  -o json",
		].join("\n"),
		examples: [
			{
				summary: "Smoke test an OpenAI provider key from the environment before saving it.",
				command: [
					"mosoo console credentials test-vendor-credential \\",
					"  --input-app-id <app-id> \\",
					"  --input-vendor-id openai \\",
					"  --input-model-id gpt-4o-mini \\",
					"  --input-api-key-env OPENAI_API_KEY \\",
					"  -o json",
				].join("\n"),
				body_shape: {
					input: {
						appId: "<app-id>",
						vendorId: "openai",
						modelId: "gpt-4o-mini",
						apiKey: "<read from OPENAI_API_KEY>",
					},
				},
				follow_up_commands: [
					'mosoo console credentials create-vendor-credential --input-app-id <app-id> --input-vendor-id openai --input-name "OpenAI" --input-api-key-env OPENAI_API_KEY -o json',
				],
			},
		],
		notes: [
			"Prefer --input-api-key-env, --input-api-key-file, or --input-api-key-stdin so provider secrets do not appear in shell history.",
			"Current Lathe required variable flags must be present; --set and --set-str can supplement body fields but do not replace required flags.",
		],
	},
	threadSessionProcessEvents: {
		aliases: ["events", "process-events"],
		short: "Poll Thread session events",
		long: "Read process events for a Thread session. Use this after start-agent-run when streamUrl is null.",
		example: "mosoo console sessions events --app-id <app-id> --session-id <session-id> -o json",
	},
	vendorCredentialList: {
		aliases: ["list"],
		short: "List provider keys",
		long: "List provider credentials for an App. Prefer control-plane-overview for summary counts and status.",
		example: "mosoo console credentials list --app-id <app-id> -o json",
	},
};

function camelToKebab(value: string): string {
	return value.replace(/([a-z0-9])([A-Z])/g, "$1-$2").replace(/_/g, "-").toLowerCase();
}

function splitCamel(value: string): string[] {
	return value
		.replace(/([a-z0-9])([A-Z])/g, "$1 $2")
		.replace(/_/g, " ")
		.toLowerCase()
		.split(/\s+/)
		.filter(Boolean);
}

function joinWords(words: string[]): string {
	if (words.length === 0) {
		return "";
	}
	return words.join(" ").replace(/\bagent api\b/g, "agent API");
}

function pluralizeWords(words: string[]): string[] {
	if (words.length === 0) {
		return words;
	}
	const last = words[words.length - 1]!;
	if (last.endsWith("s") || last.endsWith("ss") || last === "status") {
		return words;
	}
	return [...words.slice(0, -1), `${last}s`];
}

function titlePhrase(words: string[]): string {
	if (words.length === 0) {
		return "resources";
	}
	return joinWords(words);
}

function humanizeGraphQLField(field: string): string {
	if (field === "viewer") {
		return "Show the signed-in viewer";
	}

	const verbs: Record<string, string> = {
		add: "Add",
		archive: "Archive",
		auto: "Auto",
		connect: "Connect",
		create: "Create",
		delete: "Delete",
		ensure: "Ensure",
		execute: "Execute",
		export: "Export",
		import: "Import",
		poll: "Poll",
		prewarm: "Prewarm",
		publish: "Publish",
		recreate: "Recreate",
		remove: "Remove",
		rename: "Rename",
		reset: "Reset",
		restart: "Restart",
		revoke: "Revoke",
		set: "Set",
		start: "Start",
		test: "Test",
		unarchive: "Unarchive",
		unpublish: "Unpublish",
		update: "Update",
	};

	if (field.endsWith("List")) {
		return `List ${titlePhrase(pluralizeWords(splitCamel(field.slice(0, -4))))}`;
	}

	for (const [prefix, label] of Object.entries(verbs)) {
		if (!field.startsWith(prefix) || field.length === prefix.length) {
			continue;
		}
		const rest = field.slice(prefix.length);
		if (noArticleVerbs.has(prefix)) {
			return `${label} ${titlePhrase(splitCamel(rest))}`;
		}
		const article = /^[aeiou]/i.test(rest) ? "an" : "a";
		return `${label} ${article} ${titlePhrase(splitCamel(rest))}`;
	}

	return titlePhrase(splitCamel(field)).replace(/^./, (c) => c.toUpperCase());
}

function requiredFlags(field: string): string[] {
	if (field === "viewer") {
		return [];
	}
	if (field === "appList" || field.endsWith("AppList")) {
		return ["organization-id"];
	}
	if (field.includes("appId") || field.endsWith("App") || field.includes("Agent")) {
		return ["app-id"];
	}
	return [];
}

function consoleExample(group: string, use: string, flags: string[]): string {
	const prefix = `mosoo console ${group.toLowerCase()} ${use}`;
	if (flags.length === 0) {
		return prefix;
	}
	return `${prefix} \\\n  ${flags.map((flag) => `--${flag} <value>`).join(" \\\n  ")}`;
}

function buildConsoleOverlay(): Record<string, OverlayCommand> {
	const commands: Record<string, OverlayCommand> = {};
	for (const { group, field } of collectConsoleGraphQLOperations()) {
		const use = camelToKebab(field);
		const flags = requiredFlags(field);
		const base: OverlayCommand = {
			short: humanizeGraphQLField(field),
			long: `${humanizeGraphQLField(field)} via the Mosoo Console GraphQL API (${group} surface). Requires a personal access token logged in to the /api host.`,
			...(exampleFields.has(field) ? { example: consoleExample(group, use, flags) } : {}),
			notes: ["Uses POST /graphql on the console default hostname (/api)."],
			known_errors: [{ status: 401, cause: "Missing, invalid, or revoked personal access token." }],
			...consoleCommandOverrides[use],
		};
		const override = consoleCommandOverrides[field];
		commands[use] = {
			...base,
			...override,
			...(base.notes || override?.notes
				? { notes: [...(base.notes ?? []), ...(override?.notes ?? [])] }
				: {}),
			...(base.known_errors || override?.known_errors
				? { known_errors: [...(base.known_errors ?? []), ...(override?.known_errors ?? [])] }
				: {}),
		};
	}
	return commands;
}

const thread401 = [{ status: 401, cause: "Invalid personal access token." }];

function buildThreadsOverlay(): Record<string, OverlayCommand> {
	return {
		"list-events": {
			short: "List thread events",
			long: "Return recent events for a thread on the Public Thread API.",
			example: "mosoo public-thread-api events list-events --thread-id <thread-id>",
			known_errors: [...thread401, { status: 404, cause: "Thread not found for this caller." }],
		},
		send: {
			short: "Send events to a thread",
			long: "Send user messages, permission decisions, or interrupts to a thread session.",
			example: "mosoo public-thread-api events send --thread-id <thread-id> --file events.json",
			examples: [
				{
					summary: "Send a user message event to an existing thread.",
					command: "mosoo public-thread-api events send --thread-id <thread-id> --file events.json -o json",
					body_shape: {
						events: [
							{
								type: "user_message",
								text: "Continue the task with this follow-up.",
								clientRequestId: "cli-send-001",
							},
						],
					},
					output_hints: {
						list_path: "events",
					},
					follow_up_commands: [
						"mosoo public-thread-api events list-events --thread-id <thread-id> -o json",
					],
				},
			],
			known_errors: [
				...thread401,
				{ status: 409, cause: "Idempotency key reused while the original request is still processing." },
			],
		},
		stream: {
			short: "Stream thread events (SSE)",
			long: "Open a Server-Sent Events stream of thread events. Use -o raw for the event stream.",
			example: "mosoo public-thread-api events stream --thread-id <thread-id> -o raw",
		},
		add: {
			short: "Attach a file to a thread",
			long: "Associate an uploaded file with a thread.",
			example: "mosoo public-thread-api files add --thread-id <thread-id> --set fileId=<file-id>",
		},
		"list-files": {
			short: "List thread files",
			long: "List files attached to a thread.",
			example: "mosoo public-thread-api files list-files --thread-id <thread-id>",
		},
		remove: {
			short: "Remove a thread file",
			long: "Detach a file from a thread.",
			example: "mosoo public-thread-api files remove --thread-id <thread-id> --file-id <file-id>",
		},
		archive: {
			short: "Archive a thread",
			long: "Mark a thread as archived.",
			example: "mosoo public-thread-api threads archive --thread-id <thread-id>",
		},
		create: {
			short: "Create a thread for an agent",
			long: "Create a new thread against an agent API endpoint.",
			example: "mosoo public-thread-api threads create --agent-id <agent-id> --file body.json",
			examples: [
				{
					summary: "Create a Thread with an initial user message and capture the Thread ID.",
					command: "mosoo public-thread-api threads create --agent-id <agent-id> --file thread-create.json -o json",
					body_shape: {
						client_external_ref: "demo-thread-001",
						input: {
							type: "user.message",
							content: [
								{
									type: "text",
									text: "Say hello from the API.",
								},
							],
						},
					},
					output_hints: {
						id_path: "thread.id",
					},
					follow_up_commands: [
						"mosoo public-thread-api threads retrieve --thread-id <thread-id> -o json",
						"mosoo public-thread-api events list-events --thread-id <thread-id> -o json",
					],
				},
			],
			known_errors: [{ status: 404, cause: "Agent not found or not accessible to this token." }],
		},
		delete: {
			short: "Delete a thread",
			long: "Permanently delete a thread owned by this caller.",
			example: "mosoo public-thread-api threads delete --thread-id <thread-id>",
		},
		"list-for-agent": {
			short: "List threads for an agent",
			long: "List threads created against an agent API endpoint.",
			example: "mosoo public-thread-api threads list-for-agent --agent-id <agent-id>",
		},
		retrieve: {
			short: "Retrieve a thread",
			long: "Fetch thread metadata and status.",
			example: "mosoo public-thread-api threads retrieve --thread-id <thread-id>",
		},
		unarchive: {
			short: "Unarchive a thread",
			long: "Restore an archived thread to the active list.",
			example: "mosoo public-thread-api threads unarchive --thread-id <thread-id>",
		},
	};
}

function buildConsolerestOverlay(): Record<string, OverlayCommand> {
	return {
		create: {
			example: 'mosoo console-rest "access tokens" create --set label="ci"',
			notes: ["Creates a personal access token. The secret value is returned once."],
		},
		list: {
			example: 'mosoo console-rest "access tokens" list',
		},
		revoke: {
			example: 'mosoo console-rest "access tokens" revoke --token-id <token-id>',
		},
	};
}

function yamlQuote(value: string): string {
	if (/^[a-zA-Z0-9_./:-]+$/.test(value) && !value.includes("#")) {
		return value;
	}
	return JSON.stringify(value);
}

function renderYamlValue(lines: string[], key: string, value: unknown, indent: number): void {
	const pad = " ".repeat(indent);
	if (typeof value === "string") {
		lines.push(`${pad}${key}: ${yamlQuote(value)}`);
		return;
	}
	if (typeof value === "number" || typeof value === "boolean") {
		lines.push(`${pad}${key}: ${String(value)}`);
		return;
	}
	if (Array.isArray(value)) {
		if (value.length === 0) {
			lines.push(`${pad}${key}: []`);
			return;
		}
		lines.push(`${pad}${key}:`);
		for (const item of value) {
			if (typeof item === "string") {
				lines.push(`${pad}  - ${yamlQuote(item)}`);
				continue;
			}
			lines.push(`${pad}  -`);
			renderYamlRecord(lines, item as Record<string, unknown>, indent + 4);
		}
		return;
	}
	if (value && typeof value === "object") {
		lines.push(`${pad}${key}:`);
		renderYamlRecord(lines, value as Record<string, unknown>, indent + 2);
	}
}

function renderYamlRecord(lines: string[], value: Record<string, unknown>, indent: number): void {
	for (const [key, child] of Object.entries(value)) {
		renderYamlValue(lines, key, child, indent);
	}
}

function renderExamples(lines: string[], examples: OverlayExample[]): void {
	lines.push("    examples:");
	for (const example of examples) {
		lines.push(`      - summary: ${yamlQuote(example.summary)}`);
		lines.push(`        command: ${yamlQuote(example.command)}`);
		if (example.body_shape) {
			renderYamlValue(lines, "body_shape", example.body_shape, 8);
		}
		if (example.output_hints) {
			renderYamlValue(lines, "output_hints", example.output_hints, 8);
		}
		if (example.follow_up_commands?.length) {
			renderYamlValue(lines, "follow_up_commands", example.follow_up_commands, 8);
		}
	}
}

function renderOverlay(commands: Record<string, OverlayCommand>): string {
	const lines = ["commands:"];
	for (const [use, command] of Object.entries(commands).sort(([a], [b]) => a.localeCompare(b))) {
		lines.push(`  ${use}:`);
		if (command.use) {
			lines.push(`    use: ${yamlQuote(command.use)}`);
		}
		if (command.aliases?.length) {
			lines.push(`    aliases: [${command.aliases.map(yamlQuote).join(", ")}]`);
		}
		if (command.shortcuts?.length) {
			lines.push("    shortcuts:");
			for (const shortcut of command.shortcuts) {
				lines.push(`      - use: ${yamlQuote(shortcut.use)}`);
				if (shortcut.params && Object.keys(shortcut.params).length > 0) {
					renderYamlValue(lines, "params", shortcut.params, 8);
				}
			}
		}
		if (command.short) {
			lines.push(`    short: ${yamlQuote(command.short)}`);
		}
		if (command.long) {
			lines.push(`    long: ${yamlQuote(command.long)}`);
		}
		if (command.example) {
			lines.push("    example: |");
			for (const line of command.example.split("\n")) {
				lines.push(`      ${line}`);
			}
		}
		if (command.examples?.length) {
			renderExamples(lines, command.examples);
		}
		if (command.hidden !== undefined) {
			lines.push(`    hidden: ${command.hidden ? "true" : "false"}`);
		}
		if (command.ignore !== undefined) {
			lines.push(`    ignore: ${command.ignore ? "true" : "false"}`);
		}
		if (command.notes?.length) {
			lines.push("    notes:");
			for (const note of command.notes) {
				lines.push(`      - ${yamlQuote(note)}`);
			}
		}
		if (command.prerequisites?.length) {
			lines.push("    prerequisites:");
			for (const prerequisite of command.prerequisites) {
				lines.push(`      - ${yamlQuote(prerequisite)}`);
			}
		}
		if (command.known_errors?.length) {
			lines.push("    known_errors:");
			for (const err of command.known_errors) {
				lines.push(`      - status: ${err.status}`);
				lines.push(`        cause: ${yamlQuote(err.cause)}`);
			}
		}
	}
	return `${lines.join("\n")}\n`;
}

const overlays = {
	console: buildConsoleOverlay(),
	threads: buildThreadsOverlay(),
	consolerest: buildConsolerestOverlay(),
};

await mkdir(overlayDir, { recursive: true });
for (const [name, commands] of Object.entries(overlays)) {
	await writeFile(resolve(overlayDir, `${name}.yaml`), renderOverlay(commands), "utf8");
}
console.log(
	`wrote overlays/*.yaml (console=${Object.keys(overlays.console).length}, threads=12, consolerest=${Object.keys(overlays.consolerest).length})`,
);
