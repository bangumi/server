// bundle openapi files to ./dist/
const fs = require("fs/promises");
const path = require("path");

const $RefParser = require("@apidevtools/json-schema-ref-parser");
const yaml = require("js-yaml");
const lodash = require("lodash");

async function main() {
  await fs.mkdir(path.join(__dirname, "..", "dist"), { recursive: true });

  for (const filePath of ["v0.yaml", "private.yaml"]) {
    console.log(`build openapi ${filePath} => dist/${filePath}`);

    const input = path.join(__dirname, filePath);
    let schema = await $RefParser.bundle(input);

    schema = lodash.omit(schema, "x-parameters");

    if (input.endsWith("private.yaml")) {
      schema["servers"] = [
        { url: "https://next.bgm.tv", description: "Production server" },
        { url: "https://dev.bgm38.com/", description: "开发用服务器" },
      ];
    }

    const out = path.join(__dirname, "..", "dist", filePath);
    await fs.writeFile(out, yaml.dump(schema, { noRefs: true }));
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
