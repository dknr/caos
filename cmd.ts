export const withArgs = <T>(parseArgs: (args: string[]) => Partial<T>, fn: (opts?: Partial<T>) => void) => (args: string[]) => fn();

export const withDefaults = <T>(defaults: T, fn: (opts: T) => void) => 
  (partialOpts?: Partial<T>) => fn({...defaults, ...partialOpts});

export type CmdFn = (args: string[]) => void | Promise<void>;
