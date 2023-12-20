import {CmdFn} from "../cmd.ts";
import {loadOpts} from "../opts.ts";
import {buildClient} from "../client/mod.ts";
import {CaosAddr} from "../types.ts";
import {readAll} from "https://deno.land/std@0.192.0/streams/read_all.ts";
import {readerFromStreamReader} from "https://deno.land/std@0.192.0/streams/reader_from_stream_reader.ts";

const u8d = new TextDecoder();

const pull: CmdFn = async (args) => {
  // TODO: pass opts in to CmdFn
  const opts = loadOpts();
  const client = buildClient({host: opts.host});

  if (args.length == 0) {
    console.log('usage: caos pull [addr]');
    return;
  }

  let failed = false;
  const addrs = [];
  for (const addr of args) {
    const matches = await client.addr.all(addr);
    switch (matches.length) {
      case 0:
        failed = true;
        console.log(`${addr}: unknown`);
        break;
      case 1: {
        const match = matches[0];
        addrs.push(match);
        break;
      }
      default:
        failed = true;
        console.log(`${addr}: unambiguous`);
        break;
    }
  }

  if (failed) return;

  const paths = new Map<string, CaosAddr>();
  for (const addr of addrs) {
    const type = await client.tags.get(addr, 'type');
    if (type !== 'caos/path') {
      failed = true;
      console.log(`${addr}: invalid type: ${type}`);
      continue;
    }

    const data = await client.data.get(addr);
    if (!data) {
      throw new Error(`${addr}: missing data`);
    }

    const pathBytes = await readAll(readerFromStreamReader(data.getReader()));
    const text = u8d.decode(pathBytes);
    const lines = text.split('\n');
    for (let line = 0; line < lines.length; line++) {
      const parts = lines[line].split(' ', 2);
      if (parts.length !== 2) {
        failed = true;
        console.log(`${addr}: ${line}: bad path`);
        continue;
      }

      const [path_addr, path_name] = parts;
      if (paths.has(path_name)) {
        failed = true;
        console.log(`${addr}: ${line}: path collision: ${path_name}`);
        continue;
      }

      paths.set(path_name, path_addr);
    }
  }

  if (failed) return;

  const directories = new Set<string>();
  for (const name of paths.keys()) {
    if (name.startsWith('/')) {
      throw new Error('refusing to pull to an absolute path');
    }
    if (name.includes('..')) {
      throw new Error('refusing to pull to a relative path');
    }
    const directory = name.split('/').slice(0, -1).join('/');
    directories.add(directory);
  }

  for (const directory of directories.keys()) {
    console.log(directory);
    Deno.mkdirSync(directory, {recursive: true});
  }

  for (const [path_name, path_addr] of paths.entries()) {
    const data = await client.data.get(path_addr);
    if (!data) {
      throw new Error('missing data');
    }

    const file = await Deno.open(path_name, {create: true, write: true});
    await data.pipeTo(file.writable);
    console.log(`${path_addr.slice(0, 8)}: ${path_name}`);
  }
}

export default pull;
