import { CaosAddr, CaosData, CaosTagKey, CaosTags, CaosTagValue } from "../types.ts";

export type ClientConfig = {
  host: string;
};

export type CaosClient = {
  addr: {
    all: (addr: CaosAddr) => Promise<CaosAddr[]>;
  };
  data: {
    add: (data: BodyInit) => Promise<CaosAddr>;
    get: (addr: CaosAddr) => Promise<CaosData | undefined>;
  }
  tags: {
    all: (addr: CaosAddr) => Promise<CaosTags>;
    get: (addr: CaosAddr, tag: CaosTagKey) => Promise<CaosTagValue | undefined>;
    set: (
      addr: CaosAddr,
      tag: CaosTagKey,
      value: CaosTagValue,
    ) => Promise<void>;
    del: (addr: CaosAddr, tag: CaosTagKey) => Promise<void>;
  };
};

export const buildClient = ({ host }: ClientConfig): CaosClient => ({
  addr: {
    all: async (addr) => {
      const res = await fetch(`${host}/addr/${addr}`);
      return await res.json();
    },
  },
  data: {
    add: async (data) => {
      const result = await fetch(`${host}/data`, {
        method: 'post',
        body: data,
      });
      const addr = await result.text();
      return addr;
    },
    get: async (addr) => {
      const result = await fetch(`${host}/data/${addr}`);
      if (result.ok)
        return result.body!;
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
      const res = await fetch(`${host}/tags/${addr}/${tag}`, {
        method: "put",
        body: value,
      });
      if (!res.ok) {
        throw new Error(`failed request: ${res.status} ${res.statusText}`);
      }
    },
    del: async (addr, tag) => {
      await fetch(`${host}/tags/${addr}/${tag}`, { method: "delete" });
    },
  },
});
