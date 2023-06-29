import {Application, Router} from 'https://deno.land/x/oak@v12.1.0/mod.ts';
import {DB} from 'https://deno.land/x/sqlite@v3.7.2/mod.ts';
import * as path from 'https://deno.land/std@0.184.0/path/mod.ts';
import {WASMagic} from 'npm:wasmagic';
import {createSHA256} from 'https://deno.land/x/hashwasm@v4.1.0-deno2/lib/sha256.ts';
import { Caos } from "./types.ts";
import { StreamHasher } from './hash.ts';

type CaosConfig = {
  path: string;
}

const openCaosDb = (config: CaosConfig) => {
  const db = new DB(path.join(config.path, 'caos.db'));
  db.execute(`create table if not exists addrs (
    addr text primary key
  )`);

  db.execute(`create table if not exists tags (
    addr text not null,
    tag text not null,
    value text not null,
    constraint tags_pk primary key (addr, tag)
      on conflict replace,
    constraint tags_fk_addrs foreign key (addr)
      references addrs (addr)
      on update cascade
      on delete cascade
  )`);

  db.execute(`create table if not exists refs (
    addr text not null,
    ref text not null,
    to text not null,
    constraint refs_pk primary key (addr, ref, to)
      on conflict replace,
    constraint refs_fk_addr_addrs foreign key (addr)
      references objs (addr)
      on update cascade
      on delete cascade,
    constraint refs_fk_to_addrs foreign key (to)
      references objs (addr)
      on update cascade
      on delete cascade
  )`);

  return {

  };
}



const openCaos = async (config: CaosConfig): Promise<Caos> => {
  const magic = await WASMagic.create();
  
  return {
    addData: (data) => {
      const hasher = new StreamHasher("SHA-256");
      const tempFile = Deno.open()
      data.pipeThrough(hasher).pipeTo();
    }
  }
}

const serveCaos = (caos: Caos) => {
  const app = new Application();
  const router = new Router();

  router.get('/', (ctx) => {
    ctx.response.status = 301;
    ctx.response.headers.set('location', '/d10b49b4cf4f9204c4a6e4a96e5a004fa25768623667b2aec05f82e4852aaa91');
  });

  router.get('/:id', (ctx) => {
    const id = ctx.params.id;
  });

  router.post('/data', async (ctx) => {
    const data = ctx.request.body({type: 'stream'}).value;
    const addr = await caos.addData(data);
    ctx.response.body = addr;
  });

  return app.listen({port: 31923});
}
