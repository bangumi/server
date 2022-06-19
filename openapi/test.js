/*
 * SPDX-License-Identifier: AGPL-3.0-only
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>
 */
const path = require("path");
const fs = require("fs");

const lodash = require("lodash");
const $RefParser = require("@apidevtools/json-schema-ref-parser");
const validator = require("oas-validator");
const yaml = require("js-yaml");
const colors = require("colors/safe");

async function main() {
  process.chdir(__dirname);
  for (const filePath of ["v0.yaml", "private.yaml"]) {
    console.log("try to bundle", filePath);

    // JSON deep copy to remove anchor
    const raw = JSON.parse(JSON.stringify(yaml.load(fs.readFileSync(path.join(__dirname, filePath), "utf-8"))));

    const openapi = await $RefParser.bundle(raw);
    try {
      await validator.validate(openapi, {
        lint: true,
        lintSkip: ["info-contact", "contact-properties", "tag-description"],
      });
    } catch (e) {
      if (!e.options.warnings.length) {
        throw e;
      }

      for (const {
        pointer,
        ruleName,
        rule: { description },
      } of e.options.warnings) {
        const path = dataPathToJSONPath(pointer);
        console.error(ruleName, colors.red(`${description}:`), path);
      }
      throw new Error(`${e.options.warnings.length} errors, failed to validate ${filePath}`);
    }
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});

function dataPathToJSONPath(s) {
  const ptr = s.replaceAll("/", ".").replaceAll("~1", "/");

  return (
    "$" +
    ptr
      .split(".")
      .slice(1)
      .map((x) => {
        if (Number.isInteger(parseFloat(x))) {
          return `[${x}]`;
        }

        if (["/", "'", '"', "-"].filter((c) => lodash.includes(x, c)).length !== 0) {
          return `[${JSON.stringify(x)}]`;
        }

        return `.` + x;
      })
      .join("")
  );
}
