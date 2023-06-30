import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";

const name = () => {
  const router = new Router();

  router.get('/name/:path*', (ctx) => {

  })

  return router;
}