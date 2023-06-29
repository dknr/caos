import { CaosAddr, CaosTagKey, CaosTags, CaosTagValue } from "../types.ts";

export type ClientConfig = {
  host: string;
};

export type CaosClient = {
  addr: {
    all: (addr: CaosAddr) => Promise<CaosAddr[]>;
  };
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
