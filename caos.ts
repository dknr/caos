import { loadOpts } from "./opts.ts";
import { cmds } from "./console/mod.ts";

const cmd = Deno.args[0];
const fn = cmds[cmd || 'help']
if (fn) {
  const opts = loadOpts();
  fn(Deno.args.slice(1), opts);
} else {
  console.log(`command not found: ${cmd}`);
}
