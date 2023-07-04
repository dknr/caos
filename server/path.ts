import { readAll } from "https://deno.land/std@0.192.0/streams/read_all.ts";
import { readerFromStreamReader } from "https://deno.land/std@0.192.0/streams/reader_from_stream_reader.ts";
import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";
import {Caos} from "../types.ts";

const textDecoder = new TextDecoder();

const indexRow = (addr: string, path: string[]) =>
`<pre><a class="path-addr" href="/data/${path[0]}">${path[0].slice(0,8)}</a> <a class="path-name" href="/path/${addr}/${path[1]}">${path[1]}</a></pre>`;

const autoindex = (addr: string, paths: string[][]): string => `<!DOCTYPE html>
<html>
<head>
<title>caos - autoindex</title>
<style>
</style>
</head>
<body>
${(paths.map((path) => indexRow(addr, path))).join('\n')}
</body>
</html>
`

const path = (caos: Caos) => {
  const router = new Router();

  router.get('/:addr/:name*', async (ctx) => {
    const pathAddrs = caos.addr.all(ctx.params.addr);
    if (pathAddrs.length < 1) {
      ctx.response.status = 404;
      ctx.response.headers.set('reason', 'unknown address');
      return;
    } else if (pathAddrs.length > 1) {
      ctx.response.status = 300;
      return;
    }

    const pathAddr = pathAddrs[0];
    const pathType = caos.tags.get(pathAddr, 'type');
    if (pathType !== 'caos/path') {
      ctx.response.status = 404;
      ctx.response.headers.set('reason', `invalid type ${pathType}`);
      return;
    }

    const data = await caos.data.get(pathAddr);
    if (!data) {
      ctx.response.status = 404;
      ctx.response.headers.set('reason', 'no data for address');
      return;
    }

    const pathBytes = await readAll(readerFromStreamReader(data.getReader()));
    const pathFile = textDecoder.decode(pathBytes);
    const paths = pathFile.split('\n').map((line) => line.split(' ', 2));
    // names.forEach(console.log);

    if (ctx.request.url.pathname.endsWith('/')) {
      const name = ctx.params.name;
      if (name) {
        const entries = paths.filter(([_,path]) => path.startsWith(name));
        ctx.response.body = Object.fromEntries(entries);
        return;
      } else {
        // index.html, fall back to autoindex
        const entry = paths.find(([_,path]) => path === `${name}/index.html`);
        if (entry) {
          const addr = entry[0];
          const data = await caos.data.get(addr);
          const type = caos.tags.get(addr, 'type');
          if (data && type === 'index/html') {
            ctx.response.body = data;
            return;
          } else {
            ctx.response.status = 404;
            return;
          }
        } else {
          ctx.response.body = autoindex(pathAddr, paths);
          ctx.response.type = 'text/html';
          return
        }
      }
    } else {
      if (ctx.params.name) {
        // path/[addr]/[path]/[name] look up name and redirect
        const name = (ctx.params.name || '');
        const match = paths.find(([_, path]) => path === name);
        if (!match) {
          ctx.response.status = 404;
          ctx.response.headers.set('reason', 'name not found in path');
          return;
        }

        const addrs = caos.addr.all(match[0]);
        if (addrs.length > 1) {
          ctx.response.status = 300;
          return;
        }
        if (addrs.length < 1) {
          ctx.response.status = 404;
          ctx.response.headers.set('reason', 'addr does not exist for name');
          return;
        }

        const addr = addrs[0];
        ctx.response.body = await caos.data.get(addr);
        ctx.response.type = caos.tags.get(addr, 'type');
      } else {
        // path/[addr] JSON directory listing
        ctx.response.body = Object.fromEntries(paths);
      }
    }
  })

  return router.routes();
}

export default path;
