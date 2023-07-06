import {Application, Router} from "https://deno.land/x/oak@v12.1.0/mod.ts";
import log from "../log.ts";
import {Caos} from "../types.ts";
import addr from "./addr.ts";
import data from "./data.ts";
import tags from "./tags.ts";
import path from "./path.ts";
import { CaosOpts } from "../opts.ts";

export const serveCaos = (caos: Caos, opts: CaosOpts) => {
  if (opts.host !== 'http://localhost:31923') {
    console.log('refusing to serve unknown host');
    console.log(opts.host);
    console.log('try: caos opts set host http://localhost:31923');
    Deno.exit(-1);
  }

  const app = new Application();
  const router = new Router();

  router.get("/", (ctx) => {
    ctx.response.status = 302;
    ctx.response.headers.set(
      "location",
      opts.home,
    );
  });

  router.use("/addr", addr(caos));
  router.use("/data", data(caos));
  router.use("/path", path(caos));
  router.use("/tags", tags(caos));

  app.use(async (ctx, next) => {
    try {
      await next();
      log([ctx.response.status, ctx.request.method.padStart(4), ctx.request.url.pathname].join(' '));
    } catch(e) {
      log(`ERR ${ctx.request.method.padStart(4)} ${ctx.request.url.pathname}`);
      throw e;
    }
  });
  app.use(router.routes());
  app.use(router.allowedMethods());

  app.addEventListener(
    "listen",
    (e) => {
      const {home, host} = opts;
      log(`serving caos: http://localhost:${e.port}`);
      log(`host: ${host}`);
      log(`home: ${host}/${home}`);
    }
  );
  app.listen({ port: 31923 });
};
