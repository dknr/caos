import { concat } from "https://deno.land/std@0.183.0/bytes/concat.ts";
import {WASMagic} from 'npm:wasmagic';
const magic = await WASMagic.create();

export class StreamTyper extends TransformStream<Uint8Array, Uint8Array> {
  private MAX_MAGIC = 1024;
  private chunks: Uint8Array[] = [];
  private length = 0;
  type: string | null = null;

  constructor() {
    super({
      transform: (chunk, controller) => {
        controller.enqueue(chunk);
        if (this.type == null) {
          if (this.length < this.MAX_MAGIC) {
            this.length += chunk.length;
            this.chunks.push(chunk);
          }
          if (this.length >= this.MAX_MAGIC) {
            const bytes = concat(...this.chunks);
            this.type = magic.getMime(bytes);
          }
        }
      }
    });
  }
}