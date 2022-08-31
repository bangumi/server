const fs = require("fs/promises");

const $RefParser = require("@apidevtools/json-schema-ref-parser");
const yaml = require("js-yaml");
const lodash = require("lodash");

async function main() {
  const input = process.argv[2];
  let schema = await $RefParser.bundle(input);

  schema = lodash.omit(schema, "x-parameters");

  if (input.endsWith("private.yaml")) {
    schema["servers"] = [
      { url: "https://next.bgm.tv", description: "Production server" },
      { url: "https://dev.bgm38.com/", description: "开发用服务器" },
    ];
  }

  if (process.argv[3]) {
    await fs.writeFile(process.argv[3], yaml.dump(schema, { noRefs: true }));
  } else {
    console.log(schema);
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
