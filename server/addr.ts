import {Caos} from "../types.ts";
import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";

const addr = (caos: Caos) => {
  const router = new Router();

  router.get('/:addr', (ctx) => {
    const {addr} = ctx.params;
    const addrs = caos.addr.all(addr);
    ctx.response.body = addrs;
    if (addrs.length > 1) {
      ctx.response.status = 300;
    } else if (addrs.length < 1) {
      ctx.response.status = 404;
    }
  });

  return router.routes();
}

export default addr;