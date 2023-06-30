import { cmds } from "./console/mod.ts";

const cmd = Deno.args[0];
const fn = cmds[cmd || 'help']
if (fn) {
  fn(Deno.args.slice(1));
} else {
  console.log(`command not found: ${cmd}`);
}
