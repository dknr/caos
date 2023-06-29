import { CmdFn } from "../cmd.ts";
import { CaosAddr, CaosTagKey, CaosTagValue, CaosTags } from "../types.ts";

type ClientConfig = {
  host: string;
}
type CaosClient = {
  addr: {
    all: (addr: CaosAddr) => Promise<CaosAddr[]>;
  },
  tags: {
    all: (addr: CaosAddr) => Promise<CaosTags>;
    get: (addr: CaosAddr, tag: CaosTagKey) => Promise<CaosTagValue | undefined>;
    set: (addr: CaosAddr, tag: CaosTagKey, value: CaosTagValue) => Promise<void>;
    del: (addr: CaosAddr, tag: CaosTagKey) => Promise<void>;
  },
}

const buildClient = ({host}: ClientConfig): CaosClient => ({
  addr: {
    all: async (addr) => {
      const res = await fetch(`${host}/addr/${addr}`);
      return await res.json();
    }
  },
  tags: {
    all: async (addr) => {
      const res = await fetch(`${host}/tags/${addr}`);
      if (res.ok) {
        return await res.json();
      } else {
        return {};
      }
    },
    get: async (addr, tag) => {
      const res = await fetch(`${host}/tags/${addr}/${tag}`);
      if (res.ok) {
        return await res.text();
      }
    },
    set: async (addr, tag, value) => {
      const res = await fetch(`${host}/tags/${addr}/${tag}`, {method: 'put', body: value});
      if (!res.ok) {
        throw new Error(`failed request: ${res.status} ${res.statusText}`);
      }
    },
    del: async (addr, tag) => {
      await fetch(`${host}/tags/${addr}/${tag}`, {method: 'delete'});
    }
  },
});

const assertArgsCount = (args: string[], min: number, max?: number) => {
  if (args.length < min) {
    console.error('insufficent arguments');
    Deno.exit(-1);
  }
  if (max && args.length > max) {
    console.error('excessive arguments');
    Deno.exit(-1);
  }
}

const resolveAddress = async (client: CaosClient, addr: CaosAddr) => {
  const addrs = await client.addr.all(addr);
  if (addrs.length > 1) {
    console.error('address resolution returned multiple results:')
    addrs.forEach(console.log);
    Deno.exit(-1);
  }
  if (addrs.length < 1) {
    console.error('address resolution returned zero results.');
    Deno.exit(-1);
  }
  return addrs[0];
}

const get = async (client: CaosClient, args: string[]) => {
  assertArgsCount(args, 1);

  const addr = await resolveAddress(client, args[0]);
  const tag = args[1];
  if (tag) {
    const value = await client.tags.get(addr, tag);
    console.log(value);
  } else {
    const tags = await client.tags.all(addr);
    Object.entries(tags).forEach(([tag, value]) => console.log(`${tag}: ${value}`));
  }
}

const set = async (client: CaosClient, args: string[]) => {
  assertArgsCount(args, 3);

  const addr = await resolveAddress(client, args[0]);
  const tag = args[1];
  const value = args[2];

  client.tags.set(addr, tag, value);
}

const del = async (client: CaosClient, args: string[]) => {
  assertArgsCount(args, 2);

  const addr = await resolveAddress(client, args[0]);
  const tag = args[1];

  await client.tags.del(addr, tag);
}

const help = async () => {
  console.error('caos tag get <addr> [tag]');
  console.error('caos tag set <addr> <tag> <value>');
  console.error('caos tag del <addr> <tag>');
  console.error('caos tag help');
  Deno.exit(-1);
}

const ops: Record<string, (client: CaosClient, args: string[]) => void | Promise<void>> = {
  get,
  set,
  del,
  help
};

const tag: CmdFn = async (args) => {
  const client = buildClient({host: 'http://localhost:31923'});
  await (ops[args[0]] || ops.help)(client, args.slice(1));
}

export default tag;
