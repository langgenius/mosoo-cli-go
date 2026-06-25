import { cp, mkdir, readdir, readFile, rm, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "..");
const skillRoot = resolve(repositoryRoot, "publish/skills/mosoo");
const generatedSkillRoot = resolve(repositoryRoot, ".cache/lathe-skill/mosoo");
const generatedReferencesRoot = resolve(generatedSkillRoot, "references");
const outputReferenceRoot = resolve(skillRoot, "references/cli");
const outputCliGuide = resolve(skillRoot, "references/cli.md");

async function normalizeMarkdown(path: string): Promise<void> {
	const text = await readFile(path, "utf8");
	const normalized = text.replace(/[ \t]+\n/g, "\n").replace(/\n{2,}$/g, "\n");
	if (normalized !== text) {
		await writeFile(path, normalized, "utf8");
	}
}

async function normalizeMarkdownTree(path: string): Promise<void> {
	const entries = await readdir(path, { withFileTypes: true });
	for (const entry of entries) {
		const entryPath = resolve(path, entry.name);
		if (entry.isDirectory()) {
			await normalizeMarkdownTree(entryPath);
			continue;
		}
		if (entry.isFile() && entry.name.endsWith(".md")) {
			await normalizeMarkdown(entryPath);
		}
	}
}

await mkdir(outputReferenceRoot, { recursive: true });

await rm(resolve(outputReferenceRoot, "catalog.md"), { force: true });
await rm(resolve(outputReferenceRoot, "modules"), { force: true, recursive: true });
await rm(resolve(outputReferenceRoot, "mosoo"), { force: true, recursive: true });
await cp(resolve(generatedReferencesRoot, "cli.md"), outputCliGuide);
await cp(resolve(generatedReferencesRoot, "catalog.md"), resolve(outputReferenceRoot, "catalog.md"));
await cp(resolve(generatedReferencesRoot, "modules"), resolve(outputReferenceRoot, "modules"), { recursive: true });
await normalizeMarkdown(outputCliGuide);
await normalizeMarkdownTree(outputReferenceRoot);

console.log("wrote publish/skills/mosoo/references/cli/");
