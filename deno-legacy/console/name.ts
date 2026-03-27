import {CmdFn} from "../cmd.ts";
import {buildClient} from "../client/mod.ts";
import {assertArgsCount} from "../util.ts";

const name: CmdFn = async (args, opts) => {
  assertArgsCount(args, 1, 2);
  const client = buildClient(opts);

  if (args.length == 1) {
    const addr = await client.name.get(args[0]);
    console.log(addr);
    return;
  }

  if (args.length == 2) {
    void client.name.set(args[0], args[1]);
  }
}

export default name;
