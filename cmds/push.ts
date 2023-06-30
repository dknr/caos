import { expandGlob } from "https://deno.land/std@0.192.0/fs/mod.ts";
import { walk } from "https://deno.land/std@0.192.0/fs/mod.ts";
import { buildClient } from "../client/mod.ts";
import { CmdFn } from "../cmd.ts";
import hosts from '../hosts.ts';
import addr from "../server/addr.ts";
import { CaosAddr } from "../types.ts";

const textEncoder = new TextEncoder();

const push: CmdFn = async (args) => {
  const client = buildClient({host: hosts.localhost})
  const names: string[] = [];

  for (const arg of args) {
    for await (const entry of walk(arg)) {
      if (entry.isFile) {
        const file = await Deno.open(entry.path);
        const addr = await client.data.add(file.readable);
        const name = `${addr.slice(0,8)} ${entry.path}`;
        names.push(name);
        console.log(name);
      }
    }
  }

  const pathBytes = textEncoder.encode(names.join('\n'));
  const pathAddr = await client.data.add(pathBytes);
  await client.tags.set(pathAddr, 'type', 'caos/path');

  console.log(`${hosts.localhost}/path/${pathAddr.slice(0,12)}`);
};

export default push;
