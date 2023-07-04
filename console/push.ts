import { WalkEntry, walkSync } from "https://deno.land/std@0.192.0/fs/mod.ts";
import { buildClient } from "../client/mod.ts";
import { CmdFn } from "../cmd.ts";
import hosts from '../hosts.ts';

const textEncoder = new TextEncoder();

const compareEntries = (a: WalkEntry, b: WalkEntry): number => {
  if (a.path > b.path)
    return 1;
  if (a.path < b.path)
    return -1;
  return 0;
}

const push: CmdFn = async (args) => {
  const host = hosts.localhost;
  const client = buildClient({host});
  const paths: string[] = [];

  const entries = args
    .flatMap((arg) => Array.from(walkSync(arg)))
    .filter((entry) => entry.isFile);
  entries.sort(compareEntries)

  for (const entry of entries) {
    const file = await Deno.open(entry.path);
    const addr = await client.data.add(file.readable);
    const path = `${addr} ${entry.path}`;
    paths.push(path);
    console.log(`${addr.slice(0,8)} ${entry.path}`);
  }

  const pathBytes = textEncoder.encode(paths.join('\n'));
  const pathAddr = await client.data.add(pathBytes);
  await client.tags.set(pathAddr, 'type', 'caos/path');

  console.log(`${host}/path/${pathAddr.slice(0,8)}`);
};

export default push;
