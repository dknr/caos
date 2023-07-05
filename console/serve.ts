import { CmdFn } from "../cmd.ts";
import { loadOpts } from "../opts.ts";
import { serveCaos } from "../server/mod.ts";
import { openCaos } from "../store/mod.ts";

const serve: CmdFn = (args) => {
  if (args[0] === '--help' || args[0] === '-h') {
    console.log('usage: caos serve');
    return;
  }

  const opts = loadOpts();
  const caos = openCaos(opts);
  serveCaos(caos, opts);
};

export default serve;
