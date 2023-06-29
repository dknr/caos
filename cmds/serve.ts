import { Application, Router } from "https://deno.land/x/oak@v12.1.0/mod.ts";
import { openCaos } from "../store/mod.ts";
import log from "../log.ts";
import { withArgs, withDefaults } from "../cmd.ts";

type ServeOpts = {
  home: string;
  path: string;
}

const serve = (opts: ServeOpts) => {
  const app = new Application();
  const router = new Router();
  const caos = openCaos({path: opts.path});

  router.get("/", (ctx) => {
    ctx.response.status = 303;
    ctx.response.headers.set(
      "location",
      `/data/${opts.home}`,
    );
  });

  router.get("/addr/:addr", (ctx) => {
    const {addr} = ctx.params;
    const addrs = caos.addr.all(addr);
    ctx.response.body = addrs;
    if (addrs.length > 1) {
      ctx.response.status = 300;
    } else if (addrs.length < 1) {
      ctx.response.status = 404;
    }
  });

  router.post("/data", async (ctx) => {
    const data = ctx.request.body({ type: "stream" }).value;
    const result = await caos.data.add(data);
    ctx.response.body = result;
  });
  router.post("/data/:file", async (ctx) => {
    const data = ctx.request.body({ type: "stream" }).value;
    const result = await caos.data.add(data);
    ctx.response.body = result;
  });

  router.get("/data/:addr", async (ctx) => {
    const addr = ctx.params.addr;    
    const data = await caos.data.get(addr);
    if (data) {
    const type = caos.tags.get(addr, "type");
      ctx.response.headers.set(
        "content-type",
        type || "application/octet-stream",
      );
      ctx.response.body = data;
    } else {
      ctx.response.status = 404;
    }
  });

  router.get("/tags/:addr", (ctx) => {
    const addr = ctx.params.addr;
    const tags = caos.tags.all(addr);
    ctx.response.body = tags;
  });

  router.get("/tags/:addr/:tag", (ctx) => {
    const { addr, tag } = ctx.params;
    const value = caos.tags.get(addr, tag);

    if (value) {
      ctx.response.body = value;
    } else {
      ctx.response.status = 404;
    }
  });

  router.put("/tags/:addr/:tag", async (ctx) => {
    const { addr, tag } = ctx.params;
    const value = await ctx.request.body({type: 'text'}).value;
    if (value) {
      if (caos.addr.has(addr)) {
        caos.tags.set(addr, tag, value);
        ctx.response.status = 204;
      } else {
        ctx.response.status = 404;
      }
    } else {
      ctx.response.status = 400;
    }
  });

  router.delete("/tags/:addr/:tag", async (ctx) => {
    const { addr, tag } = ctx.params;
    if (caos.addr.has(addr)) {
      caos.tags.del(addr, tag);
      ctx.response.status = 204;
    }
  })

  app.use(async (ctx, next) => {
    try {
      await next();
      log([ctx.response.status, ctx.request.method, ctx.request.url.pathname].join(' '));
    } catch(e) {
      log(`ERR ${ctx.request.method.padStart(4)} ${ctx.request.url.pathname}`);
      throw e;
    }
  });
  app.use(router.routes());
  app.use(router.allowedMethods());

  app.addEventListener(
    "listen",
    (e) => log(`serving caos: http://localhost:${e.port}`),
  );
  app.listen({ port: 31923 });
};

export default withArgs(withDefaults({
  home: 'd10b49b4cf4f9204c4a6e4a96e5a004fa25768623667b2aec05f82e4852aaa91',
  path: '/tmp/caos',
}, serve));

