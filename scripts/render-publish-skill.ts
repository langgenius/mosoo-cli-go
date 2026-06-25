import { cp, mkdir, rm } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "..");
const skillRoot = resolve(repositoryRoot, "publish/skills/mosoo");
const generatedSkillRoot = resolve(skillRoot, "references/cli/mosoo");
const generatedReferencesRoot = resolve(generatedSkillRoot, "references");
const outputReferenceRoot = resolve(skillRoot, "references/cli");
const outputCliGuide = resolve(skillRoot, "references/cli.md");

await mkdir(outputReferenceRoot, { recursive: true });

await rm(resolve(outputReferenceRoot, "catalog.md"), { force: true });
await rm(resolve(outputReferenceRoot, "modules"), { force: true, recursive: true });
await cp(resolve(generatedReferencesRoot, "cli.md"), outputCliGuide);
await cp(resolve(generatedReferencesRoot, "catalog.md"), resolve(outputReferenceRoot, "catalog.md"));
await cp(resolve(generatedReferencesRoot, "modules"), resolve(outputReferenceRoot, "modules"), { recursive: true });
await rm(generatedSkillRoot, { force: true, recursive: true });

console.log("wrote publish/skills/mosoo/references/cli/");
