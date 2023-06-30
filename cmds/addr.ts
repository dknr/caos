import { buildClient } from "../client/mod.ts";
import { CmdFn } from "../cmd.ts";
import hosts from "../hosts.ts";
import { assertArgsCount } from "../util.ts";

const addr: CmdFn = async (args) => {
  assertArgsCount(args, 1,1);
  const client = await buildClient({host: hosts.localhost})
  const results = await client.addr.all(args[0]);
  if (results?.length > 0) {
    results.forEach((result) => console.log(result));
  } else {
    console.error('addr not found');
    Deno.exit(-1);
  }
}

export default addr;
