export type CaosAddr = string;
export type CaosData = ReadableStream<Uint8Array>;
export type CaosTagKey = string;
export type CaosTagValue = string;

export type CaosTags = {[tag: CaosTagKey]: CaosTagValue};

export type CaosRefType = string;
export type CaosRefs = Array<{
  ref: CaosRefType;
  to: CaosAddr;
}>;

export type Caos = {
  addr: {
    all: (addr: CaosAddr) => CaosAddr[];
    has: (addr: CaosAddr) => boolean;
  };
  data: {
    add: (data: CaosData) => Promise<CaosAddr>;
    del: (addr: CaosAddr) => Promise<void>;
    has: (addr: CaosAddr) => Promise<boolean>;
    get: (addr: CaosAddr) => Promise<CaosData | undefined>;
  };
  refs: {
    all: (addr: CaosAddr) => CaosRefs;
    add: (addr: CaosAddr, ref: CaosRefType, to: CaosAddr) => void;
    del: (addr: CaosAddr, ref: CaosRefType, to: CaosAddr) => void;
    get: (addr: CaosAddr, ref: CaosRefType) => CaosAddr[];
  };
  tags: {
    all: (addr: CaosAddr) => CaosTags;
    del: (addr: CaosAddr, tag: CaosTagKey) => void;
    get: (addr: CaosAddr, tag: CaosTagKey) => CaosTagValue | undefined;
    set: (addr: CaosAddr, tag: CaosTagKey, value: CaosTagValue) => void;
  };
}
