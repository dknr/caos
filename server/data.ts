import {Caos} from "../types.ts";
import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";

const data = (caos: Caos) => {
  const router = new Router();

  router.post("/", async (ctx) => {
    const data = ctx.request.body({type: "stream"}).value;
    const result = await caos.data.add(data);
    ctx.response.body = result;
  });

  router.post("/:file", async (ctx) => {
    const data = ctx.request.body({type: "stream"}).value;
    const result = await caos.data.add(data);
    ctx.response.body = result;
  });

  router.get("/:addr", async (ctx) => {
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

  return router.routes();
}

export default data;
