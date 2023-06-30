import { CmdFn } from "./cmd.ts";
import add from "./cmds/add.ts";
import help from "./cmds/help.ts";
import push from "./cmds/push.ts";
import serve from "./cmds/serve.ts";
import tag from "./cmds/tag.ts";
import addr from "./cmds/addr.ts";

const cmd = Deno.args[0];
const cmds: Record<string, CmdFn> = {
  add,
  addr,
  push,
  serve,
  tag,
  help: help(() => cmds),
}

const fn = cmds[cmd || 'help']
if (fn) {
  fn(Deno.args.slice(1));
} else {
  console.log(`command not found: ${cmd}`);
}