import * as ascii85 from 'https://deno.land/std@0.192.0/encoding/ascii85.ts';
import * as hex from 'https://deno.land/std@0.192.0/encoding/hex.ts';
import { StreamHasher } from "./hash.ts";

const input = await Deno.open('/home/dknr/Downloads/ewu.jpg', {read: true});

const outputPath = await Deno.makeTempFile();
const output = await Deno.open(outputPath, {create: true, write: true});

const hasher = new StreamHasher("SHA-256");
await input.readable.pipeThrough(hasher).pipeTo(output.writable);

const textDecoder = new TextDecoder();
const hash = hasher.digest();
// console.log(ascii85.encode(hash));
console.log(textDecoder.decode(hex.encode(hash)));

await Deno.remove(outputPath);
