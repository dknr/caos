import { CmdFn } from "./cmd.ts";
import add from "./cmds/add.ts";
import help from "./cmds/help.ts";
import serve from "./cmds/serve.ts";

const cmd = Deno.args[0];
const cmds: Record<string, CmdFn> = {
  add,
  serve,
  help: help(() => cmds),
}

const fn = cmds[cmd || 'help']
if (fn) {
  fn(Deno.args.slice(1));
} else {
  console.log(`command not found: ${cmd}`);
}