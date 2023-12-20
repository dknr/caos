import { Middleware } from "https://deno.land/x/oak@v12.1.0/mod.ts";

const auth = (opts: {token?: string}): Middleware => (ctx, next) => {
  if (ctx.request.method === 'GET') {
    return next();
  }

  const authorization = ctx.request.headers.get('Authorization');
  const token = authorization?.match(/Bearer (.*)/)?.[1];
  if (token && opts.token && token === opts.token) {
    return next();
  } else {
    ctx.response.status = 401;
  }
}

export default auth;
