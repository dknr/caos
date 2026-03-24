import { instantiateWasm, DigestContext } from 'https://deno.land/std@0.192.0/crypto/_wasm/mod.ts';
import {encode as encodeHex} from 'https://deno.land/std@0.192.0/encoding/hex.ts';

const wasmCrypto = instantiateWasm();
const textDecoder = new TextDecoder();

export class StreamHasher extends TransformStream<Uint8Array, Uint8Array> {
  context: DigestContext;
  constructor(algorithm = "SHA-256") {
    super({
      transform: (chunk, controller) => {
        controller.enqueue(chunk);
        this.context.update(chunk);
      }
    });
    this.context = new wasmCrypto.DigestContext(algorithm);
  }

  digest() {
    const result = this.context.digestAndDrop(undefined);
    const hexBytes = encodeHex(result);
    return textDecoder.decode(hexBytes);
  }
}
