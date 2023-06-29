export type CaosAddr = string;
export type CaosData = ReadableStream<Uint8Array>;
export type CaosTagKey = string;
export type CaosTagValue = string;

export type CaosTags = {[tag: string]: string};

export type CaosRef = {
  ref: string;
  to: string;
}
export type CaosRefs = Array<CaosRef>;

export type Caos = {
  addData: (data: CaosData) => Promise<CaosAddr>;
  setTag: (addr: CaosAddr, tag: CaosTagKey, value: CaosTagValue) => void;
  addRef: (addr: CaosAddr, ref: CaosRef) => Promise<void>;
  getData: (addr: CaosAddr) => Promise<CaosData | undefined>;
  getTags: (addr: CaosAddr) => CaosTags;
  getTag: (addr: CaosAddr, tag: CaosTagKey) => CaosTagValue | undefined;
  getRefs: (addr: CaosAddr) => CaosRefs;
  hasData: (addr: CaosAddr) => Promise<boolean>;
  delete: (addr: CaosAddr) => Promise<void>;
}
