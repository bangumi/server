const fs = require("fs")

const path = require('path')

const livereload = require('livereload');
const $RefParser = require("@apidevtools/json-schema-ref-parser");
const Koa = require('koa');
const Router = require('@koa/router');

const app = new Koa();
const router = new Router();

router
  .get('/', index)
  .get('/dist.json', show)

async function index(ctx) {
  ctx.type = "text/html; charset=utf-8"
  ctx.body = fs.createReadStream(path.join(__dirname, "static", 'index.html'))
}


async function show(ctx) {
  ctx.body = await $RefParser.bundle(path.join(__dirname, "v0.yaml"))
}

const lrServer = livereload.createServer({
  exts: ['yaml']
});
console.log(__dirname)
lrServer.watch(__dirname);

app.use(router.routes());
app.listen(3000);
