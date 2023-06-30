import { readAll } from "https://deno.land/std@0.192.0/streams/read_all.ts";
import { readerFromStreamReader } from "https://deno.land/std@0.192.0/streams/reader_from_stream_reader.ts";
import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";
import {Caos} from "../types.ts";

const textDecoder = new TextDecoder();

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
    const names = pathFile.split('\n').map((line) => line.split(' ', 2));
    names.forEach(console.log);

    const name = (ctx.params.name || '');
    const match = names.find(([_, lineName]) => lineName === name);
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
    ctx.response.status = 301;
    ctx.response.headers.set('location', `/data/${addr}`);
  })

  return router.routes();
}

export default path;
