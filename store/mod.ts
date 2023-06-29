import { Caos } from "../types.ts";
import { CaosConfig } from "./config.ts";
import {openCaosData} from "./data.ts";
import {openCaosMeta} from "./meta.ts";

export const openCaos = (config: CaosConfig): Caos => {
  Deno.mkdirSync(config.path, {recursive: true});
  const data = openCaosData(config);
  const meta = openCaosMeta(config);
  return {
    addData: async (stream) => {
      const {addr, type, size} = await data.add(stream);
      meta.addAddr(addr);
      meta.setTag(addr, 'type', type || 'application/octet-stream');
      meta.setTag(addr, 'size', size.toString());
      return addr;
    },
    setTag: (addr, tag, value) => {
      if (tag === 'size' || tag === 'type') { // immutable tags
        throw new Error(`attempted to set immutable tag ${tag}: ${value}`);
      }
      meta.setTag(addr, tag, value);
      return;
    },
    addRef: () => {
      throw new Error("not implemented");
    },
    getData: (addr) => {
      return data.get(addr);
    },
    getTag: (addr, tag) => {
      const value = meta.getTag(addr, tag);
      return String(value);
    },
    getTags: (addr) => {
      return meta.getTags(addr);
    },
    getRefs: () => {
      throw new Error("not implemented");
    },
    hasData: () => {
      throw new Error("not implemented");
    },
    delete: () => {
      throw new Error("not implemented");
    },
  };
};
