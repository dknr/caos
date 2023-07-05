import { CmdFn } from "../cmd.ts";
import opts from "./opts.ts";

const init: CmdFn = () => {
  opts(['init']); // alias
}

export default init;
