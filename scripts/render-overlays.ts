import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { collectConsoleGraphQLOperations } from "./console-graphql-sources.ts";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const overlayDir = resolve(dirname(fileURLToPath(import.meta.url)), "..", "overlays");

type OverlayCommand = {
	use?: string;
	aliases?: string[];
	short?: string;
	long?: string;
	example?: string;
	hidden?: boolean;
	ignore?: boolean;
	notes?: string[];
	prerequisites?: string[];
	known_errors?: { status: number; cause: string }[];
};

const exampleFields = new Set(["viewer", "appList", "createApp", "createAgent", "agentSessionList", "sessionList"]);
const noArticleVerbs = new Set(["set", "poll", "start", "test"]);

const consoleCommandOverrides: Record<string, Partial<OverlayCommand>> = {
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
		commands[use] = {
			short: humanizeGraphQLField(field),
			long: `${humanizeGraphQLField(field)} via the Mosoo Console GraphQL API (${group} surface). Requires a personal access token logged in to the /api host.`,
			...(exampleFields.has(field) ? { example: consoleExample(group, use, flags) } : {}),
			notes: ["Uses POST /graphql on the console default hostname (/api)."],
			known_errors: [{ status: 401, cause: "Missing, invalid, or revoked personal access token." }],
			...consoleCommandOverrides[use],
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
