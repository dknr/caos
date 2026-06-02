import { concat } from "https://deno.land/std@0.192.0/bytes/concat.ts";

export class StreamTyper extends TransformStream<Uint8Array, Uint8Array> {
  private static magic: any = null;
  private static async getMagic() {
    if (!StreamTyper.magic) {
      const { WASMagic } = await import("npm:wasmagic");
      StreamTyper.magic = await WASMagic.create();
    }
    return StreamTyper.magic;
  }

  private MAX_MAGIC = 1024;
  private chunks: Uint8Array[] = [];
  private length = 0;

  async getType() {
    const magic = await StreamTyper.getMagic();
    const bytes = concat(...this.chunks);
    return magic.getMime(bytes);
  }

  constructor() {
    super({
      transform: (chunk, controller) => {
        controller.enqueue(chunk);
        if (this.length < this.MAX_MAGIC) {
          this.length += chunk.length;
          this.chunks.push(chunk);
        }
      },
    });
  }
}
