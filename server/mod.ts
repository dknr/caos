import {Application, Router} from "https://deno.land/x/oak@v12.1.0/mod.ts";
import log from "../log.ts";
import {Caos} from "../types.ts";
import addr from "./addr.ts";
import data from "./data.ts";
import tags from "./tags.ts";
import path from "./path.ts";
import { CaosOpts } from "../opts.ts";


export const serveCaos = (caos: Caos, opts: CaosOpts) => {
  const app = new Application();
  const router = new Router();

  router.get("/", (ctx) => {
    ctx.response.status = 303;
    ctx.response.headers.set(
      "location",
      `/data/${opts.home}`,
    );
  });

  router.use("/addr", addr(caos));
  router.use("/data", data(caos));
  router.use("/tags", tags(caos));
  router.use("/path", path(caos));

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
    (e) => log(`serving caos: http://localhost:${e.port}`),
  );
  app.listen({ port: 31923 });
};
