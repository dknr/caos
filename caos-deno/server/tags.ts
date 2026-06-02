import {Caos} from "../types.ts";
import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";

const tags = (caos: Caos) => {
  const router = new Router();

  router.get("/:addr", (ctx) => {
    const addr = ctx.params.addr;
    const tags = caos.tags.all(addr);
    ctx.response.body = tags;
  });

  router.get("/:addr/:tag", (ctx) => {
    const {addr, tag} = ctx.params;
    const value = caos.tags.get(addr, tag);

    if (value) {
      ctx.response.body = value;
    } else {
      ctx.response.status = 404;
    }
  });

  router.put("/:addr/:tag", async (ctx) => {
    const {addr, tag} = ctx.params;
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

  router.delete("/:addr/:tag", async (ctx) => {
    const {addr, tag} = ctx.params;
    if (caos.addr.has(addr)) {
      caos.tags.del(addr, tag);
      ctx.response.status = 204;
    }
  })

  return router.routes();
}

export default tags;
