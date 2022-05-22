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

const $RefParser = require("@apidevtools/json-schema-ref-parser");

async function main() {
  for (const filePath of ["v0.yaml", "private.yaml"]) {
    console.log("try to bundle", filePath);
    await $RefParser.bundle(path.join(__dirname, filePath));
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
