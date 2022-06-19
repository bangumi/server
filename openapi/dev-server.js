const fs = require("fs");

const path = require("path");
const _ = require("lodash");

const livereload = require("livereload");
const $RefParser = require("@apidevtools/json-schema-ref-parser");
const Koa = require("koa");
const Router = require("@koa/router");

const app = new Koa();
const router = new Router();

router.get("/", index).get("/dist.json", show);

async function index(ctx) {
  ctx.type = "text/html; charset=utf-8";
  ctx.body = fs.createReadStream(path.join(__dirname, "static", "index.html"));
}

async function show(ctx) {
  const openapi = await $RefParser.bundle(path.join(__dirname, "v0.yaml"));
  _.set(
    openapi,
    "servers[0].url",
    `http://localhost:${process.env.HTTP_PORT ?? 3000}`
  );
  ctx.body = openapi;
}

const lrServer = livereload.createServer({
  exts: ["yaml"],
});

lrServer.watch(__dirname);

app.use(router.routes());

app.listen(3001, function () {
  console.log("http://127.0.0.1:3001/");
});
