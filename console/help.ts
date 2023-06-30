import { CmdFn } from "../cmd.ts";

const help = (cmds: () => Record<string, unknown>): CmdFn => () => console.log(`caos = content-addressed object store
usage: caos [command] [options]
commands: ${Object.keys(cmds()).join(', ')}`);

export default help;
