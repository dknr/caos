import { CmdFn } from "../cmd.ts";
import _opts from '../opts.ts';

const opts: CmdFn = (args) => {
  if (args.length < 1) {
    console.log('usage: caos opts init');
    console.log('       caos opts get');
    console.log('       caos opts set [key] [value]');
    Deno.exit(-1);
  }

  switch (args[0]) {
    case 'init':{
      const path = _opts.init();
      console.log(`caos init: ${path}`);
      opts(['get']);
      return;
    }
    case 'get':{
      const values = _opts.load();
      console.log(JSON.stringify(values, undefined, 2));
      return;
    }
    case 'set':{
      const key = args[1];
      const value = args[2];
      if (key && value) {
        _opts.save({
          ..._opts.load(),
          [key]: value,
        });
      } else {
        console.log('usage: caos opts set [key] [value]');
        Deno.exit(-1);
      }
      return;
    }
  }
}

export default opts;
