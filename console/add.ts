import { CmdFn } from "../cmd.ts";
import opts from "../opts.ts";


const add: CmdFn = async (args) => {
  const host = opts.load().host;
  const path = args.pop();
  if (!path) {
    console.log('no path specified');
    console.log('usage: caos add [options] <path>');
    Deno.exit(-1);
  }
  
  const file = await Deno.open(path);
  const postDataResult = await fetch(`${host}/data`, {
    method: 'post',
    body: file.readable,
  });

  if (!postDataResult.ok) {
    console.log('failed to upload file');
    console.log(`status: ${postDataResult.status} ${postDataResult.statusText}`);
    Deno.exit(postDataResult.status);
  }

  const addr = await postDataResult.text();
  console.log(addr);
};

export default add;
