const $RefParser = require("@apidevtools/json-schema-ref-parser");
const fs = require("fs/promises");
const yaml = require("js-yaml")

async function main() {
  let schema = await $RefParser.bundle(process.argv[2]);
  if (process.argv[3]) {
    await fs.writeFile(process.argv[3], yaml.dump(schema));
  } else {
    console.log(schema);
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
