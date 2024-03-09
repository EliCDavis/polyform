
export class BinaryWriter {

    constructor(littleEndian) {
        this.littleEndian = littleEndian;
        this.arrayBuffer = new ArrayBuffer(0, {
            maxByteLength: 1024 * 1024 * 1024
        });
        this.dataView = new DataView(this.arrayBuffer);
        this.offset = 0;
    }

    float64(f) {
        this.arrayBuffer.resize(this.offset + 8);
        this.dataView.setFloat64(this.offset, f, this.littleEndian);
        this.offset += 8;
    }

    float32(f) {
        this.arrayBuffer.resize(this.offset + 4);
        this.dataView.setFloat32(this.offset, f, this.littleEndian);
        this.offset += 4;
    }

    byte(b) {
        this.arrayBuffer.resize(this.offset + 1);
        this.dataView.setUint8(this.offset, b, this.littleEndian);
        this.offset += 1;
    }

    bool(b) {
        this.arrayBuffer.resize(this.offset + 1);
        let v = 0;
        if (b) {
            v = 1;
        }
        this.dataView.setUint8(this.offset, v, this.littleEndian);
        this.offset += 1;
    }

    buffer() {
        const buf = new ArrayBuffer(this.arrayBuffer.byteLength);
        new Uint8Array(buf).set(new Uint8Array(this.arrayBuffer));
        return buf;
    }
}