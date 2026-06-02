import {CaosOpts} from './opts.ts'

export type CmdFn = (args: string[], opts: CaosOpts) => void | Promise<void>;
