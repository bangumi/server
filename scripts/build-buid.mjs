import * as path from "node:path";
import * as fs from "node:fs";

import * as yaml from "yaml";

const data = yaml.parse(fs.readFileSync("pkg/vars/common/subject_platforms.yml", "utf-8"));

fs.writeFileSync(
  "pkg/vars/platform.go.json",
  JSON.stringify(Object.fromEntries(Object.entries(data.platforms).filter(([key, value]) => /\d/.test(key))), null, 2),
);
