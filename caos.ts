import {Application, Router} from 'https://deno.land/x/oak@v12.1.0/mod.ts';
import {openCaos} from "./store/mod.ts";

const app = new Application();
const router = new Router();
const caos = openCaos({
  path: '/tmp/caos',
});

router.get('/stat', async (ctx) => {
  ctx.response.status = 200;
})

router.post('/data', async (ctx) => {
  const data = ctx.request.body({type: 'stream'}).value;
  const result = await caos.addData(data);
  ctx.response.body = result;
});
router.post('/data/:file', async (ctx) => {
  const data = ctx.request.body({type: 'stream'}).value;
  const result = await caos.addData(data);
  ctx.response.body = result;
});


router.get('/data/:addr', async (ctx) => {
  const addr = ctx.params.addr;
  const data = await caos.getData(addr);
  const type = caos.getTag(addr, 'type');
  if (data) {
    ctx.response.headers.set('content-type', type || 'application/octet-stream');
    ctx.response.body = data;
  } else {
    ctx.response.status = 404;
  }
});

router.get('/tags/:addr', (ctx) => {
  const addr = ctx.params.addr;
  const tags = caos.getTags(addr);
  ctx.response.body = tags;
});

router.get('/tags/:addr/:tag', (ctx) => {
  const {addr, tag} = ctx.params;
  const value = caos.getTag(addr, tag);

  if (value) {
    ctx.response.body = value;
  } else {
    ctx.response.status = 404;
  }
})

app.use(router.routes());

app.addEventListener('listen', (e) => console.log(`serving caos at http://localhost:${e.port}`));
app.listen({port: 31923});
