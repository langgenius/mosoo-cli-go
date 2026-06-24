import { cp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "..");
const skillRoot = resolve(repositoryRoot, "publish/skills/mosoo");
const generatedSkillRoot = resolve(skillRoot, "references/cli/mosoo");
const generatedReferencesRoot = resolve(generatedSkillRoot, "references");
const outputReferenceFile = resolve(skillRoot, "references/cli.md");
const outputReferenceRoot = resolve(skillRoot, "references/cli");

function stripFrontmatter(markdown: string): string {
	if (!markdown.startsWith("---\n")) {
		return markdown;
	}
	const end = markdown.indexOf("\n---\n", 4);
	if (end === -1) {
		throw new Error("generated CLI Skill is missing closing frontmatter");
	}
	return markdown.slice(end + "\n---\n".length);
}

function renderCliReference(generatedSkillMarkdown: string): string {
	const body = stripFrontmatter(generatedSkillMarkdown)
		.replace(/^# mosoo CLI\s*/m, "")
		.replace(
			"Use this skill when a user asks you to operate `mosoo`, inspect its API commands, or find the right generated command for an API task.",
			"Use this reference when a user asks you to operate `mosoo`, inspect its API commands, or find the right generated command for an API task.",
		)
		.replaceAll("`references/catalog.md`", "`references/cli/catalog.md`")
		.replaceAll("`references/modules/", "`references/cli/modules/");

	return `# CLI Reference

Generated from Lathe's mosoo CLI Skill output during \`make build\`.

## Runtime State

Run:

\`\`\`sh
mosoo doctor --json
\`\`\`

Use the result to decide whether the current task targets local Mosoo runtime or
Mosoo cloud runtime before running API commands.

## Command Selection

Use generated CLI commands for Mosoo resource operations, and use
\`references/api.md\` for application code that calls an already published Agent.
Do not invent a wrapper command when the generated catalog already exposes the
operation.

For a new App, Agent creation, publishing, credential setup, or Console/API
inspection, search the generated catalog first. For app environment files only,
derive \`MOSOO_API_BASE\`, \`MOSOO_AGENT_ID\`, and \`MOSOO_API_TOKEN\` from the
published Agent/API contract instead of creating new resources.

${body.trim()}
`;
}

const generatedSkillMarkdown = await readFile(resolve(generatedSkillRoot, "SKILL.md"), "utf8");

await mkdir(outputReferenceRoot, { recursive: true });
await writeFile(outputReferenceFile, renderCliReference(generatedSkillMarkdown), "utf8");

await rm(resolve(outputReferenceRoot, "catalog.md"), { force: true });
await rm(resolve(outputReferenceRoot, "modules"), { force: true, recursive: true });
await cp(resolve(generatedReferencesRoot, "catalog.md"), resolve(outputReferenceRoot, "catalog.md"));
await cp(resolve(generatedReferencesRoot, "modules"), resolve(outputReferenceRoot, "modules"), { recursive: true });
await rm(generatedSkillRoot, { force: true, recursive: true });

console.log("wrote publish/skills/mosoo/references/cli.md and references/cli/");
