import { CaosConfig } from "./config.ts";
import * as path from 'https://deno.land/std@0.192.0/path/mod.ts';
import {DB} from 'https://deno.land/x/sqlite@v3.9.1/mod.ts';
import {CaosAddr, CaosTagKey} from "../types.ts";
import { CaosOpts } from "../opts.ts";

const ADDR_REGEX = /^[0-9a-fA-F]{64}$/;

const prepareCaosMetaDb = (db: DB) => {

  db.execute(`
create table if not exists objs (
  addr text primary key
    on conflict ignore
);
`);

  db.execute(`
create table if not exists tags (
  addr text not null,
  tag text not null,
  value text not null,
  constraint pk_tags primary key (tag, addr)
    on conflict replace,
  constraint fk_tags_objs foreign key (addr)
    references objs (addr)
    on delete cascade
    on update cascade
);
`);

  db.execute(`
create table if not exists names (
  name text primary key,
  addr text not null
    on conflict replace
);
`);

  db.execute(`
create table if not exists refs (
  ref text not null,
  src text not null,
  dst text not null,
  constraint uq_refs unique (ref, src, dst)
    on conflict ignore,
  constraint fk_refs_src foreign key (src)
    references objs
    on delete cascade
    on update cascade,
  constraint fk_refs_dst foreign key (dst)
    references objs
    on delete cascade
    on update cascade
);
`);
};

export const openCaosMeta = (opts: CaosOpts) => {
  const metaPath = path.join(opts.root, 'meta');
  Deno.mkdirSync(metaPath, {recursive: true});

  const db = new DB(path.join(metaPath, 'caos.db'));
  prepareCaosMetaDb(db);

  return {
    addAddr: (addr: CaosAddr) => {
      if (!ADDR_REGEX.test(addr)) {
        throw new Error(`invalid address format: ${addr}`);
      }
      db.query(
        `insert into objs (addr) values (?)`,
        [addr]
      );
    },

    countAddrs: () => {
      const result = db.query<[number]>(
        `select count(addr) from objs`
      ).map(([count]) => count);
      return result[0] ?? 0;
    },

    getAddrs: (partial: CaosAddr) => db.query<[string]>(
      `select addr from objs where addr like ?`,
      [partial + '%']
    ).map(([addr]) => addr),

    hasAddr: (addr: CaosAddr) => db.query(
      `select addr from objs where addr = ?`,
      [addr]
    ).length > 0,

    setTag: (addr: CaosAddr, tag: string, value: string) => db.query(
      `insert into tags (addr, tag, value) values (?,?,?)`,
      [addr, tag, value]
    ),

    getTag: (addr: CaosAddr, tag: string) => db.query<[string]>(
      `select value from tags where addr = ? and tag = ?`,
      [addr, tag]
    ).map((row) => row[0])[0],

    getTags: (addr: CaosAddr) => db.query(
      `select tag, value from tags where addr = ?`,
      [addr]
    ).map((row) => (
      {tag: String(row[0]), value: String(row[1])}
    )).reduce(
      (result, {tag, value}) => ({...result, [tag]: value}), {}
    ),

    delTag: (addr: CaosAddr, tag: CaosTagKey) => db.query(
      `delete from tags where addr = ? and tag = ?`,
      [addr, tag],
    ),

    setName: (name: string, addr: CaosAddr) => db.query(
      `insert into names (name, addr) values (?,?)`,
      [name, addr],
    ),

    getName: (name: string) => db.query<[string]>(
      `select addr from names where name = ?`,
      [name]
    ).map((row) => row[0])[0],
  }
}
