import { readAll } from "https://deno.land/std@0.192.0/streams/read_all.ts";
import { readerFromStreamReader } from "https://deno.land/std@0.192.0/streams/reader_from_stream_reader.ts";
import {Router} from "https://deno.land/x/oak@v12.1.0/router.ts";
import {Caos, CaosAddr} from "../types.ts";
import {a, div, page, pre, span} from '../html.ts';

const textDecoder = new TextDecoder();

const shortType = (type: string | undefined) => {
  return type?.split('/').map((a) => a[0]).join('/') || '???';
}

const prefix = ['', 'k', 'M', 'G', 'T', 'P', 'E'];
const shortSize = (size: number, length = 4): string => {
  const numerals = Math.ceil(Math.log10(size));
  if (numerals <= length) 
    return size.toString().padStart(length);

  let value = size;
  let result = '';
  let scale = 0;
  while (value > 0) {
    const part = value % 1000;
    result = part + prefix[scale] + result;
    value = Math.floor(value / 1000);
    scale++;
  }

  return result.slice(0,length);
}

type Result<TValue, TElse = string> = {
  ok: true;
  value: TValue
} | {
  ok: false;
  else: TElse;
};

type Path = [CaosAddr, string];
type Paths = Path[];

const path = (caos: Caos) => {
  const router = new Router();

  const row = (addr: string, path: string[]) => {
    const type = shortType(caos.tags.get(path[0], 'type'));
    const size = shortSize(parseInt(caos.tags.get(path[0], 'size') || ''));

    return div({class: 'path'},
      pre([
        a({class: 'path-addr', href: `/data/${path[0]}`}, path[0].slice(0,8)),
        type, size, 
        a({class: 'path-name', href: `/path/${addr}/${path[1]}`}, path[1]),
      ].join(' ')),
    );
  }

  const autoindex = (addr: string, paths: string[][]) => page(
    addr.slice(0,8),
    ...paths.map((path) => row(addr, path))
  );

  const openPathFile = async (addr: string): Promise<Result<Paths,{ reason?: string; status: number; }>> => {
    const pathAddrs = caos.addr.all(addr);
    if (pathAddrs.length < 1) {
      return {ok: false, else: {status: 404, reason: 'unknown addr'}};
    } else if (pathAddrs.length > 1) {
      return {ok: false, else: {status: 300}};
    }

    const pathAddr = pathAddrs[0];
    const pathType = caos.tags.get(pathAddr, 'type');
    if (pathType !== 'caos/path') {
      return {ok: false, else: {status: 404, reason: `invalid type ${pathType}`}};
    }

    const data = await caos.data.get(pathAddr);
    if (!data) {
      return {ok: false, else: {status: 404, reason: 'no data for address'}};
    }

    const pathBytes = await readAll(readerFromStreamReader(data.getReader()));
    const pathFile = textDecoder.decode(pathBytes);
    const pathLines = pathFile.split('\n');
    const paths: [string, string][] = [];
    for (const line of pathLines) {
      const path = line.match(/^(\w*) (.*)/)?.slice(1,3);
      if (path) 
        paths.push(path as [string, string]);
    }

    return {ok: true, value: paths};
  }

  router.get('/:addr', async (ctx) => {
    if (!ctx.request.url.pathname.endsWith('/')) {
      ctx.response.status = 301;
      ctx.response.headers.set('location', ctx.request.url.pathname + '/');
    }

    const pathFile = await openPathFile(ctx.params.addr);
    if (!pathFile.ok) {
      ctx.response.status = pathFile.else.status;
      if (pathFile.else.reason)
        ctx.response.headers.set('reason', pathFile.else.reason);
      return;
    }
    const paths = pathFile.value;
    const indexPath = paths.find((path) => path[1] === 'index.html');
    if (indexPath) {
      const indexData = await caos.data.get(indexPath[0]);
      const indexType = caos.tags.get(indexPath[0], 'type');
      if (indexData && indexType === 'text/html') {
        ctx.response.body = indexData;
        ctx.response.type = indexType;
      } else {
        ctx.response.status = 404;
      }
    } else {
      ctx.response.body = autoindex(ctx.params.addr, paths);
      ctx.response.type = 'text/html';
    }
  });

  router.get('/:addr/:name*', async (ctx) => {
    const pathFile = await openPathFile(ctx.params.addr);
    if (!pathFile.ok) {
      ctx.response.status = pathFile.else.status;
      if (pathFile.else.reason)
        ctx.response.headers.set('reason', pathFile.else.reason);
      return;
    }
    const paths = pathFile.value;

    if (ctx.request.url.pathname.endsWith('/')) {
      const indexName = `${ctx.params.name}/index.html`;
      const indexPath = paths.find((path) => path[1] === indexName);
      console.log({indexName, indexPath});
      if (indexPath) {
        const indexData = await caos.data.get(indexPath[0]);
        const indexType = caos.tags.get(indexPath[0], 'type');
        if (indexData && indexType === 'text/html') {
          ctx.response.body = indexData;
          ctx.response.type = indexType;
        } else {
          ctx.response.status = 404;
        }
      } else {
        ctx.response.status = 404;
      }
    } else {
      const match = paths.find(([_, path]) => path === ctx.params.name);
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
    }
  })

  return router.routes();
}

export default path;
