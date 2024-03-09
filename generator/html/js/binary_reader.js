const textDecoder = new TextDecoder();


export class BinaryReader {

    constructor(dataView) {
        this.dataView = dataView;
        this.currentOffset = 0;
    }

    RemainingLength() {
        return this.dataView.byteLength - this.currentOffset;
    }

    StringWithLength() {
        const stringLength = this.Byte();

        const str = textDecoder.decode(
            new DataView(
                this.dataView.buffer,
                this.currentOffset + this.dataView.byteOffset,
                stringLength
            )
        );

        this.currentOffset += stringLength;
        return str;
    }

    String(stringLength) {
        const str = textDecoder.decode(
            new DataView(
                this.dataView.buffer,
                this.currentOffset + this.dataView.byteOffset,
                stringLength
            )
        );

        this.currentOffset += stringLength;
        return str;
    }

    Int32() {
        const val = this.dataView.getInt32(this.currentOffset, true);
        this.currentOffset += 4;
        return val;
    }

    UInt32() {
        const val = this.dataView.getUint32(this.currentOffset, true);
        this.currentOffset += 4;
        return val;
    }

    Byte() {
        const propByte = this.dataView.getUint8(this.currentOffset);
        this.currentOffset += 1;
        return propByte;
    }

    Bool() {
        const propByte = this.dataView.getUint8(this.currentOffset);
        this.currentOffset += 1;
        return propByte == 1;
    }

    Bytes(numBytes) {
        const arr = new Uint8Array(
            this.dataView.buffer,
            this.currentOffset + this.dataView.byteOffset,
            numBytes
        );
        this.currentOffset += numBytes;
        return arr;
    }

    Binary(numBytes) {
        this.currentOffset += numBytes;
        return new DataView(
            this.dataView.buffer,
            this.dataView.byteOffset + this.currentOffset - numBytes,
            numBytes
        );
    }

    Float32() {
        const val = this.dataView.getFloat32(this.currentOffset, true);
        this.currentOffset += 4;
        return val;
    }

    Float64() {
        const val = this.dataView.getFloat64(this.currentOffset, true);
        this.currentOffset += 8;
        return val;
    }

    UVarIntArray() {
        const numInts = this.UVarInt();
        const arr = new Array < number > (numInts);
        for (let i = 0; i < numInts; i++) {
            arr[i] = this.UVarInt();
        }
        return arr;
    }

    UVarInt() {
        let x = 0;
        let s = 0;

        for (let i = 0; i < MaxVarintLen64; i++) {
            const b = this.dataView.getUint8(this.currentOffset + i);

            if (b < 0x80) {
                if (i == 9 && b > 1) {
                    throw OverflowError;
                }
                this.currentOffset += i + 1;
                return x | (b << s);
            }
            x |= (b & 0x7f) << s;
            s += 7;
        }
        throw OverflowError;
    }

    BinaryArray() {
        const length = this.UVarInt();
        this.currentOffset += length;
        return new DataView(
            this.dataView.buffer,
            this.dataView.byteOffset + this.currentOffset - length,
            length
        );
    }

    StringArray() {
        const length = this.UVarInt();
        const arr = new Array < string > (length);
        for (let i = 0; i < length; i++) {
            arr[i] = this.String();
        }
        return arr;
    }
}
