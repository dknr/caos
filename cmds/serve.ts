import { Application, Router } from "https://deno.land/x/oak@v12.1.0/mod.ts";
import { openCaos } from "../store/mod.ts";
import log from "../log.ts";

type ServeOpts = {
  home: string;
  path: string;
}

const withArgs = <T>(fn: (opts?: Partial<T>) => void) => (args: string[]) => fn();

const withDefaults = <T>(defaults: T, fn: (opts: T) => void) => 
  (partialOpts?: Partial<T>) => fn({...defaults, ...partialOpts});

const serve = (opts: ServeOpts) => {
  const app = new Application();
  const router = new Router();
  const caos = openCaos({path: opts.path});

  router.get("/", (ctx) => {
    ctx.response.status = 302;
    ctx.response.headers.set(
      "location",
      `/data/${opts.home}`,
    );
  });

  router.post("/data", async (ctx) => {
    const data = ctx.request.body({ type: "stream" }).value;
    const result = await caos.addData(data);
    ctx.response.body = result;
  });
  router.post("/data/:file", async (ctx) => {
    const data = ctx.request.body({ type: "stream" }).value;
    const result = await caos.addData(data);
    ctx.response.body = result;
  });

  router.get("/data/:addr", async (ctx) => {
    const addr = ctx.params.addr;
    const data = await caos.getData(addr);
    const type = caos.getTag(addr, "type");
    if (data) {
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
    const tags = caos.getTags(addr);
    ctx.response.body = tags;
  });

  router.get("/tags/:addr/:tag", (ctx) => {
    const { addr, tag } = ctx.params;
    const value = caos.getTag(addr, tag);

    if (value) {
      ctx.response.body = value;
    } else {
      ctx.response.status = 404;
    }
  });

  app.use(async (ctx, next) => {
    try {
      await next();
      log(
        `${ctx.response.status} ${
          ctx.request.method.padStart(4)
        } ${ctx.request.url.pathname}`,
      );
    } catch {
      log(`ERR ${ctx.request.method.padStart(4)} ${ctx.request.url.pathname}`);
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

