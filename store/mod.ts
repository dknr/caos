import { Caos } from "../types.ts";
import { CaosConfig } from "./config.ts";
import {openCaosData} from "./data.ts";
import {openCaosMeta} from "./meta.ts";
import log from "../log.ts";

const notImplemented = () => {
  throw new Error('not implemented');
}

export const openCaos = (config: CaosConfig): Caos => {
  log(`opening caos: ${config.path}`);
  Deno.mkdirSync(config.path, {recursive: true});
  const data = openCaosData(config);
  const meta = openCaosMeta(config);
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
