import { join } from "https://deno.land/std@0.193.0/path/mod.ts";

// const pid = Deno.pid;
const home = Deno.env.get('HOME');
if (!home) throw new Error('no home :(');
const configRoot = join(home, '.config/caos');
const configPath = join(configRoot, 'caos.json');
// const configLock = join(configRoot, 'caos.lock');

export type CaosOpts = {
  host: string;
  root: string;
  home: string;
}

type OptsApi = {
  init: () => void;
  load: () => CaosOpts;
  save: (value: CaosOpts) => void;
  default: CaosOpts;
}

export const opts: OptsApi = {
  init: () => {
    Deno.mkdirSync(configRoot, {recursive: true});
    opts.save(opts.default);
    return configRoot;
  },
  // TODO
  // lock: () => {
  //   Deno.writeTextFileSync(configLock, pid.toString());
  // },
  load: () => {
    try {
      const json = Deno.readTextFileSync(configPath);
      const value = JSON.parse(json);
      return value;
    } catch (e) {
      if (e instanceof Deno.errors.NotFound) {
        return opts.default;
      } else {
        throw e;
      }
    }
  },
  save: (value) => {
    const json = JSON.stringify(value);
    Deno.writeTextFileSync(configPath, json);
  },
  default: {
    root: '/tmp/caos',
    home: 'd10b49b4',
    host: 'http://localhost:31923',
  }
}

export const loadOpts = opts.load;

export default opts;
