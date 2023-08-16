import {Application, Router} from "https://deno.land/x/oak@v12.1.0/mod.ts";
import log from "../log.ts";
import {Caos} from "../types.ts";
import addr from "./addr.ts";
import data from "./data.ts";
import find from "./find.ts";
import tags from "./tags.ts";
import path from "./path.ts";
import { CaosOpts } from "../opts.ts";

export const serveCaos = (caos: Caos, opts: CaosOpts) => {
  const hostUrl = new URL(opts.host);
  if (hostUrl.hostname !== 'localhost') {
    console.log(`invalid host for serve command: ${hostUrl.hostname}`);
    Deno.exit(1);
  }
  const hostPort = parseInt(hostUrl.port);
  if (hostPort < 1024) {
    console.log(`invalid port for serve command: ${hostPort}`)
    Deno.exit(1);
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
  router.use("/find", find(caos));
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
  app.listen({ port: hostPort });
};
