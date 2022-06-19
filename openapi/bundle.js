const fs = require("fs/promises");

const $RefParser = require("@apidevtools/json-schema-ref-parser");
const yaml = require("js-yaml");
const lodash = require("lodash");

async function main() {
  let schema = await $RefParser.bundle(process.argv[2]);

  schema = lodash.omit(schema, "x-parameters");

  if (process.argv[3]) {
    await fs.writeFile(process.argv[3], yaml.dump(schema, {noRefs: true}));
  } else {
    console.log(schema);
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
