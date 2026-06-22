import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { createGraphQLSchema } from "../.cache/mosoo/apps/api/src/adapters/graphql/create-graphql-schema.ts";
import { printSchema } from "../.cache/mosoo/node_modules/graphql/index.js";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "..");
const outputPath = resolve(repositoryRoot, ".cache/mosoo/docs/graphql/console.graphql");

const schema = createGraphQLSchema();
const queryType = schema.getQueryType();
const mutationType = schema.getMutationType();
if (queryType === null || mutationType === null) {
	throw new Error("Console GraphQL schema must define Query and Mutation root types.");
}

const queryFields = queryType.getFields();
const mutationFields = mutationType.getFields();
const sdl = `${printSchema(schema).trimEnd()}\n`;
await mkdir(dirname(outputPath), { recursive: true });
await writeFile(outputPath, sdl, "utf8");
console.log(
	`wrote ${outputPath} (${Object.keys(queryFields).length} queries, ${Object.keys(mutationFields).length} mutations)`,
);
