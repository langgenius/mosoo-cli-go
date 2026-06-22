import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { excludedMutations, listFields, moduleGroups } from "./console-graphql-sources.ts";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "..");
const mosooDir = resolve(repositoryRoot, ".cache/mosoo");
const repoURL = `file://${mosooDir}`;
const pinnedTag = "local-snapshot";
const hostBase = (process.env.MOSOO_HOST_BASE ?? "http://127.0.0.1:8787").replace(/\/$/, "");
const consoleDefaultHostname = `${hostBase}/api`;
const threadsDefaultHostname = `${hostBase}/api/v1`;

const queries: string[] = [];
const mutations: string[] = [];
const groups: { match: string[]; group: string }[] = [];

for (const { group, spec } of moduleGroups) {
	const moduleQueries = listFields(spec.queryFields);
	const moduleMutations = listFields(spec.mutationFields).filter((name) => !excludedMutations.has(name));
	if (moduleQueries.length === 0 && moduleMutations.length === 0) {
		continue;
	}
	queries.push(...moduleQueries);
	mutations.push(...moduleMutations);
	groups.push({ group, match: [...moduleQueries, ...moduleMutations] });
}

queries.sort();
mutations.sort();

function yamlList(items: string[], indent: string): string {
	if (items.length === 0) {
		return " []";
	}
	return `\n${items.map((item) => `${indent}- ${item}`).join("\n")}`;
}

function yamlGroupRules(indent: string): string {
	return groups
		.map((rule) => `${indent}- match:${yamlList(rule.match, indent + "    ")}\n${indent}  group: ${rule.group}`)
		.join("\n");
}

const sourcesYaml = `sources:
  threads:
    display_name: public-thread-api
    default_hostname: ${threadsDefaultHostname}
    repo_url: ${repoURL}
    pinned_tag: ${pinnedTag}
    backend: openapi3
    openapi3:
      files:
        - docs/openapi/public-thread-api.openapi.json
  console:
    display_name: console
    default_hostname: ${consoleDefaultHostname}
    repo_url: ${repoURL}
    pinned_tag: ${pinnedTag}
    backend: graphql
    graphql:
      schema: docs/graphql/console.graphql
      expose:
        queries:${yamlList(queries, "          ")}
        mutations:${yamlList(mutations, "          ")}
      groups:
${yamlGroupRules("        ")}
      selection:
        max_depth: 3
  consolerest:
    display_name: console-rest
    default_hostname: ${consoleDefaultHostname}
    repo_url: ${repoURL}
    pinned_tag: ${pinnedTag}
    backend: openapi3
    openapi3:
      files:
        - docs/openapi/console-rest.openapi.json
`;

const syncStates = [
	["threads", "docs/openapi"],
	["console", "docs/graphql"],
	["consolerest", "docs/openapi"],
] as const;

await mkdir(resolve(repositoryRoot, "specs"), { recursive: true });
for (const [source, specDir] of syncStates) {
	await mkdir(resolve(repositoryRoot, `.cache/specs-sync/${source}/${specDir}`), { recursive: true });
	await writeFile(
		resolve(repositoryRoot, `.cache/specs-sync/${source}/sync-state.yaml`),
		`source: ${source}\nbackend: ${specDir.includes("graphql") ? "graphql" : "openapi3"}\nsynced_from: ${pinnedTag}\nresolved_sha: ${pinnedTag}\n`,
		"utf8",
	);
}
await writeFile(resolve(repositoryRoot, "specs/sources.yaml"), sourcesYaml, "utf8");
console.log(
	`wrote specs/sources.yaml (${queries.length} queries, ${mutations.length} mutations, ${groups.length} groups, 3 sources)`,
);
