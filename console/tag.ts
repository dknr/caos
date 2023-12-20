import { CaosClient,buildClient } from "../client/mod.ts";
import { CmdFn } from "../cmd.ts";
import { loadOpts } from "../opts.ts";
import {assertArgsCount, resolveAddress} from "../util.ts";

const get = async (client: CaosClient, args: string[]) => {
  assertArgsCount(args, 1);

  const addr = await resolveAddress(client, args[0]);
  const tag = args[1];
  if (tag) {
    const value = await client.tags.get(addr, tag);
    console.log(value);
  } else {
    const tags = await client.tags.all(addr);
    Object.entries(tags).forEach(([tag, value]) => console.log(`${tag}: ${value}`));
  }
}

const set = async (client: CaosClient, args: string[]) => {
  assertArgsCount(args, 3);

  const addr = await resolveAddress(client, args[0]);
  const tag = args[1];
  const value = args[2];

  client.tags.set(addr, tag, value);
}

const del = async (client: CaosClient, args: string[]) => {
  assertArgsCount(args, 2);

  const addr = await resolveAddress(client, args[0]);
  const tag = args[1];

  await client.tags.del(addr, tag);
}

const help = async () => {
  console.error('caos tag get <addr> [tag]');
  console.error('caos tag set <addr> <tag> <value>');
  console.error('caos tag del <addr> <tag>');
  console.error('caos tag help');
  Deno.exit(-1);
}

const ops: Record<string, (client: CaosClient, args: string[]) => void | Promise<void>> = {
  get,
  set,
  del,
  help
};

const tag: CmdFn = async (args) => {
  const {host} = loadOpts();
  const client = buildClient({host});
  await (ops[args[0]] || ops.help)(client, args.slice(1));
}

export default tag;
