import * as fs from "node:fs";

import * as yaml from "yaml";

const data = yaml.parse(fs.readFileSync("pkg/vars/common/subject_platforms.yml", "utf-8"));

const platforms = Object.fromEntries(Object.entries(data.platforms).filter(([key, value]) => /\d/.test(key)));

platforms[3] ??= {
  0: {
    id: 0,
    type: "",
    type_cn: "",
    alias: "",
    wiki_tpl: "",
    order: 0,
  },
};

fs.writeFileSync("pkg/vars/platform.go.json", JSON.stringify(platforms, null, 2));

fs.writeFileSync(
  "pkg/vars/staffs.go.json",
  JSON.stringify(yaml.parse(fs.readFileSync("pkg/vars/common/subject_staffs.yml", "utf-8")), null, 2),
);

fs.writeFileSync(
  "pkg/vars/relations.go.json",
  JSON.stringify(yaml.parse(fs.readFileSync("pkg/vars/common/subject_relations.yml", "utf-8")), null, 2),
);
