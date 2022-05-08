const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const VERSION = process.env.npm_package_version;
if (!VERSION) {
  console.error("version is undefined");
  process.exit(1);
}

const goFile = path.join(__dirname, "..", "config/version.go");

fs.writeFileSync(
  goFile,
  `// Code generated. DO NOT EDIT.\n\npackage config\n\nconst Version = "v${VERSION}"`
);

spawnSync("go", ["fmt", goFile]);
spawnSync("git", ["add", goFile]);
