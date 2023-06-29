import {Caos, CaosAddr, CaosData} from "../types.ts";
import { CaosConfig } from "./config.ts";
import * as path from 'https://deno.land/std@0.192.0/path/mod.ts';
import { StreamHasher, StreamTyper, StreamSizer } from '../stream/mod.ts';

export const openCaosData = (config: CaosConfig) => {
  const dataPath = path.join(config.path, 'data');
  const nameFile = (addr: CaosAddr) => path.join(dataPath, addr);
  const tempPath = path.join(config.path, 'temp');

  Deno.mkdirSync(dataPath, {recursive: true});
  Deno.mkdirSync(tempPath, {recursive: true});

  return {
    add: async (data: ReadableStream<Uint8Array>) => {
      const tempName = await Deno.makeTempFile({dir: tempPath});
      const tempFile = await Deno.open(tempName, {create: true, write: true});
      const hasher = new StreamHasher();
      const typer = new StreamTyper();
      const sizer = new StreamSizer();

      await data
        .pipeThrough(hasher)
        .pipeThrough(typer)
        .pipeThrough(sizer)
        .pipeTo(tempFile.writable);
      
      const addr = hasher.digest();
      const type = typer.type;
      const size = sizer.size;

      await Deno.copyFile(tempName, nameFile(addr))
      await Deno.remove(tempName);

      return {addr, size, type};
    },
    del: async (addr: CaosAddr) => {
      try {
        await Deno.remove(nameFile(addr));
      } catch (e) {
        if (e instanceof Deno.errors.NotFound) {
          return;
        } else {
          throw e;
        }
      }
    },
    get: (addr: CaosAddr): Promise<CaosData | undefined> => Deno.open(nameFile(addr), {read: true})
      .then((file) => file.readable)
      .catch(() => undefined),
    has: (addr: CaosAddr) => Deno.stat(nameFile(addr))
      .then((stat) => !!stat)
      .catch(() => false),
  }
}