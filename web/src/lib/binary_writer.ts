
export class BinaryWriter {

    littleEndian: boolean;
    offset: number;
    dataView: DataView;
    arrayBuffer: ArrayBuffer;


    constructor(littleEndian: boolean) {
        this.littleEndian = littleEndian;
        this.arrayBuffer = new ArrayBuffer(0, {
            maxByteLength: 5 * 1024 * 1024
        });
        this.dataView = new DataView(this.arrayBuffer);
        this.offset = 0;
    }

    float64(f: number) {
        this.arrayBuffer.resize(this.offset + 8);
        this.dataView.setFloat64(this.offset, f, this.littleEndian);
        this.offset += 8;
    }

    float32(f: number) {
        this.arrayBuffer.resize(this.offset + 4);
        this.dataView.setFloat32(this.offset, f, this.littleEndian);
        this.offset += 4;
    }

    byte(b: number) {
        this.arrayBuffer.resize(this.offset + 1);
        this.dataView.setUint8(this.offset, b);
        this.offset += 1;
    }

    bool(b: boolean) {
        this.arrayBuffer.resize(this.offset + 1);
        let v = 0;
        if (b) {
            v = 1;
        }
        this.dataView.setUint8(this.offset, v);
        this.offset += 1;
    }

    buffer() {
        const buf = new ArrayBuffer(this.arrayBuffer.byteLength);
        new Uint8Array(buf).set(new Uint8Array(this.arrayBuffer));
        return buf;
    }
}