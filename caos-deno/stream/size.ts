export class StreamSizer extends TransformStream<Uint8Array, Uint8Array> {
  size = 0;

  constructor() {
    super({
      transform: (chunk, controller) => {
        controller.enqueue(chunk);
        this.size += chunk.byteLength;
      }
    })
  }
}