import serve from "./cmds/serve.ts";

type CmdFn = (args: string[]) => void;

const help: CmdFn = () => console.log(
`usage: caos <command> [options]
commands: ${Object.keys(cmds).join(', ')}`
);

const cmd = Deno.args[0];
const cmds: Record<string, CmdFn> = {
  serve,
  help
}

cmds[cmd || 'help'](Deno.args.slice(1));
