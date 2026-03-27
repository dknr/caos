import { buildClient } from "../client/mod.ts";
import { CmdFn } from "../cmd.ts";

const add: CmdFn = async (args, opts) => {
  const client = buildClient(opts);
  const path = args.pop();
  if (!path) {
    console.log('no path specified');
    console.log('usage: caos add [options] <path>');
    Deno.exit(-1);
  }
  
  const file = await Deno.open(path);
  const addr = await client.data.add(file.readable);

  console.log(addr);
};

export default add;
