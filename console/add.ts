import { CmdFn } from "../cmd.ts";

const config = {
  host: 'http://localhost:31923'
}

const add: CmdFn = async (args) => {
  const path = args.pop();
  if (!path) {
    console.log('no path specified');
    console.log('usage: caos add [options] <path>');
    Deno.exit(-1);
  }
  
  const opts = args.reduce((acc) => acc, []);
  const file = await Deno.open(path);

  const postDataResult = await fetch(`${config.host}/data`, {
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