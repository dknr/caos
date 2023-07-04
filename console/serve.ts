import { CmdFn } from "../cmd.ts";
import { serveCaos } from "../server/mod.ts";
import { openCaos } from "../store/mod.ts";

const serve: CmdFn = (args) => {
  if (args[0] === '--help' || args[0] === '-h') {
    console.log('usage: caos serve [<path> [<home>]]');
    return;
  }

  const path = args[0] || '/tmp/caos';
  const home = args[1] || 'd10b49b4cf4f9204c4a6e4a96e5a004fa25768623667b2aec05f82e4852aaa91'
  const caos = openCaos({path});
  serveCaos(caos, {home});
};

export default serve;
