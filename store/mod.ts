import { Caos } from "../types.ts";
import {openCaosData} from "./data.ts";
import {openCaosMeta} from "./meta.ts";
import log from "../log.ts";
import { CaosOpts } from "../opts.ts";

const notImplemented = () => {
  throw new Error('not implemented');
}

export const openCaos = (opts: CaosOpts): Caos => {
  log(`opening caos: ${opts.root}`);
  Deno.mkdirSync(opts.root, {recursive: true});
  const data = openCaosData(opts);
  const meta = openCaosMeta(opts);
  return {
    addr: {
      all: meta.getAddrs,
      has: meta.hasAddr,
    },
    data: {
      add: async (stream) => {
        const {addr, type, size} = await data.add(stream);
        meta.addAddr(addr);
        meta.setTag(addr, 'type', type || 'application/octet-stream');
        meta.setTag(addr, 'size', size.toString());
        return addr;
      },
      del: data.del,
      get: data.get,
      has: data.has,
    },
    refs: {
      add: notImplemented,
      all: notImplemented,
      del: notImplemented,
      get: notImplemented,
    },
    tags: {
      all: meta.getTags,
      del: (addr, tag) => {
        if (tag === 'size' || tag === 'type') {
          throw new Error(`cannot del inherent tag: ${tag}`);
        }
        meta.delTag(addr, tag);
      },
      get: meta.getTag,
      set: (addr, tag, value) => {
        if (tag === 'size') {
          throw new Error(`cannot set immutable tag: ${tag}`);
        }
        meta.setTag(addr, tag, value);
      },
    },
  };
};
