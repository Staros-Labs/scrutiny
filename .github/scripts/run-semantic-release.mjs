import fs from "node:fs";
import process from "node:process";
import semanticRelease from "semantic-release";
const result = await semanticRelease({ cwd: process.cwd() });
const output = process.env.GITHUB_OUTPUT;
if (!output) {
  throw new Error("GITHUB_OUTPUT is required");
}
const published = result ? "true" : "false";
const version = result?.nextRelease?.version ?? "";
fs.appendFileSync(output, "new_release_published=" + published + "\n");
fs.appendFileSync(output, "new_release_version=" + version + "\n");
