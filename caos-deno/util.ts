import {CaosClient} from "./client/mod.ts";
import {CaosAddr} from "./types.ts";

export const assertArgsCount = (args: string[], min: number, max?: number) => {
  if (args.length < min) {
    console.error('insufficent arguments');
    Deno.exit(-1);
  }
  if (max && args.length > max) {
    console.error('excessive arguments');
    Deno.exit(-1);
  }
}

export const resolveAddress = async (client: CaosClient, addr: CaosAddr) => {
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