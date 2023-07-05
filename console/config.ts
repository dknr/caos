import { join } from "https://deno.land/std@0.193.0/path/mod.ts";
import { CmdFn } from "../cmd.ts";

const initConfig = () => {

}

const config: CmdFn = (args) => {
  const home = Deno.env.get('HOME');
  if (!home) throw new Error('no home :(');
  const configRoot = join(home, '.config/caos');
  const configPath = join(configRoot, 'caos.json');
  Deno.mkdirSync(configRoot, {recursive: true});
  Deno.writeTextFileSync(configPath, JSON.stringify({
    root: '/tmp/caos',
    host: 'http://localhost:31923'
  }, undefined, 2));
}

export default config;
