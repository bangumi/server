const fs = require("fs/promises");
const path = require("path");

const root = path.dirname(__dirname);

async function main() {
  await fs.rm(path.join(root, "./internal/web/frontend/static/"), { recursive: true, force: true });
  await fs.mkdir(path.join(root, "./internal/web/frontend/static/bootstrap/"), { recursive: true });

  for (const file of [
    "css/bootstrap.min.css",
    "css/bootstrap.min.css.map",
    "js/bootstrap.min.js",
    "js/bootstrap.min.js.map",
  ]) {
    const absDest = path.join(root, "./internal/web/frontend/static/bootstrap/", file);

    await fs.mkdir(path.dirname(absDest), { recursive: true });

    await fs.copyFile(path.join(root, "./node_modules/bootstrap/dist/", file), absDest);
  }

  await fs.copyFile(
    path.join(root, "./node_modules/jquery/dist/jquery.slim.min.js"),
    path.join(root, "./internal/web/frontend/static/jquery.slim.min.js")
  );
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
