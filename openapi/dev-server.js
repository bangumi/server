const fs = require("fs");

const path = require("path");

const lodash = require("lodash");
const proxy = require("koa-proxy");
const livereload = require("livereload");
const $RefParser = require("@apidevtools/json-schema-ref-parser");
const Koa = require("koa");
const Router = require("@koa/router");

const goDevServer = `http://localhost:${process.env.HTTP_PORT ?? 3000}`;

const app = new Koa();
const router = new Router();

router.get("/v0/", v0Index);
router.get("/private/", privateIndex);

router.get("/v0.json", v0Api);
router.get("/private.json", privateApi);

const privateDescription = `
## Development Note 

Swagger UI 不支持发送 Cookies 

所以 \`openapi/dev-server.js\` 在转发请求时会添加 HTTP Header \`Cookie: sessionID=dev-session-id\` 

此 \`sessionID\` 是 [bangumi/dev-env](https://github.com/bangumi/dev-env) 填充的开发用数据，
代表开发环境中包含树洞账号 <https://bgm.tv/user/382951>

因此，在浏览器的开发工具中看到的网络请求和后端实际收到的请求可能未必相同。
`;

async function v0Index(ctx) {
  ctx.type = "text/html; charset=utf-8";
  ctx.body = fs.createReadStream(path.join(__dirname, "static", "v0.html"));
}

async function privateIndex(ctx) {
  ctx.type = "text/html; charset=utf-8";
  ctx.body = fs.createReadStream(path.join(__dirname, "static", "private.html"));
}

async function v0Api(ctx) {
  const openapi = await $RefParser.bundle(path.join(__dirname, "v0.yaml"));
  ctx.body = lodash.set(openapi, "servers[0].url", goDevServer);
}

async function privateApi(ctx) {
  const origin = ctx.request.headers["origin"] ?? "http://localhost:3001";
  const openapi = await $RefParser.bundle(path.join(__dirname, "private.yaml"));

  lodash.set(openapi, "servers[0]", {
    url: origin + `/`,
    description: `proxy to ${goDevServer} on same host`,
  });

  lodash.set(openapi, "info.description", lodash.get(openapi, "info.description", "") + "\n\n" + privateDescription);

  ctx.body = openapi;
}

const lrServer = livereload.createServer({
  exts: ["yaml"],
});

lrServer.watch(__dirname);

// const proxyRouter = new Router({prefix: "/p"})

router.all("/p/(.*)", async (ctx, next) => {
  ctx.request.headers.cookie = ctx.request.headers.cookie ?? "sessionID=dev-session-id";
  return proxy({
    jar: true,
    host: goDevServer,
  })(ctx, next);
});

// app.use(proxyRouter.routes())
app.use(router.routes());

app.listen(3001, function () {
  console.log("see private swagger at http://127.0.0.1:3001/private/");
  console.log("ses v0 swagger at http://127.0.0.1:3001/v0/");
});
