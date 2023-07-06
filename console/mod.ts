import { CmdFn } from "../cmd.ts";
import add from "./add.ts";
import addr from "./addr.ts";
import help from "./help.ts";
import init from "./init.ts";
import opts from "./opts.ts";
import push from "./push.ts";
import serve from "./serve.ts";
import tag from "./tag.ts";

export const cmds: Record<string, CmdFn> = {
  add,
  addr,
  init,
  opts,
  push,
  serve,
  tag,
  help: help(() => cmds),
}
