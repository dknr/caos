import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";

const name = (store: { get: (n: string) => string | undefined; set: (n: string, addr: string) => void }) => {
  const router = new Router();

  router.get('/:name', (ctx) => {
    const addr = store.get(ctx.params.name);
    if (!addr) {
      ctx.response.status = 404;
      return;
    }

    ctx.response.body = addr;

    ctx.response.status = 302;
    ctx.response.headers.set(
      "location",
      `/${addr}`,
    );
  })

  router.post('/:name', async (ctx) => {
    const name = ctx.params.name;
    const body = ctx.request.body({ type: 'text' });
    const addr = await body.value;

    store.set(name, addr);

    ctx.response.status = 200;
    ctx.response.body = `${name} ${addr}`;
  })

  return router.routes();
}

export default name;
