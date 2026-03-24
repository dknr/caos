import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";

const longestName = 'data/60ea108498aed632cb2c4c0bc73f757fb75ace85cabfa092d55df8a0fc3b586b'

const name = () => {
  const router = new Router();
  const names = new Map<string, string>(); // TODO: persistence

  router.get('/:name', (ctx) => {
    const addr = names.get(ctx.params.name);
    if (!addr) {
      ctx.response.status = 404;
      return;
    }

    ctx.response.body = addr;

    ctx.response.status = 302;
    ctx.response.headers.set(
      "location",
      `/${addr}`  // TODO: robustify request root
    );
  })

  router.post('/:name', async (ctx) => {
    const name = ctx.params.name;
    const body = ctx.request.body({ type: 'text' });
    const addr = await body.value;

    names.set(name, addr);

    ctx.response.status = 200;
    ctx.response.body = `${name} ${addr}`;
  })

  return router.routes();
}

export default name;
