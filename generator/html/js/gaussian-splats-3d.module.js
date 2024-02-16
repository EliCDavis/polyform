import * as THREE from 'three';
import { Ray as Ray$1, Plane, MathUtils, EventDispatcher, Vector3, MOUSE, TOUCH, Quaternion, Spherical, Vector2 } from 'three';

/**
 * SplatBuffer: Container for splat data from a single scene/file and capable of (mediocre) compression.
 */
class SplatBuffer {

    static CenterComponentCount = 3;
    static ScaleComponentCount = 3;
    static RotationComponentCount = 4;
    static ColorComponentCount = 4;

    static CompressionLevels = {
        0: {
            BytesPerCenter: 12,
            BytesPerScale: 12,
            BytesPerColor: 4,
            BytesPerRotation: 16,
            ScaleRange: 1
        },
        1: {
            BytesPerCenter: 6,
            BytesPerScale: 6,
            BytesPerColor: 4,
            BytesPerRotation: 8,
            ScaleRange: 32767
        }
    };

    static CovarianceSizeFloats = 6;
    static CovarianceSizeBytes = 24;

    static HeaderSizeBytes = 1024;

    constructor(bufferData) {
        this.headerBufferData = new ArrayBuffer(SplatBuffer.HeaderSizeBytes);
        this.headerArrayUint8 = new Uint8Array(this.headerBufferData);
        this.headerArrayUint32 = new Uint32Array(this.headerBufferData);
        this.headerArrayFloat32 = new Float32Array(this.headerBufferData);
        this.headerArrayUint8.set(new Uint8Array(bufferData, 0, SplatBuffer.HeaderSizeBytes));
        this.versionMajor = this.headerArrayUint8[0];
        this.versionMinor = this.headerArrayUint8[1];
        this.headerExtraK = this.headerArrayUint8[2];
        this.compressionLevel = this.headerArrayUint8[3];
        this.splatCount = this.headerArrayUint32[1];
        this.bucketSize = this.headerArrayUint32[2];
        this.bucketCount = this.headerArrayUint32[3];
        this.bucketBlockSize = this.headerArrayFloat32[4];
        this.halfBucketBlockSize = this.bucketBlockSize / 2.0;
        this.bytesPerBucket = this.headerArrayUint32[5];
        this.compressionScaleRange = this.headerArrayUint32[6] || SplatBuffer.CompressionLevels[this.compressionLevel].ScaleRange;
        this.compressionScaleFactor = this.halfBucketBlockSize / this.compressionScaleRange;

        const dataBufferSizeBytes = bufferData.byteLength - SplatBuffer.HeaderSizeBytes;
        this.splatBufferData = new ArrayBuffer(dataBufferSizeBytes);
        new Uint8Array(this.splatBufferData).set(new Uint8Array(bufferData, SplatBuffer.HeaderSizeBytes, dataBufferSizeBytes));

        this.bytesPerCenter = SplatBuffer.CompressionLevels[this.compressionLevel].BytesPerCenter;
        this.bytesPerScale = SplatBuffer.CompressionLevels[this.compressionLevel].BytesPerScale;
        this.bytesPerColor = SplatBuffer.CompressionLevels[this.compressionLevel].BytesPerColor;
        this.bytesPerRotation = SplatBuffer.CompressionLevels[this.compressionLevel].BytesPerRotation;

        this.bytesPerSplat = this.bytesPerCenter + this.bytesPerScale + this.bytesPerColor + this.bytesPerRotation;

        this.linkBufferArrays();
    }

    linkBufferArrays() {
        let FloatArray = (this.compressionLevel === 0) ? Float32Array : Uint16Array;
        this.centerArray = new FloatArray(this.splatBufferData, 0, this.splatCount * SplatBuffer.CenterComponentCount);
        this.scaleArray = new FloatArray(this.splatBufferData, this.bytesPerCenter * this.splatCount,
                                         this.splatCount * SplatBuffer.ScaleComponentCount);
        this.colorArray = new Uint8Array(this.splatBufferData, (this.bytesPerCenter + this.bytesPerScale) * this.splatCount,
                                         this.splatCount * SplatBuffer.ColorComponentCount);
        this.rotationArray = new FloatArray(this.splatBufferData,
                                             (this.bytesPerCenter + this.bytesPerScale + this.bytesPerColor) * this.splatCount,
                                              this.splatCount * SplatBuffer.RotationComponentCount);
        this.bucketsBase = this.splatCount * this.bytesPerSplat;
    }

    fbf(f) {
        if (this.compressionLevel === 0) {
            return f;
        } else {
            return THREE.DataUtils.fromHalfFloat(f);
        }
    };

    getHeaderBufferData() {
        return this.headerBufferData;
    }

    getSplatBufferData() {
        return this.splatBufferData;
    }

    getSplatCount() {
        return this.splatCount;
    }

    getSplatCenter(index, outCenter, transform) {
        let bucket = [0, 0, 0];
        const centerBase = index * SplatBuffer.CenterComponentCount;
        if (this.compressionLevel > 0) {
            const sf = this.compressionScaleFactor;
            const sr = this.compressionScaleRange;
            const bucketIndex = Math.floor(index / this.bucketSize);
            bucket = new Float32Array(this.splatBufferData, this.bucketsBase + bucketIndex * this.bytesPerBucket, 3);
            outCenter.x = (this.centerArray[centerBase] - sr) * sf + bucket[0];
            outCenter.y = (this.centerArray[centerBase + 1] - sr) * sf + bucket[1];
            outCenter.z = (this.centerArray[centerBase + 2] - sr) * sf + bucket[2];
        } else {
            outCenter.x = this.centerArray[centerBase];
            outCenter.y = this.centerArray[centerBase + 1];
            outCenter.z = this.centerArray[centerBase + 2];
        }
        if (transform) outCenter.applyMatrix4(transform);
    }

    getSplatScaleAndRotation = function() {

        const scaleMatrix = new THREE.Matrix4();
        const rotationMatrix = new THREE.Matrix4();
        const tempMatrix = new THREE.Matrix4();
        const tempPosition = new THREE.Vector3();

        return function(index, outScale, outRotation, transform) {
            const scaleBase = index * SplatBuffer.ScaleComponentCount;
            outScale.set(this.fbf(this.scaleArray[scaleBase]),
                         this.fbf(this.scaleArray[scaleBase + 1]),
                         this.fbf(this.scaleArray[scaleBase + 2]));
            const rotationBase = index * SplatBuffer.RotationComponentCount;
            outRotation.set(this.fbf(this.rotationArray[rotationBase + 1]), this.fbf(this.rotationArray[rotationBase + 2]),
                            this.fbf(this.rotationArray[rotationBase + 3]), this.fbf(this.rotationArray[rotationBase]));
            if (transform) {
                scaleMatrix.makeScale(outScale.x, outScale.y, outScale.z);
                rotationMatrix.makeRotationFromQuaternion(outRotation);
                tempMatrix.copy(scaleMatrix).multiply(rotationMatrix).multiply(transform);
                tempMatrix.decompose(tempPosition, outRotation, outScale);
            }
        };

    }();

    getSplatColor(index, outColor, transform) {
        const colorBase = index * SplatBuffer.ColorComponentCount;
        outColor.set(this.colorArray[colorBase], this.colorArray[colorBase + 1],
                     this.colorArray[colorBase + 2], this.colorArray[colorBase + 3]);
        // TODO: apply transform for spherical harmonics
    }

    fillSplatCenterArray(outCenterArray, destOffset, transform) {
        const splatCount = this.splatCount;
        let bucket = [0, 0, 0];
        const center = new THREE.Vector3();
        for (let i = 0; i < splatCount; i++) {
            const centerSrcBase = i * SplatBuffer.CenterComponentCount;
            const centerDestBase = (i + destOffset) * SplatBuffer.CenterComponentCount;
            if (this.compressionLevel > 0) {
                const bucketIndex = Math.floor(i / this.bucketSize);
                bucket = new Float32Array(this.splatBufferData, this.bucketsBase + bucketIndex * this.bytesPerBucket, 3);
                const sf = this.compressionScaleFactor;
                const sr = this.compressionScaleRange;
                center.x = (this.centerArray[centerSrcBase] - sr) * sf + bucket[0];
                center.y = (this.centerArray[centerSrcBase + 1] - sr) * sf + bucket[1];
                center.z = (this.centerArray[centerSrcBase + 2] - sr) * sf + bucket[2];
            } else {
                center.x = this.centerArray[centerSrcBase];
                center.y = this.centerArray[centerSrcBase + 1];
                center.z = this.centerArray[centerSrcBase + 2];
            }
            if (transform) {
                center.applyMatrix4(transform);
            }
            outCenterArray[centerDestBase] = center.x;
            outCenterArray[centerDestBase + 1] = center.y;
            outCenterArray[centerDestBase + 2] = center.z;
        }
    }

    fillSplatCovarianceArray(covarianceArray, destOffset, transform) {
        const splatCount = this.splatCount;

        const scale = new THREE.Vector3();
        const rotation = new THREE.Quaternion();
        const rotationMatrix = new THREE.Matrix3();
        const scaleMatrix = new THREE.Matrix3();
        const covarianceMatrix = new THREE.Matrix3();
        const transformedCovariance = new THREE.Matrix3();
        const transform3x3 = new THREE.Matrix3();
        const transform3x3Transpose = new THREE.Matrix3();
        const tempMatrix4 = new THREE.Matrix4();

        for (let i = 0; i < splatCount; i++) {
            const scaleBase = i * SplatBuffer.ScaleComponentCount;
            scale.set(this.fbf(this.scaleArray[scaleBase]),
                      this.fbf(this.scaleArray[scaleBase + 1]),
                      this.fbf(this.scaleArray[scaleBase + 2]));
            tempMatrix4.makeScale(scale.x, scale.y, scale.z);
            scaleMatrix.setFromMatrix4(tempMatrix4);

            const rotationBase = i * SplatBuffer.RotationComponentCount;
            rotation.set(this.fbf(this.rotationArray[rotationBase + 1]),
                         this.fbf(this.rotationArray[rotationBase + 2]),
                         this.fbf(this.rotationArray[rotationBase + 3]),
                         this.fbf(this.rotationArray[rotationBase]));
            tempMatrix4.makeRotationFromQuaternion(rotation);
            rotationMatrix.setFromMatrix4(tempMatrix4);

            covarianceMatrix.copy(rotationMatrix).multiply(scaleMatrix);
            transformedCovariance.copy(covarianceMatrix).transpose().premultiply(covarianceMatrix);
            const covBase = SplatBuffer.CovarianceSizeFloats * (i + destOffset);

            if (transform) {
                transform3x3.setFromMatrix4(transform);
                transform3x3Transpose.copy(transform3x3).transpose();
                transformedCovariance.multiply(transform3x3Transpose);
                transformedCovariance.premultiply(transform3x3);
            }

            covarianceArray[covBase] = transformedCovariance.elements[0];
            covarianceArray[covBase + 1] = transformedCovariance.elements[3];
            covarianceArray[covBase + 2] = transformedCovariance.elements[6];
            covarianceArray[covBase + 3] = transformedCovariance.elements[4];
            covarianceArray[covBase + 4] = transformedCovariance.elements[7];
            covarianceArray[covBase + 5] = transformedCovariance.elements[8];
        }
    }

    fillSplatColorArray(outColorArray, destOffset, transform) {
        const splatCount = this.splatCount;
        for (let i = 0; i < splatCount; i++) {
            const colorSrcBase = i * SplatBuffer.ColorComponentCount;
            const colorDestBase = (i + destOffset) * SplatBuffer.ColorComponentCount;
            outColorArray[colorDestBase] = this.colorArray[colorSrcBase];
            outColorArray[colorDestBase + 1] = this.colorArray[colorSrcBase + 1];
            outColorArray[colorDestBase + 2] = this.colorArray[colorSrcBase + 2];
            outColorArray[colorDestBase + 3] = this.colorArray[colorSrcBase + 3];
            // TODO: implement application of transform for spherical harmonics
        }
    }
}

/**
 * AbortablePromise: A quick & dirty wrapper for JavaScript's Promise class that allows the underlying
 * asynchronous operation to be cancelled. It is only meant for simple situations where no complex promise
 * chaining or merging occurs. It needs a significant amount of work to truly replicate the full
 * functionality of JavaScript's Promise class. Look at Util.fetchWithProgress() for example usage.
 *
 * This class was primarily added to allow splat scene downloads to be cancelled. It has not been tested
 * very thoroughly and the implementation is kinda janky. If you can at all help it, please avoid using it :)
 */
class AbortablePromise {

    constructor(promiseFunc, abortHandler) {

        let promiseResolve;
        let promiseReject;
        this.promise = new Promise((resolve, reject) => {
            promiseResolve = resolve.bind(this);
            promiseReject = reject.bind(this);
        });

        const resolve = (...args) => {
            promiseResolve(...args);
        };

        const reject = (error) => {
            promiseReject(error);
        };

        promiseFunc(resolve.bind(this), reject.bind(this));
        this.abortHandler = abortHandler;
    }

    then(onResolve) {
        return new AbortablePromise((resolve, reject) => {
            this.promise = this.promise
            .then((...args) => {
                const onResolveResult = onResolve(...args);
                if (onResolveResult instanceof Promise || onResolveResult instanceof AbortablePromise) {
                    onResolveResult.then((...args2) => {
                        resolve(...args2);
                    });
                } else {
                    resolve(onResolveResult);
                }
            })
            .catch((error) => {
                reject(error);
            });
        }, this.abortHandler);
    }

    catch(onFail) {
        return new AbortablePromise((resolve) => {
            this.promise = this.promise.then((...args) => {
                resolve(...args);
            })
            .catch(onFail);
        }, this.abortHandler);
    }

    abort() {
        if (this.abortHandler) this.abortHandler();
    }

    static resolve(data) {
        return new AbortablePromise((resolve) => {
            resolve(data);
        });
    }

    static reject(error) {
        return new AbortablePromise((resolve, reject) => {
            reject(error);
        });
    }
}

const floatToHalf = function() {

    const floatView = new Float32Array(1);
    const int32View = new Int32Array(floatView.buffer);

    return function(val) {
        floatView[0] = val;
        const x = int32View[0];

        let bits = (x >> 16) & 0x8000;
        let m = (x >> 12) & 0x07ff;
        const e = (x >> 23) & 0xff;

        if (e < 103) return bits;

        if (e > 142) {
            bits |= 0x7c00;
            bits |= ((e == 255) ? 0 : 1) && (x & 0x007fffff);
            return bits;
        }

        if (e < 113) {
            m |= 0x0800;
            bits |= (m >> (114 - e)) + ((m >> (113 - e)) & 1);
            return bits;
        }

        bits |= (( e - 112) << 10) | (m >> 1);
        bits += m & 1;
        return bits;
    };

}();

const uintEncodedFloat = function() {

    const floatView = new Float32Array(1);
    const int32View = new Int32Array(floatView.buffer);

    return function(f) {
        floatView[0] = f;
        return int32View[0];
    };

}();

const rgbaToInteger = function(r, g, b, a) {
    return r + (g << 8) + (b << 16) + (a << 24);
};

const fetchWithProgress = function(path, onProgress) {

    const abortController = new AbortController();
    const signal = abortController.signal;
    let aborted = false;
    let rejectFunc = null;
    const abortHandler = () => {
        abortController.abort();
        rejectFunc('Fetch aborted');
        aborted = true;
    };

    return new AbortablePromise((resolve, reject) => {
        rejectFunc = reject;
        fetch(path, { signal })
        .then(async (data) => {
            const reader = data.body.getReader();
            let bytesDownloaded = 0;
            let _fileSize = data.headers.get('Content-Length');
            let fileSize = _fileSize ? parseInt(_fileSize) : undefined;

            const chunks = [];

            while (!aborted) {
                try {
                    const { value: chunk, done } = await reader.read();
                    if (done) {
                        if (onProgress) {
                            onProgress(100, '100%', chunk);
                        }
                        const buffer = new Blob(chunks).arrayBuffer();
                        resolve(buffer);
                        break;
                    }
                    bytesDownloaded += chunk.length;
                    let percent;
                    let percentLabel;
                    if (fileSize !== undefined) {
                        percent = bytesDownloaded / fileSize * 100;
                        percentLabel = `${percent.toFixed(2)}%`;
                    }
                    chunks.push(chunk);
                    if (onProgress) {
                        onProgress(percent, percentLabel, chunk);
                    }
                } catch (error) {
                    reject(error);
                    break;
                }
            }
        });
    }, abortHandler);

};

const clamp = function(val, min, max) {
    return Math.max(Math.min(val, max), min);
};

const getCurrentTime = function() {
    return performance.now() / 1000;
};

const disposeAllMeshes = (object3D) => {
    if (object3D.geometry) {
        object3D.geometry.dispose();
        object3D.geometry = null;
    }
    if (object3D.material) {
        object3D.material.dispose();
        object3D.material = null;
    }
    if (object3D.children) {
        for (let child of object3D.children) {
            disposeAllMeshes(child);
        }
    }
};

const SplatBufferBucketSize = 256;
const SplatBufferBucketBlockSize = 5.0;

class UncompressedSplatArray {

    constructor() {
        this.splatCount = 0;
        this.scale_0 = [];
        this.scale_1 = [];
        this.scale_2 = [];
        this.rot_0 = [];
        this.rot_1 = [];
        this.rot_2 = [];
        this.rot_3 = [];
        this.x = [];
        this.y = [];
        this.z = [];
        this.f_dc_0 = [];
        this.f_dc_1 = [];
        this.f_dc_2 = [];
        this.opacity = [];
    }

    addSplat(x, y, z, scale0, scale1, scale2, rot0, rot1, rot2, rot3, r, g, b, opacity) {
        this.x.push(x);
        this.y.push(y);
        this.z.push(z);
        this.scale_0.push(scale0);
        this.scale_1.push(scale1);
        this.scale_2.push(scale2);
        this.rot_0.push(rot0);
        this.rot_1.push(rot1);
        this.rot_2.push(rot2);
        this.rot_3.push(rot3);
        this.f_dc_0.push(r);
        this.f_dc_1.push(g);
        this.f_dc_2.push(b);
        this.opacity.push(opacity);
        this.splatCount++;
    }
}

class SplatCompressor {

    constructor(compressionLevel = 0, minimumAlpha = 1, blockSize = SplatBufferBucketBlockSize, bucketSize = SplatBufferBucketSize) {
        this.compressionLevel = compressionLevel;
        this.minimumAlpha = minimumAlpha;
        this.bucketSize = bucketSize;
        this.blockSize = blockSize;
    }

    static createEmptyUncompressedSplatArray() {
        return new UncompressedSplatArray();
    }

    uncompressedSplatArrayToSplatBuffer(splatArray) {

        const validSplats = SplatCompressor.createEmptyUncompressedSplatArray();
        validSplats.addSplat(0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0);

        for (let i = 0; i < splatArray.splatCount; i++) {
            let alpha;
            if (splatArray['opacity'][i]) {
                alpha = splatArray['opacity'][i];
            } else {
                alpha = 255;
            }
            if (alpha >= this.minimumAlpha) {
                validSplats.addSplat(splatArray['x'][i], splatArray['y'][i], splatArray['z'][i],
                                     splatArray['scale_0'][i], splatArray['scale_1'][i], splatArray['scale_2'][i],
                                     splatArray['rot_0'][i], splatArray['rot_1'][i], splatArray['rot_2'][i], splatArray['rot_3'][i],
                                     splatArray['f_dc_0'][i], splatArray['f_dc_1'][i], splatArray['f_dc_2'][i], splatArray['opacity'][i]);
            }
        }

        const buckets = this.computeBucketsForUncompressedSplatArray(validSplats);

        const paddedSplatCount = buckets.length * this.bucketSize;
        const headerSize = SplatBuffer.HeaderSizeBytes;
        const header = new Uint8Array(new ArrayBuffer(headerSize));
        header[3] = this.compressionLevel;
        (new Uint32Array(header.buffer, 4, 1))[0] = paddedSplatCount;

        let bytesPerCenter = SplatBuffer.CompressionLevels[this.compressionLevel].BytesPerCenter;
        let bytesPerScale = SplatBuffer.CompressionLevels[this.compressionLevel].BytesPerScale;
        let bytesPerColor = SplatBuffer.CompressionLevels[this.compressionLevel].BytesPerColor;
        let bytesPerRotation = SplatBuffer.CompressionLevels[this.compressionLevel].BytesPerRotation;
        const centerBuffer = new ArrayBuffer(bytesPerCenter * paddedSplatCount);
        const scaleBuffer = new ArrayBuffer(bytesPerScale * paddedSplatCount);
        const colorBuffer = new ArrayBuffer(bytesPerColor * paddedSplatCount);
        const rotationBuffer = new ArrayBuffer(bytesPerRotation * paddedSplatCount);

        const blockHalfSize = this.blockSize / 2.0;
        const compressionScaleRange = SplatBuffer.CompressionLevels[this.compressionLevel].ScaleRange;
        const compressionScaleFactor = compressionScaleRange / blockHalfSize;
        const doubleCompressionScaleRange = compressionScaleRange * 2 + 1;

        const bucketCenter = new THREE.Vector3();
        const bucketCenterDelta = new THREE.Vector3();
        let outSplatIndex = 0;
        for (let b = 0; b < buckets.length; b++) {
            const bucket = buckets[b];
            bucketCenter.fromArray(bucket.center);
            for (let i = 0; i < bucket.splats.length; i++) {
                let row = bucket.splats[i];
                let invalidSplat = false;
                if (row === 0) {
                    invalidSplat = true;
                }

                if (this.compressionLevel === 0) {
                    const center = new Float32Array(centerBuffer, outSplatIndex * bytesPerCenter, 3);
                    const scale = new Float32Array(scaleBuffer, outSplatIndex * bytesPerScale, 3);
                    const rot = new Float32Array(rotationBuffer, outSplatIndex * bytesPerRotation, 4);
                    if (validSplats['scale_0'][row] !== undefined) {
                        const quat = new THREE.Quaternion(validSplats['rot_1'][row], validSplats['rot_2'][row],
                                                          validSplats['rot_3'][row], validSplats['rot_0'][row]);
                        quat.normalize();
                        rot.set([quat.w, quat.x, quat.y, quat.z]);
                        scale.set([validSplats['scale_0'][row], validSplats['scale_1'][row], validSplats['scale_2'][row]]);
                    } else {
                        scale.set([0.01, 0.01, 0.01]);
                        rot.set([1.0, 0.0, 0.0, 0.0]);
                    }
                    center.set([validSplats['x'][row], validSplats['y'][row], validSplats['z'][row]]);
                } else {
                    const center = new Uint16Array(centerBuffer, outSplatIndex * bytesPerCenter, 3);
                    const scale = new Uint16Array(scaleBuffer, outSplatIndex * bytesPerScale, 3);
                    const rot = new Uint16Array(rotationBuffer, outSplatIndex * bytesPerRotation, 4);
                    const thf = THREE.DataUtils.toHalfFloat.bind(THREE.DataUtils);
                    if (validSplats['scale_0'][row] !== undefined) {
                        const quat = new THREE.Quaternion(validSplats['rot_1'][row], validSplats['rot_2'][row],
                                                          validSplats['rot_3'][row], validSplats['rot_0'][row]);
                        quat.normalize();
                        rot.set([thf(quat.w), thf(quat.x), thf(quat.y), thf(quat.z)]);
                        scale.set([thf(validSplats['scale_0'][row]), thf(validSplats['scale_1'][row]), thf(validSplats['scale_2'][row])]);
                    } else {
                        scale.set([thf(0.01), thf(0.01), thf(0.01)]);
                        rot.set([thf(1.), 0, 0, 0]);
                    }
                    bucketCenterDelta.set(validSplats['x'][row], validSplats['y'][row], validSplats['z'][row]).sub(bucketCenter);
                    bucketCenterDelta.x = Math.round(bucketCenterDelta.x * compressionScaleFactor) + compressionScaleRange;
                    bucketCenterDelta.x = clamp(bucketCenterDelta.x, 0, doubleCompressionScaleRange);
                    bucketCenterDelta.y = Math.round(bucketCenterDelta.y * compressionScaleFactor) + compressionScaleRange;
                    bucketCenterDelta.y = clamp(bucketCenterDelta.y, 0, doubleCompressionScaleRange);
                    bucketCenterDelta.z = Math.round(bucketCenterDelta.z * compressionScaleFactor) + compressionScaleRange;
                    bucketCenterDelta.z = clamp(bucketCenterDelta.z, 0, doubleCompressionScaleRange);
                    center.set([bucketCenterDelta.x, bucketCenterDelta.y, bucketCenterDelta.z]);
                }

                const rgba = new Uint8ClampedArray(colorBuffer, outSplatIndex * bytesPerColor, 4);
                if (invalidSplat) {
                    rgba[0] = 255;
                    rgba[1] = 0;
                    rgba[2] = 0;
                    rgba[3] = 0;
                } else {
                    if (validSplats['f_dc_0'][row] !== undefined) {
                        rgba.set([validSplats['f_dc_0'][row], validSplats['f_dc_1'][row], validSplats['f_dc_2'][row]]);
                    } else {
                        rgba.set([255, 0, 0]);
                    }
                    if (validSplats['opacity'][row] !== undefined) {
                        rgba[3] = validSplats['opacity'][row];
                    } else {
                        rgba[3] = 255;
                    }
                }

                outSplatIndex++;
            }
        }

        const bytesPerBucket = 12;
        const bucketsSize = bytesPerBucket * buckets.length;
        const splatDataBufferSize = centerBuffer.byteLength + scaleBuffer.byteLength +
                                    colorBuffer.byteLength + rotationBuffer.byteLength;

        const headerArrayUint32 = new Uint32Array(header.buffer);
        const headerArrayFloat32 = new Float32Array(header.buffer);
        let unifiedBufferSize = headerSize + splatDataBufferSize;
        if (this.compressionLevel > 0) {
            unifiedBufferSize += bucketsSize;
            headerArrayUint32[2] = this.bucketSize;
            headerArrayUint32[3] = buckets.length;
            headerArrayFloat32[4] = this.blockSize;
            headerArrayUint32[5] = bytesPerBucket;
            headerArrayUint32[6] = SplatBuffer.CompressionLevels[this.compressionLevel].ScaleRange;
        }

        const unifiedBuffer = new ArrayBuffer(unifiedBufferSize);
        new Uint8Array(unifiedBuffer, 0, headerSize).set(header);
        new Uint8Array(unifiedBuffer, headerSize, centerBuffer.byteLength).set(new Uint8Array(centerBuffer));
        new Uint8Array(unifiedBuffer, headerSize + centerBuffer.byteLength, scaleBuffer.byteLength).set(new Uint8Array(scaleBuffer));
        new Uint8Array(unifiedBuffer, headerSize + centerBuffer.byteLength + scaleBuffer.byteLength,
                    colorBuffer.byteLength).set(new Uint8Array(colorBuffer));
        new Uint8Array(unifiedBuffer, headerSize + centerBuffer.byteLength + scaleBuffer.byteLength + colorBuffer.byteLength,
                    rotationBuffer.byteLength).set(new Uint8Array(rotationBuffer));

        if (this.compressionLevel > 0) {
            const bucketArray = new Float32Array(unifiedBuffer, headerSize + splatDataBufferSize, buckets.length * 3);
            for (let i = 0; i < buckets.length; i++) {
                const bucket = buckets[i];
                const base = i * 3;
                bucketArray[base] = bucket.center[0];
                bucketArray[base + 1] = bucket.center[1];
                bucketArray[base + 2] = bucket.center[2];
            }
        }

        const splatBuffer = new SplatBuffer(unifiedBuffer);
        return splatBuffer;
    }

    computeBucketsForUncompressedSplatArray(splatArray) {
        let splatCount = splatArray.splatCount;
        const blockSize = this.blockSize;
        const halfBlockSize = blockSize / 2.0;

        const min = new THREE.Vector3();
        const max = new THREE.Vector3();

        // ignore the first splat since it's the invalid designator
        for (let i = 1; i < splatCount; i++) {
            const center = [splatArray['x'][i], splatArray['y'][i], splatArray['z'][i]];
            if (i === 0 || center[0] < min.x) min.x = center[0];
            if (i === 0 || center[0] > max.x) max.x = center[0];
            if (i === 0 || center[1] < min.y) min.y = center[1];
            if (i === 0 || center[1] > max.y) max.y = center[1];
            if (i === 0 || center[2] < min.z) min.z = center[2];
            if (i === 0 || center[2] > max.z) max.z = center[2];
        }

        const dimensions = new THREE.Vector3().copy(max).sub(min);
        const yBlocks = Math.ceil(dimensions.y / blockSize);
        const zBlocks = Math.ceil(dimensions.z / blockSize);

        const blockCenter = new THREE.Vector3();
        const fullBuckets = [];
        const partiallyFullBuckets = {};

        // ignore the first splat since it's the invalid designator
        for (let i = 1; i < splatCount; i++) {
            const center = [splatArray['x'][i], splatArray['y'][i], splatArray['z'][i]];
            const xBlock = Math.floor((center[0] - min.x) / blockSize);
            const yBlock = Math.floor((center[1] - min.y) / blockSize);
            const zBlock = Math.floor((center[2] - min.z) / blockSize);

            blockCenter.x = xBlock * blockSize + min.x + halfBlockSize;
            blockCenter.y = yBlock * blockSize + min.y + halfBlockSize;
            blockCenter.z = zBlock * blockSize + min.z + halfBlockSize;

            const bucketId = xBlock * (yBlocks * zBlocks) + yBlock * zBlocks + zBlock;
            let bucket = partiallyFullBuckets[bucketId];
            if (!bucket) {
                partiallyFullBuckets[bucketId] = bucket = {
                    'splats': [],
                    'center': blockCenter.toArray()
                };
            }

            bucket.splats.push(i);
            if (bucket.splats.length >= this.bucketSize) {
                fullBuckets.push(bucket);
                partiallyFullBuckets[bucketId] = null;
            }
        }

        // fill partially full buckets with invalid splats (splat 0)
        // to get them up to this.bucketSize
        for (let bucketId in partiallyFullBuckets) {
            if (partiallyFullBuckets.hasOwnProperty(bucketId)) {
                const bucket = partiallyFullBuckets[bucketId];
                if (bucket) {
                    while (bucket.splats.length < this.bucketSize) {
                        bucket.splats.push(0);
                    }
                    fullBuckets.push(bucket);
                }
            }
        }

        return fullBuckets;
    }
}

class PlyParser {

    constructor(plyBuffer) {
        this.plyBuffer = plyBuffer;
    }

    decodeHeader(plyBuffer) {
        const decoder = new TextDecoder();
        let headerOffset = 0;
        let headerText = '';

        console.log('.PLY size: ' + plyBuffer.byteLength + ' bytes');

        const readChunkSize = 100;

        while (true) {
            if (headerOffset + readChunkSize >= plyBuffer.byteLength) {
                throw new Error('End of file reached while searching for end of header');
            }
            const headerChunk = new Uint8Array(plyBuffer, headerOffset, readChunkSize);
            headerText += decoder.decode(headerChunk);
            headerOffset += readChunkSize;

            const endHeaderTestChunk = new Uint8Array(plyBuffer, Math.max(0, headerOffset - readChunkSize * 2), readChunkSize * 2);
            const endHeaderTestText = decoder.decode(endHeaderTestChunk);
            if (endHeaderTestText.includes('end_header')) {
                break;
            }
        }

        const headerLines = headerText.split('\n');

        let splatCount = 0;
        let propertyTypes = {};

        for (let i = 0; i < headerLines.length; i++) {
            const line = headerLines[i].trim();
            if (line.startsWith('element vertex')) {
                const splatCountMatch = line.match(/\d+/);
                if (splatCountMatch) {
                    splatCount = parseInt(splatCountMatch[0]);
                }
            } else if (line.startsWith('property')) {
                const propertyMatch = line.match(/(\w+)\s+(\w+)\s+(\w+)/);
                if (propertyMatch) {
                    const propertyType = propertyMatch[2];
                    const propertyName = propertyMatch[3];
                    propertyTypes[propertyName] = propertyType;
                }
            } else if (line === 'end_header') {
                break;
            }
        }

        const vertexByteOffset = headerText.indexOf('end_header') + 'end_header'.length + 1;
        const vertexData = new DataView(plyBuffer, vertexByteOffset);

        return {
            'splatCount': splatCount,
            'propertyTypes': propertyTypes,
            'vertexData': vertexData,
            'headerOffset': headerOffset
        };
    }

    readRawVertexFast(vertexData, offset, fieldOffsets, propertiesToRead, propertyTypes, outVertex) {
        let rawVertex = outVertex || {};
        for (let property of propertiesToRead) {
            const propertyType = propertyTypes[property];
            if (propertyType === 'float') {
                rawVertex[property] = vertexData.getFloat32(offset + fieldOffsets[property], true);
            } else if (propertyType === 'uchar') {
                rawVertex[property] = vertexData.getUint8(offset + fieldOffsets[property]) / 255.0;
            }
        }
    }

    parseToSplatBuffer(compressionLevel, minimumAlpha, blockSize, bucketSize) {

        const startTime = performance.now();

        console.log('Parsing PLY to SPLAT...');

        const {splatCount, propertyTypes, vertexData} = this.decodeHeader(this.plyBuffer);

        // figure out the SH degree from the number of coefficients
        let nRestCoeffs = 0;
        for (const propertyName in propertyTypes) {
            if (propertyName.startsWith('f_rest_')) {
                nRestCoeffs += 1;
            }
        }
        const nCoeffsPerColor = nRestCoeffs / 3;

        // TODO: Eventually properly support multiple degree spherical harmonics
        // const sphericalHarmonicsDegree = Math.sqrt(nCoeffsPerColor + 1) - 1;
        const sphericalHarmonicsDegree = 0;

        console.log('Detected degree', sphericalHarmonicsDegree, 'with ', nCoeffsPerColor, 'coefficients per color');

        // figure out the order in which spherical harmonics should be read
        const shFeatureOrder = [];
        for (let rgb = 0; rgb < 3; ++rgb) {
            shFeatureOrder.push(`f_dc_${rgb}`);
        }
        for (let i = 0; i < nCoeffsPerColor; ++i) {
            for (let rgb = 0; rgb < 3; ++rgb) {
                shFeatureOrder.push(`f_rest_${rgb * nCoeffsPerColor + i}`);
            }
        }

        let plyRowSize = 0;
        let fieldOffsets = {};
        const fieldSize = {
            'double': 8,
            'int': 4,
            'uint': 4,
            'float': 4,
            'short': 2,
            'ushort': 2,
            'uchar': 1,
        };
        for (let fieldName in propertyTypes) {
            if (propertyTypes.hasOwnProperty(fieldName)) {
                const type = propertyTypes[fieldName];
                fieldOffsets[fieldName] = plyRowSize;
                plyRowSize += fieldSize[type];
            }
        }

        let rawVertex = {};

        const propertiesToRead = ['scale_0', 'scale_1', 'scale_2', 'rot_0', 'rot_1', 'rot_2', 'rot_3',
                                  'x', 'y', 'z', 'f_dc_0', 'f_dc_1', 'f_dc_2', 'opacity'];

        const splatArray = SplatCompressor.createEmptyUncompressedSplatArray();

        for (let row = 0; row < splatCount; row++) {
            this.readRawVertexFast(vertexData, row * plyRowSize, fieldOffsets, propertiesToRead, propertyTypes, rawVertex);
            if (rawVertex['scale_0'] !== undefined) {
                splatArray['scale_0'][row] = Math.exp(rawVertex['scale_0']);
                splatArray['scale_1'][row] = Math.exp(rawVertex['scale_1']);
                splatArray['scale_2'][row] = Math.exp(rawVertex['scale_2']);
            } else {
                splatArray['scale_0'][row] = 0.01;
                splatArray['scale_1'][row] = 0.01;
                splatArray['scale_2'][row] = 0.01;
            }

            if (rawVertex['f_dc_0'] !== undefined) {
                const SH_C0 = 0.28209479177387814;
                splatArray['f_dc_0'][row] = (0.5 + SH_C0 * rawVertex['f_dc_0']) * 255;
                splatArray['f_dc_1'][row] = (0.5 + SH_C0 * rawVertex['f_dc_1']) * 255;
                splatArray['f_dc_2'][row] = (0.5 + SH_C0 * rawVertex['f_dc_2']) * 255;
            } else {
                splatArray['f_dc_0'][row] = 0;
                splatArray['f_dc_1'][row] = 0;
                splatArray['f_dc_2'][row] = 0;
            }
            if (rawVertex['opacity'] !== undefined) {
                splatArray['opacity'][row] = (1 / (1 + Math.exp(-rawVertex['opacity']))) * 255;
            }

            splatArray['rot_0'][row] = rawVertex['rot_0'];
            splatArray['rot_1'][row] = rawVertex['rot_1'];
            splatArray['rot_2'][row] = rawVertex['rot_2'];
            splatArray['rot_3'][row] = rawVertex['rot_3'];

            splatArray['x'][row] = rawVertex['x'];
            splatArray['y'][row] = rawVertex['y'];
            splatArray['z'][row] = rawVertex['z'];
            splatArray.splatCount++;
        }

        const splatCompressor = new SplatCompressor(compressionLevel, minimumAlpha, blockSize, bucketSize);
        const splatBuffer = splatCompressor.uncompressedSplatArrayToSplatBuffer(splatArray);

        console.log('Total valid splats: ', splatBuffer.getSplatCount(), 'out of', splatCount);

        const endTime = performance.now();

        console.log('Parsing PLY to SPLAT complete!');
        console.log('Total time: ', (endTime - startTime).toFixed(2) + ' ms');

        return splatBuffer;
    }

}

class PlyLoader {

    constructor() {
        this.splatBuffer = null;
    }

    loadFromURL(fileName, onProgress, compressionLevel, minimumAlpha, blockSize, bucketSize) {
        return fetchWithProgress(fileName, onProgress).then((plyFileData) => {
            const plyParser = new PlyParser(plyFileData);
            const splatBuffer = plyParser.parseToSplatBuffer(compressionLevel, minimumAlpha, blockSize, bucketSize);
            this.splatBuffer = splatBuffer;
            return splatBuffer;
        });
    }

}

const SceneFormat = {
    'Splat': 0,
    'KSplat': 1,
    'Ply': 2
};

class SplatLoader {

    constructor(splatBuffer = null) {
        this.splatBuffer = splatBuffer;
        this.downLoadLink = null;
    }

    static isFileSplatFormat(fileName) {
        return SplatLoader.isCustomSplatFormat(fileName) || SplatLoader.isStandardSplatFormat(fileName);
    }

    static isCustomSplatFormat(fileName) {
        return fileName.endsWith('.ksplat');
    }

    static isStandardSplatFormat(fileName) {
        return fileName.endsWith('.splat');
    }

    loadFromURL(fileName, onProgress, compressionLevel, minimumAlpha, blockSize, bucketSize, format) {
        return fetchWithProgress(fileName, onProgress).then((bufferData) => {
            const isCustomSplatFormat = format === SceneFormat.KSplat || SplatLoader.isCustomSplatFormat(fileName);
            let splatBuffer;
            if (isCustomSplatFormat) {
                splatBuffer = new SplatBuffer(bufferData);
            } else {
                const splatCompressor = new SplatCompressor(compressionLevel, minimumAlpha, blockSize, bucketSize);
                const splatArray = SplatLoader.parseStandardSplatToUncompressedSplatArray(bufferData);
                splatBuffer = splatCompressor.uncompressedSplatArrayToSplatBuffer(splatArray);
            }
            return splatBuffer;
        });
    }

    static parseStandardSplatToUncompressedSplatArray(inBuffer) {
        // Standard .splat row layout:
        // XYZ - Position (Float32)
        // XYZ - Scale (Float32)
        // RGBA - colors (uint8)
        // IJKL - quaternion/rot (uint8)

        const InBufferRowSizeBytes = 32;
        const splatCount = inBuffer.byteLength / InBufferRowSizeBytes;

        const splatArray = SplatCompressor.createEmptyUncompressedSplatArray();

        for (let i = 0; i < splatCount; i++) {
            const inCenterSizeBytes = 3 * 4;
            const inScaleSizeBytes = 3 * 4;
            const inColorSizeBytes = 4;
            const inBase = i * InBufferRowSizeBytes;
            const inCenter = new Float32Array(inBuffer, inBase, 3);
            const inScale = new Float32Array(inBuffer, inBase + inCenterSizeBytes, 3);
            const inColor = new Uint8Array(inBuffer, inBase + inCenterSizeBytes + inScaleSizeBytes, 4);
            const inRotation = new Uint8Array(inBuffer, inBase + inCenterSizeBytes + inScaleSizeBytes + inColorSizeBytes, 4);

            const quat = new THREE.Quaternion((inRotation[1] - 128) / 128, (inRotation[2] - 128) / 128,
                                              (inRotation[3] - 128) / 128, (inRotation[0] - 128) / 128);
            quat.normalize();

            splatArray.addSplat(inCenter[0], inCenter[1], inCenter[2], inScale[0], inScale[1], inScale[2],
                                quat.w, quat.x, quat.y, quat.z, inColor[0], inColor[1], inColor[2], inColor[3]);
        }

        return splatArray;
    }

    setFromBuffer(splatBuffer) {
        this.splatBuffer = splatBuffer;
    }

    downloadFile(fileName) {
        const headerData = new Uint8Array(this.splatBuffer.getHeaderBufferData());
        const splatData = new Uint8Array(this.splatBuffer.getSplatBufferData());
        const blob = new Blob([headerData.buffer, splatData.buffer], {
            type: 'application/octet-stream',
        });

        if (!this.downLoadLink) {
            this.downLoadLink = document.createElement('a');
            document.body.appendChild(this.downLoadLink);
        }
        this.downLoadLink.download = fileName;
        this.downLoadLink.href = URL.createObjectURL(blob);
        this.downLoadLink.click();
    }

}

/*
Copyright © 2010-2024 three.js authors & Mark Kellogg

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
*/


// OrbitControls performs orbiting, dollying (zooming), and panning.
// Unlike TrackballControls, it maintains the "up" direction object.up (+Y by default).
//
//    Orbit - left mouse / touch: one-finger move
//    Zoom - middle mouse, or mousewheel / touch: two-finger spread or squish
//    Pan - right mouse, or left mouse + ctrl/meta/shiftKey, or arrow keys / touch: two-finger move

const _changeEvent = { type: 'change' };
const _startEvent = { type: 'start' };
const _endEvent = { type: 'end' };
const _ray = new Ray$1();
const _plane = new Plane();
const TILT_LIMIT = Math.cos( 70 * MathUtils.DEG2RAD );

class OrbitControls extends EventDispatcher {

    constructor( object, domElement ) {

        super();

        this.object = object;
        this.domElement = domElement;
        this.domElement.style.touchAction = 'none'; // disable touch scroll

        // Set to false to disable this control
        this.enabled = true;

        // "target" sets the location of focus, where the object orbits around
        this.target = new Vector3();

        // How far you can dolly in and out ( PerspectiveCamera only )
        this.minDistance = 0;
        this.maxDistance = Infinity;

        // How far you can zoom in and out ( OrthographicCamera only )
        this.minZoom = 0;
        this.maxZoom = Infinity;

        // How far you can orbit vertically, upper and lower limits.
        // Range is 0 to Math.PI radians.
        this.minPolarAngle = 0; // radians
        this.maxPolarAngle = Math.PI; // radians

        // How far you can orbit horizontally, upper and lower limits.
        // If set, the interval [min, max] must be a sub-interval of [- 2 PI, 2 PI], with ( max - min < 2 PI )
        this.minAzimuthAngle = - Infinity; // radians
        this.maxAzimuthAngle = Infinity; // radians

        // Set to true to enable damping (inertia)
        // If damping is enabled, you must call controls.update() in your animation loop
        this.enableDamping = false;
        this.dampingFactor = 0.05;

        // This option actually enables dollying in and out; left as "zoom" for backwards compatibility.
        // Set to false to disable zooming
        this.enableZoom = true;
        this.zoomSpeed = 1.0;

        // Set to false to disable rotating
        this.enableRotate = true;
        this.rotateSpeed = 1.0;

        // Set to false to disable panning
        this.enablePan = true;
        this.panSpeed = 1.0;
        this.screenSpacePanning = true; // if false, pan orthogonal to world-space direction camera.up
        this.keyPanSpeed = 7.0; // pixels moved per arrow key push
        this.zoomToCursor = false;

        // Set to true to automatically rotate around the target
        // If auto-rotate is enabled, you must call controls.update() in your animation loop
        this.autoRotate = false;
        this.autoRotateSpeed = 2.0; // 30 seconds per orbit when fps is 60

        // The four arrow keys
        this.keys = { LEFT: 'KeyA', UP: 'KeyW', RIGHT: 'KeyD', BOTTOM: 'KeyS' };

        // Mouse buttons
        this.mouseButtons = { LEFT: MOUSE.ROTATE, MIDDLE: MOUSE.DOLLY, RIGHT: MOUSE.PAN };

        // Touch fingers
        this.touches = { ONE: TOUCH.ROTATE, TWO: TOUCH.DOLLY_PAN };

        // for reset
        this.target0 = this.target.clone();
        this.position0 = this.object.position.clone();
        this.zoom0 = this.object.zoom;

        // the target DOM element for key events
        this._domElementKeyEvents = null;

        //
        // public methods
        //

        this.getPolarAngle = function() {

            return spherical.phi;

        };

        this.getAzimuthalAngle = function() {

            return spherical.theta;

        };

        this.getDistance = function() {

            return this.object.position.distanceTo( this.target );

        };

        this.listenToKeyEvents = function( domElement ) {

            domElement.addEventListener( 'keydown', onKeyDown );
            this._domElementKeyEvents = domElement;

        };

        this.stopListenToKeyEvents = function() {

            this._domElementKeyEvents.removeEventListener( 'keydown', onKeyDown );
            this._domElementKeyEvents = null;

        };

        this.saveState = function() {

            scope.target0.copy( scope.target );
            scope.position0.copy( scope.object.position );
            scope.zoom0 = scope.object.zoom;

        };

        this.reset = function() {

            scope.target.copy( scope.target0 );
            scope.object.position.copy( scope.position0 );
            scope.object.zoom = scope.zoom0;

            scope.object.updateProjectionMatrix();
            scope.dispatchEvent( _changeEvent );

            scope.update();

            state = STATE.NONE;

        };

        // this method is exposed, but perhaps it would be better if we can make it private...
        this.update = function() {

            const offset = new Vector3();

            // so camera.up is the orbit axis
            const quat = new Quaternion().setFromUnitVectors( object.up, new Vector3( 0, 1, 0 ) );
            const quatInverse = quat.clone().invert();

            const lastPosition = new Vector3();
            const lastQuaternion = new Quaternion();
            const lastTargetPosition = new Vector3();

            const twoPI = 2 * Math.PI;

            return function update() {

                quat.setFromUnitVectors( object.up, new Vector3( 0, 1, 0 ) );
                quatInverse.copy(quat).invert();

                const position = scope.object.position;

                offset.copy( position ).sub( scope.target );

                // rotate offset to "y-axis-is-up" space
                offset.applyQuaternion( quat );

                // angle from z-axis around y-axis
                spherical.setFromVector3( offset );

                if ( scope.autoRotate && state === STATE.NONE ) {

                    rotateLeft( getAutoRotationAngle() );

                }

                if ( scope.enableDamping ) {

                    spherical.theta += sphericalDelta.theta * scope.dampingFactor;
                    spherical.phi += sphericalDelta.phi * scope.dampingFactor;

                } else {

                    spherical.theta += sphericalDelta.theta;
                    spherical.phi += sphericalDelta.phi;

                }

                // restrict theta to be between desired limits

                let min = scope.minAzimuthAngle;
                let max = scope.maxAzimuthAngle;

                if ( isFinite( min ) && isFinite( max ) ) {

                    if ( min < - Math.PI ) min += twoPI; else if ( min > Math.PI ) min -= twoPI;

                    if ( max < - Math.PI ) max += twoPI; else if ( max > Math.PI ) max -= twoPI;

                    if ( min <= max ) {

                        spherical.theta = Math.max( min, Math.min( max, spherical.theta ) );

                    } else {

                        spherical.theta = ( spherical.theta > ( min + max ) / 2 ) ?
                            Math.max( min, spherical.theta ) :
                            Math.min( max, spherical.theta );

                    }

                }

                // restrict phi to be between desired limits
                spherical.phi = Math.max( scope.minPolarAngle, Math.min( scope.maxPolarAngle, spherical.phi ) );

                spherical.makeSafe();


                // move target to panned location

                if ( scope.enableDamping === true ) {

                    scope.target.addScaledVector( panOffset, scope.dampingFactor );

                } else {

                    scope.target.add( panOffset );

                }

                // adjust the camera position based on zoom only if we're not zooming to the cursor or if it's an ortho camera
                // we adjust zoom later in these cases
                if ( scope.zoomToCursor && performCursorZoom || scope.object.isOrthographicCamera ) {

                    spherical.radius = clampDistance( spherical.radius );

                } else {

                    spherical.radius = clampDistance( spherical.radius * scale );

                }


                offset.setFromSpherical( spherical );

                // rotate offset back to "camera-up-vector-is-up" space
                offset.applyQuaternion( quatInverse );

                position.copy( scope.target ).add( offset );

                scope.object.lookAt( scope.target );

                if ( scope.enableDamping === true ) {

                    sphericalDelta.theta *= ( 1 - scope.dampingFactor );
                    sphericalDelta.phi *= ( 1 - scope.dampingFactor );

                    panOffset.multiplyScalar( 1 - scope.dampingFactor );

                } else {

                    sphericalDelta.set( 0, 0, 0 );

                    panOffset.set( 0, 0, 0 );

                }

                // adjust camera position
                let zoomChanged = false;
                if ( scope.zoomToCursor && performCursorZoom ) {

                    let newRadius = null;
                    if ( scope.object.isPerspectiveCamera ) {

                        // move the camera down the pointer ray
                        // this method avoids floating point error
                        const prevRadius = offset.length();
                        newRadius = clampDistance( prevRadius * scale );

                        const radiusDelta = prevRadius - newRadius;
                        scope.object.position.addScaledVector( dollyDirection, radiusDelta );
                        scope.object.updateMatrixWorld();

                    } else if ( scope.object.isOrthographicCamera ) {

                        // adjust the ortho camera position based on zoom changes
                        const mouseBefore = new Vector3( mouse.x, mouse.y, 0 );
                        mouseBefore.unproject( scope.object );

                        scope.object.zoom = Math.max( scope.minZoom, Math.min( scope.maxZoom, scope.object.zoom / scale ) );
                        scope.object.updateProjectionMatrix();
                        zoomChanged = true;

                        const mouseAfter = new Vector3( mouse.x, mouse.y, 0 );
                        mouseAfter.unproject( scope.object );

                        scope.object.position.sub( mouseAfter ).add( mouseBefore );
                        scope.object.updateMatrixWorld();

                        newRadius = offset.length();

                    } else {

                        console.warn( 'WARNING: OrbitControls.js encountered an unknown camera type - zoom to cursor disabled.' );
                        scope.zoomToCursor = false;

                    }

                    // handle the placement of the target
                    if ( newRadius !== null ) {

                        if ( this.screenSpacePanning ) {

                            // position the orbit target in front of the new camera position
                            scope.target.set( 0, 0, - 1 )
                                .transformDirection( scope.object.matrix )
                                .multiplyScalar( newRadius )
                                .add( scope.object.position );

                        } else {

                            // get the ray and translation plane to compute target
                            _ray.origin.copy( scope.object.position );
                            _ray.direction.set( 0, 0, - 1 ).transformDirection( scope.object.matrix );

                            // if the camera is 20 degrees above the horizon then don't adjust the focus target to avoid
                            // extremely large values
                            if ( Math.abs( scope.object.up.dot( _ray.direction ) ) < TILT_LIMIT ) {

                                object.lookAt( scope.target );

                            } else {

                                _plane.setFromNormalAndCoplanarPoint( scope.object.up, scope.target );
                                _ray.intersectPlane( _plane, scope.target );

                            }

                        }

                    }

                } else if ( scope.object.isOrthographicCamera ) {

                    scope.object.zoom = Math.max( scope.minZoom, Math.min( scope.maxZoom, scope.object.zoom / scale ) );
                    scope.object.updateProjectionMatrix();
                    zoomChanged = true;

                }

                scale = 1;
                performCursorZoom = false;

                // update condition is:
                // min(camera displacement, camera rotation in radians)^2 > EPS
                // using small-angle approximation cos(x/2) = 1 - x^2 / 8

                if ( zoomChanged ||
                    lastPosition.distanceToSquared( scope.object.position ) > EPS ||
                    8 * ( 1 - lastQuaternion.dot( scope.object.quaternion ) ) > EPS ||
                    lastTargetPosition.distanceToSquared( scope.target ) > 0 ) {

                    scope.dispatchEvent( _changeEvent );

                    lastPosition.copy( scope.object.position );
                    lastQuaternion.copy( scope.object.quaternion );
                    lastTargetPosition.copy( scope.target );

                    zoomChanged = false;

                    return true;

                }

                return false;

            };

        }();

        this.dispose = function() {

            scope.domElement.removeEventListener( 'contextmenu', onContextMenu );

            scope.domElement.removeEventListener( 'pointerdown', onPointerDown );
            scope.domElement.removeEventListener( 'pointercancel', onPointerUp );
            scope.domElement.removeEventListener( 'wheel', onMouseWheel );

            scope.domElement.removeEventListener( 'pointermove', onPointerMove );
            scope.domElement.removeEventListener( 'pointerup', onPointerUp );


            if ( scope._domElementKeyEvents !== null ) {

                scope._domElementKeyEvents.removeEventListener( 'keydown', onKeyDown );
                scope._domElementKeyEvents = null;

            }

        };

        //
        // internals
        //

        const scope = this;

        const STATE = {
            NONE: - 1,
            ROTATE: 0,
            DOLLY: 1,
            PAN: 2,
            TOUCH_ROTATE: 3,
            TOUCH_PAN: 4,
            TOUCH_DOLLY_PAN: 5,
            TOUCH_DOLLY_ROTATE: 6
        };

        let state = STATE.NONE;

        const EPS = 0.000001;

        // current position in spherical coordinates
        const spherical = new Spherical();
        const sphericalDelta = new Spherical();

        let scale = 1;
        const panOffset = new Vector3();

        const rotateStart = new Vector2();
        const rotateEnd = new Vector2();
        const rotateDelta = new Vector2();

        const panStart = new Vector2();
        const panEnd = new Vector2();
        const panDelta = new Vector2();

        const dollyStart = new Vector2();
        const dollyEnd = new Vector2();
        const dollyDelta = new Vector2();

        const dollyDirection = new Vector3();
        const mouse = new Vector2();
        let performCursorZoom = false;

        const pointers = [];
        const pointerPositions = {};

        function getAutoRotationAngle() {

            return 2 * Math.PI / 60 / 60 * scope.autoRotateSpeed;

        }

        function getZoomScale() {

            return Math.pow( 0.95, scope.zoomSpeed );

        }

        function rotateLeft( angle ) {

            sphericalDelta.theta -= angle;

        }

        function rotateUp( angle ) {

            sphericalDelta.phi -= angle;

        }

        const panLeft = function() {

            const v = new Vector3();

            return function panLeft( distance, objectMatrix ) {

                v.setFromMatrixColumn( objectMatrix, 0 ); // get X column of objectMatrix
                v.multiplyScalar( - distance );

                panOffset.add( v );

            };

        }();

        const panUp = function() {

            const v = new Vector3();

            return function panUp( distance, objectMatrix ) {

                if ( scope.screenSpacePanning === true ) {

                    v.setFromMatrixColumn( objectMatrix, 1 );

                } else {

                    v.setFromMatrixColumn( objectMatrix, 0 );
                    v.crossVectors( scope.object.up, v );

                }

                v.multiplyScalar( distance );

                panOffset.add( v );

            };

        }();

        // deltaX and deltaY are in pixels; right and down are positive
        const pan = function() {

            const offset = new Vector3();

            return function pan( deltaX, deltaY ) {

                const element = scope.domElement;

                if ( scope.object.isPerspectiveCamera ) {

                    // perspective
                    const position = scope.object.position;
                    offset.copy( position ).sub( scope.target );
                    let targetDistance = offset.length();

                    // half of the fov is center to top of screen
                    targetDistance *= Math.tan( ( scope.object.fov / 2 ) * Math.PI / 180.0 );

                    // we use only clientHeight here so aspect ratio does not distort speed
                    panLeft( 2 * deltaX * targetDistance / element.clientHeight, scope.object.matrix );
                    panUp( 2 * deltaY * targetDistance / element.clientHeight, scope.object.matrix );

                } else if ( scope.object.isOrthographicCamera ) {

                    // orthographic
                    panLeft( deltaX * ( scope.object.right - scope.object.left ) /
                                        scope.object.zoom / element.clientWidth, scope.object.matrix );
                    panUp( deltaY * ( scope.object.top - scope.object.bottom ) / scope.object.zoom /
                                      element.clientHeight, scope.object.matrix );

                } else {

                    // camera neither orthographic nor perspective
                    console.warn( 'WARNING: OrbitControls.js encountered an unknown camera type - pan disabled.' );
                    scope.enablePan = false;

                }

            };

        }();

        function dollyOut( dollyScale ) {

            if ( scope.object.isPerspectiveCamera || scope.object.isOrthographicCamera ) {

                scale /= dollyScale;

            } else {

                console.warn( 'WARNING: OrbitControls.js encountered an unknown camera type - dolly/zoom disabled.' );
                scope.enableZoom = false;

            }

        }

        function dollyIn( dollyScale ) {

            if ( scope.object.isPerspectiveCamera || scope.object.isOrthographicCamera ) {

                scale *= dollyScale;

            } else {

                console.warn( 'WARNING: OrbitControls.js encountered an unknown camera type - dolly/zoom disabled.' );
                scope.enableZoom = false;

            }

        }

        function updateMouseParameters( event ) {

            if ( ! scope.zoomToCursor ) {

                return;

            }

            performCursorZoom = true;

            const rect = scope.domElement.getBoundingClientRect();
            const x = event.clientX - rect.left;
            const y = event.clientY - rect.top;
            const w = rect.width;
            const h = rect.height;

            mouse.x = ( x / w ) * 2 - 1;
            mouse.y = - ( y / h ) * 2 + 1;

            dollyDirection.set( mouse.x, mouse.y, 1 ).unproject( object ).sub( object.position ).normalize();

        }

        function clampDistance( dist ) {

            return Math.max( scope.minDistance, Math.min( scope.maxDistance, dist ) );

        }

        //
        // event callbacks - update the object state
        //

        function handleMouseDownRotate( event ) {

            rotateStart.set( event.clientX, event.clientY );

        }

        function handleMouseDownDolly( event ) {

            updateMouseParameters( event );
            dollyStart.set( event.clientX, event.clientY );

        }

        function handleMouseDownPan( event ) {

            panStart.set( event.clientX, event.clientY );

        }

        function handleMouseMoveRotate( event ) {

            rotateEnd.set( event.clientX, event.clientY );

            rotateDelta.subVectors( rotateEnd, rotateStart ).multiplyScalar( scope.rotateSpeed );

            const element = scope.domElement;

            rotateLeft( 2 * Math.PI * rotateDelta.x / element.clientHeight ); // yes, height

            rotateUp( 2 * Math.PI * rotateDelta.y / element.clientHeight );

            rotateStart.copy( rotateEnd );

            scope.update();

        }

        function handleMouseMoveDolly( event ) {

            dollyEnd.set( event.clientX, event.clientY );

            dollyDelta.subVectors( dollyEnd, dollyStart );

            if ( dollyDelta.y > 0 ) {

                dollyOut( getZoomScale() );

            } else if ( dollyDelta.y < 0 ) {

                dollyIn( getZoomScale() );

            }

            dollyStart.copy( dollyEnd );

            scope.update();

        }

        function handleMouseMovePan( event ) {

            panEnd.set( event.clientX, event.clientY );

            panDelta.subVectors( panEnd, panStart ).multiplyScalar( scope.panSpeed );

            pan( panDelta.x, panDelta.y );

            panStart.copy( panEnd );

            scope.update();

        }

        function handleMouseWheel( event ) {

            updateMouseParameters( event );

            if ( event.deltaY < 0 ) {

                dollyIn( getZoomScale() );

            } else if ( event.deltaY > 0 ) {

                dollyOut( getZoomScale() );

            }

            scope.update();

        }

        function handleKeyDown( event ) {

            let needsUpdate = false;

            switch ( event.code ) {

                case scope.keys.UP:

                    if ( event.ctrlKey || event.metaKey || event.shiftKey ) {

                        rotateUp( 2 * Math.PI * scope.rotateSpeed / scope.domElement.clientHeight );

                    } else {

                        pan( 0, scope.keyPanSpeed );

                    }

                    needsUpdate = true;
                    break;

                case scope.keys.BOTTOM:

                    if ( event.ctrlKey || event.metaKey || event.shiftKey ) {

                        rotateUp( - 2 * Math.PI * scope.rotateSpeed / scope.domElement.clientHeight );

                    } else {

                        pan( 0, - scope.keyPanSpeed );

                    }

                    needsUpdate = true;
                    break;

                case scope.keys.LEFT:

                    if ( event.ctrlKey || event.metaKey || event.shiftKey ) {

                        rotateLeft( 2 * Math.PI * scope.rotateSpeed / scope.domElement.clientHeight );

                    } else {

                        pan( scope.keyPanSpeed, 0 );

                    }

                    needsUpdate = true;
                    break;

                case scope.keys.RIGHT:

                    if ( event.ctrlKey || event.metaKey || event.shiftKey ) {

                        rotateLeft( - 2 * Math.PI * scope.rotateSpeed / scope.domElement.clientHeight );

                    } else {

                        pan( - scope.keyPanSpeed, 0 );

                    }

                    needsUpdate = true;
                    break;

            }

            if ( needsUpdate ) {

                // prevent the browser from scrolling on cursor keys
                event.preventDefault();

                scope.update();

            }


        }

        function handleTouchStartRotate() {

            if ( pointers.length === 1 ) {

                rotateStart.set( pointers[0].pageX, pointers[0].pageY );

            } else {

                const x = 0.5 * ( pointers[0].pageX + pointers[1].pageX );
                const y = 0.5 * ( pointers[0].pageY + pointers[1].pageY );

                rotateStart.set( x, y );

            }

        }

        function handleTouchStartPan() {

            if ( pointers.length === 1 ) {

                panStart.set( pointers[0].pageX, pointers[0].pageY );

            } else {

                const x = 0.5 * ( pointers[0].pageX + pointers[1].pageX );
                const y = 0.5 * ( pointers[0].pageY + pointers[1].pageY );

                panStart.set( x, y );

            }

        }

        function handleTouchStartDolly() {

            const dx = pointers[0].pageX - pointers[1].pageX;
            const dy = pointers[0].pageY - pointers[1].pageY;

            const distance = Math.sqrt( dx * dx + dy * dy );

            dollyStart.set( 0, distance );

        }

        function handleTouchStartDollyPan() {

            if ( scope.enableZoom ) handleTouchStartDolly();

            if ( scope.enablePan ) handleTouchStartPan();

        }

        function handleTouchStartDollyRotate() {

            if ( scope.enableZoom ) handleTouchStartDolly();

            if ( scope.enableRotate ) handleTouchStartRotate();

        }

        function handleTouchMoveRotate( event ) {

            if ( pointers.length == 1 ) {

                rotateEnd.set( event.pageX, event.pageY );

            } else {

                const position = getSecondPointerPosition( event );

                const x = 0.5 * ( event.pageX + position.x );
                const y = 0.5 * ( event.pageY + position.y );

                rotateEnd.set( x, y );

            }

            rotateDelta.subVectors( rotateEnd, rotateStart ).multiplyScalar( scope.rotateSpeed );

            const element = scope.domElement;

            rotateLeft( 2 * Math.PI * rotateDelta.x / element.clientHeight ); // yes, height

            rotateUp( 2 * Math.PI * rotateDelta.y / element.clientHeight );

            rotateStart.copy( rotateEnd );

        }

        function handleTouchMovePan( event ) {

            if ( pointers.length === 1 ) {

                panEnd.set( event.pageX, event.pageY );

            } else {

                const position = getSecondPointerPosition( event );

                const x = 0.5 * ( event.pageX + position.x );
                const y = 0.5 * ( event.pageY + position.y );

                panEnd.set( x, y );

            }

            panDelta.subVectors( panEnd, panStart ).multiplyScalar( scope.panSpeed );

            pan( panDelta.x, panDelta.y );

            panStart.copy( panEnd );

        }

        function handleTouchMoveDolly( event ) {

            const position = getSecondPointerPosition( event );

            const dx = event.pageX - position.x;
            const dy = event.pageY - position.y;

            const distance = Math.sqrt( dx * dx + dy * dy );

            dollyEnd.set( 0, distance );

            dollyDelta.set( 0, Math.pow( dollyEnd.y / dollyStart.y, scope.zoomSpeed ) );

            dollyOut( dollyDelta.y );

            dollyStart.copy( dollyEnd );

        }

        function handleTouchMoveDollyPan( event ) {

            if ( scope.enableZoom ) handleTouchMoveDolly( event );

            if ( scope.enablePan ) handleTouchMovePan( event );

        }

        function handleTouchMoveDollyRotate( event ) {

            if ( scope.enableZoom ) handleTouchMoveDolly( event );

            if ( scope.enableRotate ) handleTouchMoveRotate( event );

        }

        //
        // event handlers - FSM: listen for events and reset state
        //

        function onPointerDown( event ) {

            if ( scope.enabled === false ) return;

            if ( pointers.length === 0 ) {

                scope.domElement.setPointerCapture( event.pointerId );

                scope.domElement.addEventListener( 'pointermove', onPointerMove );
                scope.domElement.addEventListener( 'pointerup', onPointerUp );

            }

            //

            addPointer( event );

            if ( event.pointerType === 'touch' ) {

                onTouchStart( event );

            } else {

                onMouseDown( event );

            }

        }

        function onPointerMove( event ) {

            if ( scope.enabled === false ) return;

            if ( event.pointerType === 'touch' ) {

                onTouchMove( event );

            } else {

                onMouseMove( event );

            }

        }

        function onPointerUp( event ) {

            removePointer( event );

            if ( pointers.length === 0 ) {

                scope.domElement.releasePointerCapture( event.pointerId );

                scope.domElement.removeEventListener( 'pointermove', onPointerMove );
                scope.domElement.removeEventListener( 'pointerup', onPointerUp );

            }

            scope.dispatchEvent( _endEvent );

            state = STATE.NONE;

        }

        function onMouseDown( event ) {

            let mouseAction;

            switch ( event.button ) {

                case 0:

                    mouseAction = scope.mouseButtons.LEFT;
                    break;

                case 1:

                    mouseAction = scope.mouseButtons.MIDDLE;
                    break;

                case 2:

                    mouseAction = scope.mouseButtons.RIGHT;
                    break;

                default:

                    mouseAction = - 1;

            }

            switch ( mouseAction ) {

                case MOUSE.DOLLY:

                    if ( scope.enableZoom === false ) return;

                    handleMouseDownDolly( event );

                    state = STATE.DOLLY;

                    break;

                case MOUSE.ROTATE:

                    if ( event.ctrlKey || event.metaKey || event.shiftKey ) {

                        if ( scope.enablePan === false ) return;

                        handleMouseDownPan( event );

                        state = STATE.PAN;

                    } else {

                        if ( scope.enableRotate === false ) return;

                        handleMouseDownRotate( event );

                        state = STATE.ROTATE;

                    }

                    break;

                case MOUSE.PAN:

                    if ( event.ctrlKey || event.metaKey || event.shiftKey ) {

                        if ( scope.enableRotate === false ) return;

                        handleMouseDownRotate( event );

                        state = STATE.ROTATE;

                    } else {

                        if ( scope.enablePan === false ) return;

                        handleMouseDownPan( event );

                        state = STATE.PAN;

                    }

                    break;

                default:

                    state = STATE.NONE;

            }

            if ( state !== STATE.NONE ) {

                scope.dispatchEvent( _startEvent );

            }

        }

        function onMouseMove( event ) {

            switch ( state ) {

                case STATE.ROTATE:

                    if ( scope.enableRotate === false ) return;

                    handleMouseMoveRotate( event );

                    break;

                case STATE.DOLLY:

                    if ( scope.enableZoom === false ) return;

                    handleMouseMoveDolly( event );

                    break;

                case STATE.PAN:

                    if ( scope.enablePan === false ) return;

                    handleMouseMovePan( event );

                    break;

            }

        }

        function onMouseWheel( event ) {

            if ( scope.enabled === false || scope.enableZoom === false || state !== STATE.NONE ) return;

            event.preventDefault();

            scope.dispatchEvent( _startEvent );

            handleMouseWheel( event );

            scope.dispatchEvent( _endEvent );

        }

        function onKeyDown( event ) {

            if ( scope.enabled === false || scope.enablePan === false ) return;

            handleKeyDown( event );

        }

        function onTouchStart( event ) {

            trackPointer( event );

            switch ( pointers.length ) {

                case 1:

                    switch ( scope.touches.ONE ) {

                        case TOUCH.ROTATE:

                            if ( scope.enableRotate === false ) return;

                            handleTouchStartRotate();

                            state = STATE.TOUCH_ROTATE;

                            break;

                        case TOUCH.PAN:

                            if ( scope.enablePan === false ) return;

                            handleTouchStartPan();

                            state = STATE.TOUCH_PAN;

                            break;

                        default:

                            state = STATE.NONE;

                    }

                    break;

                case 2:

                    switch ( scope.touches.TWO ) {

                        case TOUCH.DOLLY_PAN:

                            if ( scope.enableZoom === false && scope.enablePan === false ) return;

                            handleTouchStartDollyPan();

                            state = STATE.TOUCH_DOLLY_PAN;

                            break;

                        case TOUCH.DOLLY_ROTATE:

                            if ( scope.enableZoom === false && scope.enableRotate === false ) return;

                            handleTouchStartDollyRotate();

                            state = STATE.TOUCH_DOLLY_ROTATE;

                            break;

                        default:

                            state = STATE.NONE;

                    }

                    break;

                default:

                    state = STATE.NONE;

            }

            if ( state !== STATE.NONE ) {

                scope.dispatchEvent( _startEvent );

            }

        }

        function onTouchMove( event ) {

            trackPointer( event );

            switch ( state ) {

                case STATE.TOUCH_ROTATE:

                    if ( scope.enableRotate === false ) return;

                    handleTouchMoveRotate( event );

                    scope.update();

                    break;

                case STATE.TOUCH_PAN:

                    if ( scope.enablePan === false ) return;

                    handleTouchMovePan( event );

                    scope.update();

                    break;

                case STATE.TOUCH_DOLLY_PAN:

                    if ( scope.enableZoom === false && scope.enablePan === false ) return;

                    handleTouchMoveDollyPan( event );

                    scope.update();

                    break;

                case STATE.TOUCH_DOLLY_ROTATE:

                    if ( scope.enableZoom === false && scope.enableRotate === false ) return;

                    handleTouchMoveDollyRotate( event );

                    scope.update();

                    break;

                default:

                    state = STATE.NONE;

            }

        }

        function onContextMenu( event ) {

            if ( scope.enabled === false ) return;

            event.preventDefault();

        }

        function addPointer( event ) {

            pointers.push( event );

        }

        function removePointer( event ) {

            delete pointerPositions[event.pointerId];

            for ( let i = 0; i < pointers.length; i ++ ) {

                if ( pointers[i].pointerId == event.pointerId ) {

                    pointers.splice( i, 1 );
                    return;

                }

            }

        }

        function trackPointer( event ) {

            let position = pointerPositions[event.pointerId];

            if ( position === undefined ) {

                position = new Vector2();
                pointerPositions[event.pointerId] = position;

            }

            position.set( event.pageX, event.pageY );

        }

        function getSecondPointerPosition( event ) {

            const pointer = ( event.pointerId === pointers[0].pointerId ) ? pointers[1] : pointers[0];

            return pointerPositions[pointer.pointerId];

        }

        //

        scope.domElement.addEventListener( 'contextmenu', onContextMenu );

        scope.domElement.addEventListener( 'pointerdown', onPointerDown );
        scope.domElement.addEventListener( 'pointercancel', onPointerUp );
        scope.domElement.addEventListener( 'wheel', onMouseWheel, { passive: false } );

        // force an update at start

        this.update();

    }

}

class LoadingSpinner {

    constructor(message, container) {
        this.message = message || 'Loading...';
        this.container = container || document.body;

        this.spinnerDivContainerOuter = document.createElement('div');
        this.spinnerDivContainerOuter.className = 'outerContainer';
        this.spinnerDivContainerOuter.style.display = 'none';

        this.spinnerDivContainer = document.createElement('div');
        this.spinnerDivContainer.className = 'container';

        this.spinnerDiv = document.createElement('div');
        this.spinnerDiv.className = 'loader';

        this.messageDiv = document.createElement('div');
        this.messageDiv.className = 'message';
        this.messageDiv.innerHTML = this.message;

        this.spinnerDivContainer.appendChild(this.spinnerDiv);
        this.spinnerDivContainer.appendChild(this.messageDiv);
        this.spinnerDivContainerOuter.appendChild(this.spinnerDivContainer);
        this.container.appendChild(this.spinnerDivContainerOuter);

        const style = document.createElement('style');
        style.innerHTML = `

            .message {
                font-family: arial;
                font-size: 12pt;
                color: #ffffff;
                text-align: center;
                padding-top:15px;
                width: 180px;
            }

            .outerContainer {
                width: 100%;
                height: 100%;
            }

            .container {
                position: absolute;
                top: 50%;
                left: 50%;
                transform: translate(-80px, -80px);
                width: 180px;
            }

            .loader {
                width: 120px;        /* the size */
                padding: 15px;       /* the border thickness */
                background: #07e8d6; /* the color */
                z-index:99999;
            
                aspect-ratio: 1;
                border-radius: 50%;
                --_m: 
                    conic-gradient(#0000,#000),
                    linear-gradient(#000 0 0) content-box;
                -webkit-mask: var(--_m);
                    mask: var(--_m);
                -webkit-mask-composite: source-out;
                    mask-composite: subtract;
                box-sizing: border-box;
                animation: load 1s linear infinite;
                margin-left: 30px;
            }
            
            @keyframes load {
                to{transform: rotate(1turn)}
            }

        `;
        this.spinnerDivContainerOuter.appendChild(style);
    }

    show() {
        this.spinnerDivContainerOuter.style.display = 'block';
    }

    hide() {
        this.spinnerDivContainerOuter.style.display = 'none';
    }

    setContainer(container) {
        if (this.container) {
            this.container.removeChild(this.spinnerDivContainerOuter);
        }
        this.container = container;
        this.container.appendChild(this.spinnerDivContainerOuter);
        this.spinnerDivContainerOuter.style.zIndex = this.container.style.zIndex + 1;
    }

    setMessage(msg) {
        this.messageDiv.innerHTML = msg;
    }
}

class ArrowHelper extends THREE.Object3D {

    constructor(dir = new THREE.Vector3(0, 0, 1), origin = new THREE.Vector3(0, 0, 0), length = 1,
                radius = 0.1, color = 0xffff00, headLength = length * 0.2, headRadius = headLength * 0.2) {
        super();

        this.type = 'ArrowHelper';

        const lineGeometry = new THREE.CylinderGeometry(radius, radius, length, 32);
        lineGeometry.translate(0, length / 2.0, 0);
        const coneGeometry = new THREE.CylinderGeometry( 0, headRadius, headLength, 32);
        coneGeometry.translate(0, length, 0);

        this.position.copy( origin );

        this.line = new THREE.Mesh(lineGeometry, new THREE.MeshBasicMaterial({color: color, toneMapped: false}));
        this.line.matrixAutoUpdate = false;
        this.add(this.line);

        this.cone = new THREE.Mesh(coneGeometry, new THREE.MeshBasicMaterial({color: color, toneMapped: false}));
        this.cone.matrixAutoUpdate = false;
        this.add(this.cone);

        this.setDirection(dir);
    }

    setDirection( dir ) {
        if (dir.y > 0.99999) {
            this.quaternion.set(0, 0, 0, 1);
        } else if (dir.y < - 0.99999) {
            this.quaternion.set(1, 0, 0, 0);
        } else {
            _axis.set(dir.z, 0, -dir.x).normalize();
            const radians = Math.acos(dir.y);
            this.quaternion.setFromAxisAngle(_axis, radians);
        }
    }

    setColor( color ) {
        this.line.material.color.set(color);
        this.cone.material.color.set(color);
    }

    copy(source) {
        super.copy(source, false);
        this.line.copy(source.line);
        this.cone.copy(source.cone);
        return this;
    }

    dispose() {
        this.line.geometry.dispose();
        this.line.material.dispose();
        this.cone.geometry.dispose();
        this.cone.material.dispose();
    }

}

class SceneHelper {

    constructor(threeScene) {
        this.threeScene = threeScene;
        this.splatRenderTarget = null;
        this.renderTargetCopyQuad = null;
        this.renderTargetCopyCamera = null;
        this.meshCursor = null;
        this.focusMarker = null;
        this.controlPlane = null;
        this.debugRoot = null;
        this.secondaryDebugRoot = null;
    }

    updateSplatRenderTargetForRenderDimensions(width, height) {
        this.destroySplatRendertarget();
        this.splatRenderTarget = new THREE.WebGLRenderTarget(width, height, {
            format: THREE.RGBAFormat,
            stencilBuffer: false,
            depthBuffer: true,

        });
        this.splatRenderTarget.depthTexture = new THREE.DepthTexture(width, height);
        this.splatRenderTarget.depthTexture.format = THREE.DepthFormat;
        this.splatRenderTarget.depthTexture.type = THREE.UnsignedIntType;
    }

    destroySplatRendertarget() {
        if (this.splatRenderTarget) {
            this.splatRenderTarget = null;
        }
    }

    setupRenderTargetCopyObjects() {
        const uniforms = {
            'sourceColorTexture': {
                'type': 't',
                'value': null
            },
            'sourceDepthTexture': {
                'type': 't',
                'value': null
            },
        };
        const renderTargetCopyMaterial = new THREE.ShaderMaterial({
            vertexShader: `
                varying vec2 vUv;
                void main() {
                    vUv = uv;
                    gl_Position = vec4( position.xy, 0.0, 1.0 );    
                }
            `,
            fragmentShader: `
                #include <common>
                #include <packing>
                varying vec2 vUv;
                uniform sampler2D sourceColorTexture;
                uniform sampler2D sourceDepthTexture;
                void main() {
                    vec4 color = texture2D(sourceColorTexture, vUv);
                    float fragDepth = texture2D(sourceDepthTexture, vUv).x;
                    gl_FragDepth = fragDepth;
                    gl_FragColor = vec4(color.rgb, color.a * 2.0);
              }
            `,
            uniforms: uniforms,
            depthWrite: false,
            depthTest: false,
            transparent: true,
            blending: THREE.CustomBlending,
            blendSrc: THREE.SrcAlphaFactor,
            blendSrcAlpha: THREE.SrcAlphaFactor,
            blendDst: THREE.OneMinusSrcAlphaFactor,
            blendDstAlpha: THREE.OneMinusSrcAlphaFactor
        });
        renderTargetCopyMaterial.extensions.fragDepth = true;
        this.renderTargetCopyQuad = new THREE.Mesh(new THREE.PlaneGeometry(2, 2), renderTargetCopyMaterial);
        this.renderTargetCopyCamera = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);
    }

    destroyRenderTargetCopyObjects() {
        if (this.renderTargetCopyQuad) {
            disposeAllMeshes(this.renderTargetCopyQuad);
            this.renderTargetCopyQuad = null;
        }
    }

    setupMeshCursor() {
        if (!this.meshCursor) {
            const coneGeometry = new THREE.ConeGeometry(0.5, 1.5, 32);
            const coneMaterial = new THREE.MeshBasicMaterial({color: 0xFFFFFF});

            const downArrow = new THREE.Mesh(coneGeometry, coneMaterial);
            downArrow.rotation.set(0, 0, Math.PI);
            downArrow.position.set(0, 1, 0);
            const upArrow = new THREE.Mesh(coneGeometry, coneMaterial);
            upArrow.position.set(0, -1, 0);
            const leftArrow = new THREE.Mesh(coneGeometry, coneMaterial);
            leftArrow.rotation.set(0, 0, Math.PI / 2.0);
            leftArrow.position.set(1, 0, 0);
            const rightArrow = new THREE.Mesh(coneGeometry, coneMaterial);
            rightArrow.rotation.set(0, 0, -Math.PI / 2.0);
            rightArrow.position.set(-1, 0, 0);

            this.meshCursor = new THREE.Object3D();
            this.meshCursor.add(downArrow);
            this.meshCursor.add(upArrow);
            this.meshCursor.add(leftArrow);
            this.meshCursor.add(rightArrow);
            this.meshCursor.scale.set(0.1, 0.1, 0.1);
            this.threeScene.add(this.meshCursor);
            this.meshCursor.visible = false;
        }
    }

    destroyMeshCursor() {
        if (this.meshCursor) {
            disposeAllMeshes(this.meshCursor);
            this.threeScene.remove(this.meshCursor);
            this.meshCursor = null;
        }
    }

    setMeshCursorVisibility(visible) {
        this.meshCursor.visible = visible;
    }

    setMeshCursorPosition(position) {
        this.meshCursor.position.copy(position);
    }

    positionAndOrientMeshCursor(position, camera) {
        this.meshCursor.position.copy(position);
        this.meshCursor.up.copy(camera.up);
        this.meshCursor.lookAt(camera.position);
    }

    setupFocusMarker() {
        if (!this.focusMarker) {
            const sphereGeometry = new THREE.SphereGeometry(.5, 32, 32);
            const focusMarkerMaterial = SceneHelper.buildFocusMarkerMaterial();
            focusMarkerMaterial.depthTest = false;
            focusMarkerMaterial.depthWrite = false;
            focusMarkerMaterial.transparent = true;
            this.focusMarker = new THREE.Mesh(sphereGeometry, focusMarkerMaterial);
        }
    }

    destroyFocusMarker() {
        if (this.focusMarker) {
            disposeAllMeshes(this.focusMarker);
            this.focusMarker = null;
        }
    }

    updateFocusMarker = function() {

        const tempPosition = new THREE.Vector3();
        const tempMatrix = new THREE.Matrix4();

        return function(position, camera, viewport) {
            tempMatrix.copy(camera.matrixWorld).invert();
            tempPosition.copy(position).applyMatrix4(tempMatrix);
            tempPosition.normalize().multiplyScalar(10);
            tempPosition.applyMatrix4(camera.matrixWorld);
            this.focusMarker.position.copy(tempPosition);
            this.focusMarker.material.uniforms.realFocusPosition.value.copy(position);
            this.focusMarker.material.uniforms.viewport.value.copy(viewport);
            this.focusMarker.material.uniformsNeedUpdate = true;
        };

    }();

    setFocusMarkerVisibility(visible) {
        this.focusMarker.visible = visible;
    }

    setFocusMarkerOpacity(opacity) {
        this.focusMarker.material.uniforms.opacity.value = opacity;
        this.focusMarker.material.uniformsNeedUpdate = true;
    }

    getFocusMarkerOpacity() {
        return this.focusMarker.material.uniforms.opacity.value;
    }

    setupControlPlane() {
        if (!this.controlPlane) {
            const planeGeometry = new THREE.PlaneGeometry(1, 1);
            planeGeometry.rotateX(-Math.PI / 2);
            const planeMaterial = new THREE.MeshBasicMaterial({color: 0xffffff});
            planeMaterial.transparent = true;
            planeMaterial.opacity = 0.6;
            planeMaterial.depthTest = false;
            planeMaterial.depthWrite = false;
            planeMaterial.side = THREE.DoubleSide;
            const planeMesh = new THREE.Mesh(planeGeometry, planeMaterial);

            const arrowDir = new THREE.Vector3(0, 1, 0);
            arrowDir.normalize();
            const arrowOrigin = new THREE.Vector3(0, 0, 0);
            const arrowLength = 0.5;
            const arrowRadius = 0.01;
            const arrowColor = 0x00dd00;
            const arrowHelper = new ArrowHelper(arrowDir, arrowOrigin, arrowLength, arrowRadius, arrowColor, 0.1, 0.03);

            this.controlPlane = new THREE.Object3D();
            this.controlPlane.add(planeMesh);
            this.controlPlane.add(arrowHelper);
        }
    }

    destroyControlPlane() {
        if (this.controlPlane) {
            disposeAllMeshes(this.controlPlane);
            this.controlPlane = null;
        }
    }

    setControlPlaneVisibility(visible) {
        this.controlPlane.visible = visible;
    }

    positionAndOrientControlPlane = function() {

        const tempQuaternion = new THREE.Quaternion();
        const defaultUp = new THREE.Vector3(0, 1, 0);

        return function(position, up) {
            tempQuaternion.setFromUnitVectors(defaultUp, up);
            this.controlPlane.position.copy(position);
            this.controlPlane.quaternion.copy(tempQuaternion);
        };

    }();

    addDebugMeshes() {
        this.debugRoot = this.createDebugMeshes();
        this.secondaryDebugRoot = this.createSecondaryDebugMeshes();
        this.threeScene.add(this.debugRoot);
        this.threeScene.add(this.secondaryDebugRoot);
    }

    destroyDebugMeshes() {
        for (let debugRoot of [this.debugRoot, this.secondaryDebugRoot]) {
            if (debugRoot) {
                disposeAllMeshes(debugRoot);
                this.threeScene.remove(debugRoot);
            }
        }
        this.debugRoot = null;
        this.secondaryDebugRoot = null;
    }

    createDebugMeshes(renderOrder) {
        const sphereGeometry = new THREE.SphereGeometry(1, 32, 32);
        const debugMeshRoot = new THREE.Object3D();

        const createMesh = (color, position) => {
            let sphereMesh = new THREE.Mesh(sphereGeometry, SceneHelper.buildDebugMaterial(color));
            sphereMesh.renderOrder = renderOrder;
            debugMeshRoot.add(sphereMesh);
            sphereMesh.position.fromArray(position);
        };

        createMesh(0xff0000, [-50, 0, 0]);
        createMesh(0xff0000, [50, 0, 0]);
        createMesh(0x00ff00, [0, 0, -50]);
        createMesh(0x00ff00, [0, 0, 50]);
        createMesh(0xffaa00, [5, 0, 5]);

        return debugMeshRoot;
    }

    createSecondaryDebugMeshes(renderOrder) {
        const boxGeometry = new THREE.BoxGeometry(3, 3, 3);
        const debugMeshRoot = new THREE.Object3D();

        let boxColor = 0xBBBBBB;
        const createMesh = (position) => {
            let boxMesh = new THREE.Mesh(boxGeometry, SceneHelper.buildDebugMaterial(boxColor));
            boxMesh.renderOrder = renderOrder;
            debugMeshRoot.add(boxMesh);
            boxMesh.position.fromArray(position);
        };

        let separation = 10;
        createMesh([-separation, 0, -separation]);
        createMesh([-separation, 0, separation]);
        createMesh([separation, 0, -separation]);
        createMesh([separation, 0, separation]);

        return debugMeshRoot;
    }

    static buildDebugMaterial(color) {
        const vertexShaderSource = `
            #include <common>
            varying float ndcDepth;

            void main() {
                gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(position.xyz, 1.0);
                ndcDepth = gl_Position.z / gl_Position.w;
                gl_Position.x = gl_Position.x / gl_Position.w;
                gl_Position.y = gl_Position.y / gl_Position.w;
                gl_Position.z = 0.0;
                gl_Position.w = 1.0;
    
            }
        `;

        const fragmentShaderSource = `
            #include <common>
            uniform vec3 color;
            varying float ndcDepth;
            void main() {
                gl_FragDepth = (ndcDepth + 1.0) / 2.0;
                gl_FragColor = vec4(color.rgb, 0.0);
            }
        `;

        const uniforms = {
            'color': {
                'type': 'v3',
                'value': new THREE.Color(color)
            },
        };

        const material = new THREE.ShaderMaterial({
            uniforms: uniforms,
            vertexShader: vertexShaderSource,
            fragmentShader: fragmentShaderSource,
            transparent: false,
            depthTest: true,
            depthWrite: true,
            side: THREE.FrontSide
        });
        material.extensions.fragDepth = true;

        return material;
    }

    static buildFocusMarkerMaterial(color) {
        const vertexShaderSource = `
            #include <common>

            uniform vec2 viewport;
            uniform vec3 realFocusPosition;

            varying vec4 ndcPosition;
            varying vec4 ndcCenter;
            varying vec4 ndcFocusPosition;

            void main() {
                float radius = 0.01;

                vec4 viewPosition = modelViewMatrix * vec4(position.xyz, 1.0);
                vec4 viewCenter = modelViewMatrix * vec4(0.0, 0.0, 0.0, 1.0);

                vec4 viewFocusPosition = modelViewMatrix * vec4(realFocusPosition, 1.0);

                ndcPosition = projectionMatrix * viewPosition;
                ndcPosition = ndcPosition * vec4(1.0 / ndcPosition.w);
                ndcCenter = projectionMatrix * viewCenter;
                ndcCenter = ndcCenter * vec4(1.0 / ndcCenter.w);

                ndcFocusPosition = projectionMatrix * viewFocusPosition;
                ndcFocusPosition = ndcFocusPosition * vec4(1.0 / ndcFocusPosition.w);

                gl_Position = projectionMatrix * viewPosition;

            }
        `;

        const fragmentShaderSource = `
            #include <common>
            uniform vec3 color;
            uniform vec2 viewport;
            uniform float opacity;

            varying vec4 ndcPosition;
            varying vec4 ndcCenter;
            varying vec4 ndcFocusPosition;

            void main() {
                vec2 screenPosition = vec2(ndcPosition) * viewport;
                vec2 screenCenter = vec2(ndcCenter) * viewport;

                vec2 screenVec = screenPosition - screenCenter;

                float projectedRadius = length(screenVec);

                float lineWidth = 0.0005 * viewport.y;
                float aaRange = 0.0025 * viewport.y;
                float radius = 0.06 * viewport.y;
                float radDiff = abs(projectedRadius - radius) - lineWidth;
                float alpha = 1.0 - clamp(radDiff / 5.0, 0.0, 1.0); 

                gl_FragColor = vec4(color.rgb, alpha * opacity);
            }
        `;

        const uniforms = {
            'color': {
                'type': 'v3',
                'value': new THREE.Color(color)
            },
            'realFocusPosition': {
                'type': 'v3',
                'value': new THREE.Vector3()
            },
            'viewport': {
                'type': 'v2',
                'value': new THREE.Vector2()
            },
            'opacity': {
                'value': 0.0
            }
        };

        const material = new THREE.ShaderMaterial({
            uniforms: uniforms,
            vertexShader: vertexShaderSource,
            fragmentShader: fragmentShaderSource,
            transparent: true,
            depthTest: false,
            depthWrite: false,
            side: THREE.FrontSide
        });

        return material;
    }

    dispose() {
        this.destroyMeshCursor();
        this.destroyFocusMarker();
        this.destroyDebugMeshes();
        this.destroyControlPlane();
        this.destroyRenderTargetCopyObjects();
        this.destroySplatRendertarget();
    }
}

const VectorRight = new THREE.Vector3(1, 0, 0);
const VectorUp = new THREE.Vector3(0, 1, 0);
const VectorBackward = new THREE.Vector3(0, 0, 1);

class Ray {

    constructor(origin = new THREE.Vector3(), direction = new THREE.Vector3()) {
        this.origin = new THREE.Vector3();
        this.direction = new THREE.Vector3();
        this.setParameters(origin, direction);
    }

    setParameters(origin, direction) {
        this.origin.copy(origin);
        this.direction.copy(direction).normalize();
    }

    boxContainsPoint(box, point, epsilon) {
        return point.x < box.min.x - epsilon || point.x > box.max.x + epsilon ||
               point.y < box.min.y - epsilon || point.y > box.max.y + epsilon ||
               point.z < box.min.z - epsilon || point.z > box.max.z + epsilon ? false : true;
    }

    intersectBox = function() {

        const planeIntersectionPoint = new THREE.Vector3();
        const planeIntersectionPointArray = [];
        const originArray = [];
        const directionArray = [];

        return function(box, outHit) {

            originArray[0] = this.origin.x;
            originArray[1] = this.origin.y;
            originArray[2] = this.origin.z;
            directionArray[0] = this.direction.x;
            directionArray[1] = this.direction.y;
            directionArray[2] = this.direction.z;

            if (this.boxContainsPoint(box, this.origin, 0.0001)) {
                if (outHit) {
                    outHit.origin.copy(this.origin);
                    outHit.normal.set(0, 0, 0);
                    outHit.distance = -1;
                }
                return true;
            }

            for (let i = 0; i < 3; i++) {
                if (directionArray[i] == 0.0) continue;

                const hitNormal = i == 0 ? VectorRight : i == 1 ? VectorUp : VectorBackward;
                const extremeVec = directionArray[i] < 0 ? box.max : box.min;
                let multiplier = -Math.sign(directionArray[i]);
                planeIntersectionPointArray[0] = i == 0 ? extremeVec.x : i == 1 ? extremeVec.y : extremeVec.z;
                let toSide = planeIntersectionPointArray[0] - originArray[i];

                if (toSide * multiplier < 0) {
                    const idx1 = (i + 1) % 3;
                    const idx2 = (i + 2) % 3;
                    planeIntersectionPointArray[2] = directionArray[idx1] / directionArray[i] * toSide + originArray[idx1];
                    planeIntersectionPointArray[1] = directionArray[idx2] / directionArray[i] * toSide + originArray[idx2];
                    planeIntersectionPoint.set(planeIntersectionPointArray[i],
                                               planeIntersectionPointArray[idx2],
                                               planeIntersectionPointArray[idx1]);
                    if (this.boxContainsPoint(box, planeIntersectionPoint, 0.0001)) {
                        if (outHit) {
                            outHit.origin.copy(planeIntersectionPoint);
                            outHit.normal.copy(hitNormal).multiplyScalar(multiplier);
                            outHit.distance = planeIntersectionPoint.sub(this.origin).length();
                        }
                        return true;
                    }
                }
            }

            return false;
        };

    }();

    intersectSphere = function() {

        const toSphereCenterVec = new THREE.Vector3();

        return function(center, radius, outHit) {
            toSphereCenterVec.copy(center).sub(this.origin);
            const toClosestApproach = toSphereCenterVec.dot(this.direction);
            const toClosestApproachSq = toClosestApproach * toClosestApproach;
            const toSphereCenterSq = toSphereCenterVec.dot(toSphereCenterVec);
            const diffSq = toSphereCenterSq - toClosestApproachSq;
            const radiusSq = radius * radius;

            if (diffSq > radiusSq) return false;

            const thc = Math.sqrt(radiusSq - diffSq);
            const t0 = toClosestApproach - thc;
            const t1 = toClosestApproach + thc;

            if (t1 < 0) return false;
            let t = t0 < 0 ? t1 : t0;

            if (outHit) {
                outHit.origin.copy(this.origin).addScaledVector(this.direction, t);
                outHit.normal.copy(outHit.origin).sub(center).normalize();
                outHit.distance = t;
            }
            return true;
        };

    }();
}

class Hit {

    constructor() {
        this.origin = new THREE.Vector3();
        this.normal = new THREE.Vector3();
        this.distance = 0;
        this.splatIndex = 0;
    }

    set(origin, normal, distance, splatIndex) {
        this.origin.copy(origin);
        this.normal.copy(normal);
        this.distance = distance;
        this.splatIndex = splatIndex;
    }

    clone() {
        const hitClone = new Hit();
        hitClone.origin.copy(this.origin);
        hitClone.normal.copy(this.normal);
        hitClone.distance = this.distance;
        hitClone.splatIndex = this.splatIndex;
        return hitClone;
    }

}

class Raycaster {

    constructor(origin, direction, raycastAgainstTrueSplatEllipsoid = false) {
        this.ray = new Ray(origin, direction);
        this.raycastAgainstTrueSplatEllipsoid = raycastAgainstTrueSplatEllipsoid;
    }

    setFromCameraAndScreenPosition = function() {

        const ndcCoords = new THREE.Vector2();

        return function(camera, screenPosition, screenDimensions) {
            ndcCoords.x = screenPosition.x / screenDimensions.x * 2.0 - 1.0;
            ndcCoords.y = (screenDimensions.y - screenPosition.y) / screenDimensions.y * 2.0 - 1.0;
            if (camera.isPerspectiveCamera) {
                this.ray.origin.setFromMatrixPosition(camera.matrixWorld);
                this.ray.direction.set(ndcCoords.x, ndcCoords.y, 0.5 ).unproject(camera).sub(this.ray.origin).normalize();
                this.camera = camera;
            } else if (camera.isOrthographicCamera) {
                this.ray.origin.set(screenPosition.x, screenPosition.y,
                                   (camera.near + camera.far) / (camera.near - camera.far)).unproject(camera);
                this.ray.direction.set(0, 0, -1).transformDirection(camera.matrixWorld);
                this.camera = camera;
            } else {
                throw new Error('Raycaster::setFromCameraAndScreenPosition() -> Unsupported camera type');
            }
        };

    }();

    intersectSplatMesh = function() {

        const toLocal = new THREE.Matrix4();
        const fromLocal = new THREE.Matrix4();
        const sceneTransform = new THREE.Matrix4();
        const localRay = new Ray();
        const tempPoint = new THREE.Vector3();

        return function(splatMesh, outHits = []) {
            const splatTree = splatMesh.getSplatTree();

            for (let s = 0; s < splatTree.subTrees.length; s++) {
                const subTree = splatTree.subTrees[s];

                fromLocal.copy(splatMesh.matrixWorld);
                splatMesh.getSceneTransform(s, sceneTransform);
                fromLocal.multiply(sceneTransform);
                toLocal.copy(fromLocal).invert();

                localRay.origin.copy(this.ray.origin).applyMatrix4(toLocal);
                localRay.direction.copy(this.ray.origin).add(this.ray.direction);
                localRay.direction.applyMatrix4(toLocal).sub(localRay.origin).normalize();

                const outHitsForSubTree = [];
                if (subTree.rootNode) {
                    this.castRayAtSplatTreeNode(localRay, splatTree, subTree.rootNode, outHitsForSubTree);
                }

                outHitsForSubTree.forEach((hit) => {
                    hit.origin.applyMatrix4(fromLocal);
                    hit.normal.applyMatrix4(fromLocal).normalize();
                    hit.distance = tempPoint.copy(hit.origin).sub(this.ray.origin).length();
                });

                outHits.push(...outHitsForSubTree);
            }

            outHits.sort((a, b) => {
                if (a.distance > b.distance) return 1;
                else return -1;
            });

            return outHits;
        };

    }();

    castRayAtSplatTreeNode = function() {

        const tempColor = new THREE.Vector4();
        const tempCenter = new THREE.Vector3();
        const tempScale = new THREE.Vector3();
        const tempRotation = new THREE.Quaternion();
        const tempHit = new Hit();
        const scaleEpsilon = 0.0000001;

        const origin = new THREE.Vector3(0, 0, 0);
        const uniformScaleMatrix = new THREE.Matrix4();
        const scaleMatrix = new THREE.Matrix4();
        const rotationMatrix = new THREE.Matrix4();
        const toSphereSpace = new THREE.Matrix4();
        const fromSphereSpace = new THREE.Matrix4();
        const tempRay = new Ray();

        return function(ray, splatTree, node, outHits = []) {
            if (!ray.intersectBox(node.boundingBox)) {
                return;
            }
            if (node.data.indexes && node.data.indexes.length > 0) {
                for (let i = 0; i < node.data.indexes.length; i++) {
                    const splatGlobalIndex = node.data.indexes[i];
                    splatTree.splatMesh.getSplatColor(splatGlobalIndex, tempColor, false);
                    splatTree.splatMesh.getSplatCenter(splatGlobalIndex, tempCenter, false);
                    splatTree.splatMesh.getSplatScaleAndRotation(splatGlobalIndex, tempScale, tempRotation, false);

                    if (tempScale.x <= scaleEpsilon || tempScale.y <= scaleEpsilon || tempScale.z <= scaleEpsilon) {
                        continue;
                    }

                    if (!this.raycastAgainstTrueSplatEllipsoid) {
                        const radius = (tempScale.x + tempScale.y + tempScale.z) / 3;
                        if (ray.intersectSphere(tempCenter, radius, tempHit)) {
                            const hitClone = tempHit.clone();
                            hitClone.splatIndex = splatGlobalIndex;
                            outHits.push(hitClone);
                        }
                    } else {
                        scaleMatrix.makeScale(tempScale.x, tempScale.y, tempScale.z);
                        rotationMatrix.makeRotationFromQuaternion(tempRotation);
                        const uniformScale = Math.log10(tempColor.w) * 2.0;
                        uniformScaleMatrix.makeScale(uniformScale, uniformScale, uniformScale);
                        fromSphereSpace.copy(uniformScaleMatrix).multiply(rotationMatrix).multiply(scaleMatrix);
                        toSphereSpace.copy(fromSphereSpace).invert();
                        tempRay.origin.copy(ray.origin).sub(tempCenter).applyMatrix4(toSphereSpace);
                        tempRay.direction.copy(ray.origin).add(ray.direction).sub(tempCenter);
                        tempRay.direction.applyMatrix4(toSphereSpace).sub(tempRay.origin).normalize();
                        if (tempRay.intersectSphere(origin, 1.0, tempHit)) {
                            const hitClone = tempHit.clone();
                            hitClone.splatIndex = splatGlobalIndex;
                            hitClone.origin.applyMatrix4(fromSphereSpace).add(tempCenter);
                            outHits.push(hitClone);
                        }
                    }
                }
             }
            if (node.children && node.children.length > 0) {
                for (let child of node.children) {
                    this.castRayAtSplatTreeNode(ray, splatTree, child, outHits);
                }
            }
            return outHits;
        };

    }();
}

/**
 * SplatScene: Descriptor for a single splat scene managed by an instance of SplatMesh.
 */
class SplatScene {

    constructor(splatBuffer, position = new THREE.Vector3(), quaternion = new THREE.Quaternion(), scale = new THREE.Vector3(1, 1, 1)) {
        this.splatBuffer = splatBuffer;
        this.position = position.clone();
        this.quaternion = quaternion.clone();
        this.scale = scale.clone();
        this.transform = new THREE.Matrix4();
        this.updateTransform();
    }

    copyTransformData(otherScene) {
        this.position.copy(otherScene.position);
        this.quaternion.copy(otherScene.quaternion);
        this.scale.copy(otherScene.scale);
        this.transform.copy(otherScene.transform);
    }

    updateTransform() {
        this.transform.compose(this.position, this.quaternion, this.scale);
    }
}

let idGen = 0;

class SplatTreeNode {

    constructor(min, max, depth, id) {
        this.min = new THREE.Vector3().copy(min);
        this.max = new THREE.Vector3().copy(max);
        this.boundingBox = new THREE.Box3(this.min, this.max);
        this.center = new THREE.Vector3().copy(this.max).sub(this.min).multiplyScalar(0.5).add(this.min);
        this.depth = depth;
        this.children = [];
        this.data = null;
        this.id = id || idGen++;
    }

}

class SplatSubTree {

    constructor(maxDepth, maxCentersPerNode) {
        this.maxDepth = maxDepth;
        this.maxCentersPerNode = maxCentersPerNode;
        this.sceneDimensions = new THREE.Vector3();
        this.sceneMin = new THREE.Vector3();
        this.sceneMax = new THREE.Vector3();
        this.splatMesh = null;
        this.rootNode = null;
        this.addedIndexes = {};
        this.nodesWithIndexes = [];
    }

}

/**
 * SplatTree: Octree tailored to splat data from a SplatMesh instance
 */
class SplatTree {

    constructor(maxDepth, maxCentersPerNode) {
        this.maxDepth = maxDepth;
        this.maxCentersPerNode = maxCentersPerNode;
        this.splatMesh = null;
        this.subTrees = [];
    }

    processSplatMesh(splatMesh, filterFunc = () => true) {
        this.splatMesh = splatMesh;
        this.subTrees = [];
        const center = new THREE.Vector3();

        const buildSubTree = function(splatOffset, splatCount, maxDepth, maxCentersPerNode) {
            const subTree = new SplatSubTree(maxDepth, maxCentersPerNode);
            let validSplatCount = 0;
            const indexes = [];
            for (let i = 0; i < splatCount; i++) {
                const globalSplatIndex = i + splatOffset;
                if (filterFunc(globalSplatIndex)) {
                    splatMesh.getSplatCenter(globalSplatIndex, center);
                    if (validSplatCount === 0 || center.x < subTree.sceneMin.x) subTree.sceneMin.x = center.x;
                    if (validSplatCount === 0 || center.x > subTree.sceneMax.x) subTree.sceneMax.x = center.x;
                    if (validSplatCount === 0 || center.y < subTree.sceneMin.y) subTree.sceneMin.y = center.y;
                    if (validSplatCount === 0 || center.y > subTree.sceneMax.y) subTree.sceneMax.y = center.y;
                    if (validSplatCount === 0 || center.z < subTree.sceneMin.z) subTree.sceneMin.z = center.z;
                    if (validSplatCount === 0 || center.z > subTree.sceneMax.z) subTree.sceneMax.z = center.z;
                    validSplatCount++;
                    indexes.push(globalSplatIndex);
                }
            }

            subTree.sceneDimensions.copy(subTree.sceneMax).sub(subTree.sceneMin);

            subTree.rootNode = new SplatTreeNode(subTree.sceneMin, subTree.sceneMax, 0);
            subTree.rootNode.data = {
                'indexes': indexes
            };

            return subTree;
        };

        if (splatMesh.dynamicMode) {
            let splatOffset = 0;
            for (let s = 0; s < splatMesh.scenes.length; s++) {
                const scene = splatMesh.getScene(s);
                const splatCount = scene.splatBuffer.getSplatCount();
                const subTree = buildSubTree(splatOffset, splatCount, this.maxDepth, this.maxCentersPerNode);
                this.subTrees[s] = subTree;
                SplatTree.processNode(subTree, subTree.rootNode, splatMesh);
                splatOffset += splatCount;
            }
        } else {
            const subTree = buildSubTree(0, splatMesh.getSplatCount(), this.maxDepth, this.maxCentersPerNode);
            this.subTrees[0] = subTree;
            SplatTree.processNode(subTree, subTree.rootNode, splatMesh);
        }
    }

    static processNode(tree, node, splatMesh) {
        const splatCount = node.data.indexes.length;

        if (splatCount < tree.maxCentersPerNode || node.depth > tree.maxDepth) {
            const newIndexes = [];
            for (let i = 0; i < node.data.indexes.length; i++) {
                if (!tree.addedIndexes[node.data.indexes[i]]) {
                    newIndexes.push(node.data.indexes[i]);
                    tree.addedIndexes[node.data.indexes[i]] = true;
                }
            }
            node.data.indexes = newIndexes;
            tree.nodesWithIndexes.push(node);
            return;
        }

        const nodeDimensions = new THREE.Vector3().copy(node.max).sub(node.min);
        const halfDimensions = new THREE.Vector3().copy(nodeDimensions).multiplyScalar(0.5);

        const nodeCenter = new THREE.Vector3().copy(node.min).add(halfDimensions);

        const childrenBounds = [
            // top section, clockwise from upper-left (looking from above, +Y)
            new THREE.Box3(new THREE.Vector3(nodeCenter.x - halfDimensions.x, nodeCenter.y, nodeCenter.z - halfDimensions.z),
                           new THREE.Vector3(nodeCenter.x, nodeCenter.y + halfDimensions.y, nodeCenter.z)),
            new THREE.Box3(new THREE.Vector3(nodeCenter.x, nodeCenter.y, nodeCenter.z - halfDimensions.z),
                           new THREE.Vector3(nodeCenter.x + halfDimensions.x, nodeCenter.y + halfDimensions.y, nodeCenter.z)),
            new THREE.Box3(new THREE.Vector3(nodeCenter.x, nodeCenter.y, nodeCenter.z),
                           new THREE.Vector3(nodeCenter.x + halfDimensions.x,
                                             nodeCenter.y + halfDimensions.y, nodeCenter.z + halfDimensions.z)),
            new THREE.Box3(new THREE.Vector3(nodeCenter.x - halfDimensions.x, nodeCenter.y, nodeCenter.z ),
                           new THREE.Vector3(nodeCenter.x, nodeCenter.y + halfDimensions.y, nodeCenter.z + halfDimensions.z)),

            // bottom section, clockwise from lower-left (looking from above, +Y)
            new THREE.Box3(new THREE.Vector3(nodeCenter.x - halfDimensions.x,
                                             nodeCenter.y - halfDimensions.y, nodeCenter.z - halfDimensions.z),
                           new THREE.Vector3(nodeCenter.x, nodeCenter.y, nodeCenter.z)),
            new THREE.Box3(new THREE.Vector3(nodeCenter.x, nodeCenter.y - halfDimensions.y, nodeCenter.z - halfDimensions.z),
                           new THREE.Vector3(nodeCenter.x + halfDimensions.x, nodeCenter.y, nodeCenter.z)),
            new THREE.Box3(new THREE.Vector3(nodeCenter.x, nodeCenter.y - halfDimensions.y, nodeCenter.z),
                           new THREE.Vector3(nodeCenter.x + halfDimensions.x, nodeCenter.y, nodeCenter.z + halfDimensions.z)),
            new THREE.Box3(new THREE.Vector3(nodeCenter.x - halfDimensions.x, nodeCenter.y - halfDimensions.y, nodeCenter.z),
                           new THREE.Vector3(nodeCenter.x, nodeCenter.y, nodeCenter.z + halfDimensions.z)),
        ];

        const splatCounts = [];
        const baseIndexes = [];
        for (let i = 0; i < childrenBounds.length; i++) {
            splatCounts[i] = 0;
            baseIndexes[i] = [];
        }

        const center = new THREE.Vector3();
        for (let i = 0; i < splatCount; i++) {
            const splatGlobalIndex = node.data.indexes[i];
            splatMesh.getSplatCenter(splatGlobalIndex, center);
            for (let j = 0; j < childrenBounds.length; j++) {
                if (childrenBounds[j].containsPoint(center)) {
                    splatCounts[j]++;
                    baseIndexes[j].push(splatGlobalIndex);
                }
            }
        }

        for (let i = 0; i < childrenBounds.length; i++) {
            const childNode = new SplatTreeNode(childrenBounds[i].min, childrenBounds[i].max, node.depth + 1);
            childNode.data = {
                'indexes': baseIndexes[i]
            };
            node.children.push(childNode);
        }

        node.data = {};
        for (let child of node.children) {
            SplatTree.processNode(tree, child, splatMesh);
        }
    }

    countLeaves() {

        let leafCount = 0;
        this.visitLeaves(() => {
            leafCount++;
        });

        return leafCount;
    }

    visitLeaves(visitFunc) {

        const visitLeavesFromNode = (node, visitFunc) => {
            if (node.children.length === 0) visitFunc(node);
            for (let child of node.children) {
                visitLeavesFromNode(child, visitFunc);
            }
        };

        for (let subTree of this.subTrees) {
            visitLeavesFromNode(subTree.rootNode, visitFunc);
        }
    }

}

class Constants {

    static DepthMapRange = 1 << 16;
    static MemoryPageSize = 65536;
    static BytesPerFloat = 4;
    static BytesPerInt = 4;
    static MaxScenes = 32;

}

const dummyGeometry = new THREE.BufferGeometry();
const dummyMaterial = new THREE.MeshBasicMaterial();

/**
 * SplatMesh: Container for one or more splat scenes, abstracting them into a single unified container for
 * splat data. Additionally contains data structures and code to make the splat data renderable as a Three.js mesh.
 */
class SplatMesh extends THREE.Mesh {

    constructor(dynamicMode = true, halfPrecisionCovariancesOnGPU = false, devicePixelRatio = 1,
                enableDistancesComputationOnGPU = true, integerBasedDistancesComputation = false) {
        super(dummyGeometry, dummyMaterial);
        // Reference to a Three.js renderer
        this.renderer = undefined;
        // Use 16-bit floating point values when storing splat covariance data in textures, instead of 32-bit
        this.halfPrecisionCovariancesOnGPU = halfPrecisionCovariancesOnGPU;
        // When 'dynamicMode' is true, scenes are assumed to be non-static. Dynamic scenes are handled differently
        // and certain optimizations cannot be made for them. Additionally, by default, all splat data retrieved from
        // this splat mesh will not have their scene transform applied to them if the splat mesh is dynamic. That
        // can be overriden via parameters to the individual functions that are used to retrieve splat data.
        this.dynamicMode = dynamicMode;
        // Ratio of the resolution in physical pixels to the resolution in CSS pixels for the current display device
        this.devicePixelRatio = devicePixelRatio;
        // Use a transform feedback to calculate splat distances from the camera
        this.enableDistancesComputationOnGPU = enableDistancesComputationOnGPU;
        // Use a faster integer-based approach for calculating splat distances from the camera
        this.integerBasedDistancesComputation = integerBasedDistancesComputation;
        // The individual splat scenes stored in this splat mesh, each containing their own transform
        this.scenes = [];
        // Special octree tailored to SplatMesh instances
        this.splatTree = null;
        // Textures in which splat data will be stored for rendering
        this.splatDataTextures = null;
        this.distancesTransformFeedback = {
            'id': null,
            'vertexShader': null,
            'fragmentShader': null,
            'program': null,
            'centersBuffer': null,
            'transformIndexesBuffer': null,
            'outDistancesBuffer': null,
            'centersLoc': -1,
            'modelViewProjLoc': -1,
            'transformIndexesLoc': -1,
            'transformsLocs': []
        };
        this.globalSplatIndexToLocalSplatIndexMap = [];
        this.globalSplatIndexToSceneIndexMap = [];
    }

    /**
     * Build the Three.js material that is used to render the splats.
     * @param {number} dynamicMode If true, it means the scene geometry represented by this splat mesh is not stationary or
     *                             that the splat count might change
     * @return {THREE.ShaderMaterial}
     */
    static buildMaterial(dynamicMode = false) {

        // Contains the code to project 3D covariance to 2D and from there calculate the quad (using the eigen vectors of the
        // 2D covariance) that is ultimately rasterized
        let vertexShaderSource = `
            precision highp float;
            #include <common>

            attribute uint splatIndex;

            uniform highp sampler2D covariancesTexture;
            uniform highp usampler2D centersColorsTexture;`;

        if (dynamicMode) {
            vertexShaderSource += `
                uniform highp usampler2D transformIndexesTexture;
                uniform highp mat4 transforms[${Constants.MaxScenes}];
                uniform vec2 transformIndexesTextureSize;
            `;
        }

        vertexShaderSource += `
            uniform vec2 focal;
            uniform vec2 viewport;
            uniform vec2 basisViewport;
            uniform vec2 covariancesTextureSize;
            uniform vec2 centersColorsTextureSize;

            varying vec4 vColor;
            varying vec2 vUv;

            varying vec2 vPosition;

            const float sqrt8 = sqrt(8.0);

            const vec4 encodeNorm4 = vec4(1.0 / 255.0, 1.0 / 255.0, 1.0 / 255.0, 1.0 / 255.0);
            const uvec4 mask4 = uvec4(uint(0x000000FF), uint(0x0000FF00), uint(0x00FF0000), uint(0xFF000000));
            const uvec4 shift4 = uvec4(0, 8, 16, 24);
            vec4 uintToRGBAVec (uint u) {
               uvec4 urgba = mask4 & u;
               urgba = urgba >> shift4;
               vec4 rgba = vec4(urgba) * encodeNorm4;
               return rgba;
            }

            vec2 getDataUV(in int stride, in int offset, in vec2 dimensions) {
                vec2 samplerUV = vec2(0.0, 0.0);
                float d = float(splatIndex * uint(stride) + uint(offset)) / dimensions.x;
                samplerUV.y = float(floor(d)) / dimensions.y;
                samplerUV.x = fract(d);
                return samplerUV;
            }

            void main () {

                uvec4 sampledCenterColor = texture(centersColorsTexture, getDataUV(1, 0, centersColorsTextureSize));
                vec3 splatCenter = uintBitsToFloat(uvec3(sampledCenterColor.gba));`;

            if (dynamicMode) {
                vertexShaderSource += `
                    uint transformIndex = texture(transformIndexesTexture, getDataUV(1, 0, transformIndexesTextureSize)).r;
                    mat4 transform = transforms[transformIndex];
                    mat4 transformModelViewMatrix = modelViewMatrix * transform;
                `;
            } else {
                vertexShaderSource += `mat4 transformModelViewMatrix = modelViewMatrix;`;
            }

            vertexShaderSource += `
                vec4 viewCenter = transformModelViewMatrix * vec4(splatCenter, 1.0);

                vec4 clipCenter = projectionMatrix * viewCenter;

                float clip = 1.2 * clipCenter.w;
                if (clipCenter.z < -clip || clipCenter.x < -clip || clipCenter.x > clip || clipCenter.y < -clip || clipCenter.y > clip) {
                    gl_Position = vec4(0.0, 0.0, 2.0, 1.0);
                    return;
                }

                vPosition = position.xy;
                vColor = uintToRGBAVec(sampledCenterColor.r);

                vec2 sampledCovarianceA = texture(covariancesTexture, getDataUV(3, 0, covariancesTextureSize)).rg;
                vec2 sampledCovarianceB = texture(covariancesTexture, getDataUV(3, 1, covariancesTextureSize)).rg;
                vec2 sampledCovarianceC = texture(covariancesTexture, getDataUV(3, 2, covariancesTextureSize)).rg;

                vec3 cov3D_M11_M12_M13 = vec3(sampledCovarianceA.rg, sampledCovarianceB.r);
                vec3 cov3D_M22_M23_M33 = vec3(sampledCovarianceB.g, sampledCovarianceC.rg);

                // Construct the 3D covariance matrix
                mat3 Vrk = mat3(
                    cov3D_M11_M12_M13.x, cov3D_M11_M12_M13.y, cov3D_M11_M12_M13.z,
                    cov3D_M11_M12_M13.y, cov3D_M22_M23_M33.x, cov3D_M22_M23_M33.y,
                    cov3D_M11_M12_M13.z, cov3D_M22_M23_M33.y, cov3D_M22_M23_M33.z
                );

                // Construct the Jacobian of the affine approximation of the projection matrix. It will be used to transform the
                // 3D covariance matrix instead of using the actual projection matrix because that transformation would
                // require a non-linear component (perspective division) which would yield a non-gaussian result. (This assumes
                // the current projection is a perspective projection).
                float s = 1.0 / (viewCenter.z * viewCenter.z);
                mat3 J = mat3(
                    focal.x / viewCenter.z, 0., -(focal.x * viewCenter.x) * s,
                    0., focal.y / viewCenter.z, -(focal.y * viewCenter.y) * s,
                    0., 0., 0.
                );

                // Concatenate the projection approximation with the model-view transformation
                mat3 W = transpose(mat3(transformModelViewMatrix));
                mat3 T = W * J;

                // Transform the 3D covariance matrix (Vrk) to compute the 2D covariance matrix
                mat3 cov2Dm = transpose(T) * Vrk * T;

                cov2Dm[0][0] += 0.3;
                cov2Dm[1][1] += 0.3;

                // We are interested in the upper-left 2x2 portion of the projected 3D covariance matrix because
                // we only care about the X and Y values. We want the X-diagonal, cov2Dm[0][0],
                // the Y-diagonal, cov2Dm[1][1], and the correlation between the two cov2Dm[0][1]. We don't
                // need cov2Dm[1][0] because it is a symetric matrix.
                vec3 cov2Dv = vec3(cov2Dm[0][0], cov2Dm[0][1], cov2Dm[1][1]);

                vec3 ndcCenter = clipCenter.xyz / clipCenter.w;

                // We now need to solve for the eigen-values and eigen vectors of the 2D covariance matrix
                // so that we can determine the 2D basis for the splat. This is done using the method described
                // here: https://people.math.harvard.edu/~knill/teaching/math21b2004/exhibits/2dmatrices/index.html
                // After calculating the eigen-values and eigen-vectors, we calculate the basis for rendering the splat
                // by normalizing the eigen-vectors and then multiplying them by (sqrt(8) * eigen-value), which is
                // equal to scaling them by sqrt(8) standard deviations.
                //
                // This is a different approach than in the original work at INRIA. In that work they compute the
                // max extents of the projected splat in screen space to form a screen-space aligned bounding rectangle
                // which forms the geometry that is actually rasterized. The dimensions of that bounding box are 3.0
                // times the maximum eigen-value, or 3 standard deviations. They then use the inverse 2D covariance
                // matrix (called 'conic') in the CUDA rendering thread to determine fragment opacity by calculating the
                // full gaussian: exp(-0.5 * (X - mean) * conic * (X - mean)) * splat opacity
                float a = cov2Dv.x;
                float d = cov2Dv.z;
                float b = cov2Dv.y;
                float D = a * d - b * b;
                float trace = a + d;
                float traceOver2 = 0.5 * trace;
                float term2 = sqrt(max(0.1f, traceOver2 * traceOver2 - D));
                float eigenValue1 = traceOver2 + term2;
                float eigenValue2 = traceOver2 - term2;

                float transparentAdjust = step(1.0 / 255.0, vColor.a);
                eigenValue2 = eigenValue2 * transparentAdjust; // hide splat if alpha is zero

                vec2 eigenVector1 = normalize(vec2(b, eigenValue1 - a));
                // since the eigen vectors are orthogonal, we derive the second one from the first
                vec2 eigenVector2 = vec2(eigenVector1.y, -eigenVector1.x);

                // We use sqrt(8) standard deviations instead of 3 to eliminate more of the splat with a very low opacity.
                vec2 basisVector1 = eigenVector1 * sqrt8 * sqrt(eigenValue1);
                vec2 basisVector2 = eigenVector2 * sqrt8 * sqrt(eigenValue2);

                vec2 ndcOffset = vec2(vPosition.x * basisVector1 + vPosition.y * basisVector2) * basisViewport * 2.0;

                // Similarly scale the position data we send to the fragment shader
                vPosition *= sqrt8;

                gl_Position = vec4(ndcCenter.xy  + ndcOffset, ndcCenter.z, 1.0);

            }`;

        const fragmentShaderSource = `
            precision highp float;
            #include <common>
 
            uniform vec3 debugColor;

            varying vec4 vColor;
            varying vec2 vUv;

            varying vec2 vPosition;

            void main () {
                // Compute the positional squared distance from the center of the splat to the current fragment.
                float A = dot(vPosition, vPosition);
                // Since the positional data in vPosition has been scaled by sqrt(8), the squared result will be
                // scaled by a factor of 8. If the squared result is larger than 8, it means it is outside the ellipse
                // defined by the rectangle formed by vPosition. It also means it's farther
                // away than sqrt(8) standard deviations from the mean.
                if (A > 8.0) discard;
                vec3 color = vColor.rgb;

                // Since the rendered splat is scaled by sqrt(8), the inverse covariance matrix that is part of
                // the gaussian formula becomes the identity matrix. We're then left with (X - mean) * (X - mean),
                // and since 'mean' is zero, we have X * X, which is the same as A:
                float opacity = exp(-0.5 * A) * vColor.a;

                gl_FragColor = vec4(color.rgb, opacity);
            }`;

        const uniforms = {
            'covariancesTexture': {
                'type': 't',
                'value': null
            },
            'centersColorsTexture': {
                'type': 't',
                'value': null
            },
            'focal': {
                'type': 'v2',
                'value': new THREE.Vector2()
            },
            'viewport': {
                'type': 'v2',
                'value': new THREE.Vector2()
            },
            'basisViewport': {
                'type': 'v2',
                'value': new THREE.Vector2()
            },
            'debugColor': {
                'type': 'v3',
                'value': new THREE.Color()
            },
            'covariancesTextureSize': {
                'type': 'v2',
                'value': new THREE.Vector2(1024, 1024)
            },
            'centersColorsTextureSize': {
                'type': 'v2',
                'value': new THREE.Vector2(1024, 1024)
            }
        };

        if (dynamicMode) {
            uniforms['transformIndexesTexture'] = {
                'type': 't',
                'value': null
            };
            const transformMatrices = [];
            for (let i = 0; i < Constants.MaxScenes; i++) {
                transformMatrices.push(new THREE.Matrix4());
            }
            uniforms['transforms'] = {
                'type': 'mat4',
                'value': transformMatrices
            };
            uniforms['transformIndexesTextureSize'] = {
                'type': 'v2',
                'value': new THREE.Vector2(1024, 1024)
            };
        }

        const material = new THREE.ShaderMaterial({
            uniforms: uniforms,
            vertexShader: vertexShaderSource,
            fragmentShader: fragmentShaderSource,
            transparent: true,
            alphaTest: 1.0,
            blending: THREE.NormalBlending,
            depthTest: true,
            depthWrite: false,
            side: THREE.DoubleSide
        });

        return material;
    }

    /**
     * Build the Three.js geometry that will be used to render the splats. The geometry is instanced and is made up of
     * vertices for a single quad as well as an attribute buffer for the splat indexes.
     * @param {number} maxSplatCount The maximum number of splats that the geometry will need to accomodate
     * @return {THREE.InstancedBufferGeometry}
     */
    static buildGeomtery(maxSplatCount) {

        const baseGeometry = new THREE.BufferGeometry();
        baseGeometry.setIndex([0, 1, 2, 0, 2, 3]);

        // Vertices for the instanced quad
        const positionsArray = new Float32Array(4 * 3);
        const positions = new THREE.BufferAttribute(positionsArray, 3);
        baseGeometry.setAttribute('position', positions);
        positions.setXYZ(0, -1.0, -1.0, 0.0);
        positions.setXYZ(1, -1.0, 1.0, 0.0);
        positions.setXYZ(2, 1.0, 1.0, 0.0);
        positions.setXYZ(3, 1.0, -1.0, 0.0);
        positions.needsUpdate = true;

        const geometry = new THREE.InstancedBufferGeometry().copy(baseGeometry);

        // Splat index buffer
        const splatIndexArray = new Uint32Array(maxSplatCount);
        const splatIndexes = new THREE.InstancedBufferAttribute(splatIndexArray, 1, false);
        splatIndexes.setUsage(THREE.DynamicDrawUsage);
        geometry.setAttribute('splatIndex', splatIndexes);

        geometry.instanceCount = maxSplatCount;

        return geometry;
    }

    /**
     * Build a container for each scene managed by this splat mesh based on an instance of SplatBuffer, along with optional
     * transform data (position, scale, rotation) passed to the splat mesh during the build process.
     * @param {Array<THREE.Matrix4>} splatBuffers SplatBuffer instances containing splats for each scene
     * @param {Array<object>} sceneOptions Array of options objects: {
     *
     *         position (Array<number>):   Position of the scene, acts as an offset from its default position, defaults to [0, 0, 0]
     *
     *         rotation (Array<number>):   Rotation of the scene represented as a quaternion, defaults to [0, 0, 0, 1]
     *
     *         scale (Array<number>):      Scene's scale, defaults to [1, 1, 1]
     * }
     * @return {Array<THREE.Matrix4>}
     */
    static buildScenes(splatBuffers, sceneOptions) {
        const scenes = [];
        scenes.length = splatBuffers.length;
        for (let i = 0; i < splatBuffers.length; i++) {
            const splatBuffer = splatBuffers[i];
            const options = sceneOptions[i] || {};
            let positionArray = options['position'] || [0, 0, 0];
            let rotationArray = options['rotation'] || [0, 0, 0, 1];
            let scaleArray = options['scale'] || [1, 1, 1];
            const position = new THREE.Vector3().fromArray(positionArray);
            const rotation = new THREE.Quaternion().fromArray(rotationArray);
            const scale = new THREE.Vector3().fromArray(scaleArray);
            scenes[i] = SplatMesh.createScene(splatBuffer, position, rotation, scale);
        }
        return scenes;
    }

    static createScene(splatBuffer, position, rotation, scale) {
        return new SplatScene(splatBuffer, position, rotation, scale);
    }

    /**
     * Build data structures that map global splat indexes (based on a unified index across all splat buffers) to
     * local data within a single scene.
     * @param {Array<SplatBuffer>} splatBuffers Instances of SplatBuffer off which to build the maps
     * @return {object}
     */
    static buildSplatIndexMaps(splatBuffers) {
        const localSplatIndexMap = [];
        const sceneIndexMap = [];
        let totalSplatCount = 0;
        for (let s = 0; s < splatBuffers.length; s++) {
            const splatBuffer = splatBuffers[s];
            const splatCount = splatBuffer.getSplatCount();
            for (let i = 0; i < splatCount; i++) {
                localSplatIndexMap[totalSplatCount] = i;
                sceneIndexMap[totalSplatCount] = s;
                totalSplatCount++;
            }
        }
        return {
            localSplatIndexMap,
            sceneIndexMap
        };
    }

    /**
     * Build an instance of SplatTree (a specialized octree) for the given splat mesh.
     * @param {SplatMesh} splatMesh SplatMesh instance for which the splat tree will be built
     * @param {Array<number>} minAlphas Array of minimum splat slphas for each scene
     * @return {SplatTree}
     */
    static buildSplatTree(splatMesh, minAlphas = []) {
        // TODO: expose SplatTree constructor parameters (maximumDepth and maxCentersPerNode) so that they can
        // be configured on a per-scene basis
        const splatTree = new SplatTree(8, 1000);
        console.time('SplatTree build');
        const splatColor = new THREE.Vector4();
        splatTree.processSplatMesh(splatMesh, (splatIndex) => {
            splatMesh.getSplatColor(splatIndex, splatColor);
            const sceneIndex = splatMesh.getSceneIndexForSplat(splatIndex);
            const minAlpha = minAlphas[sceneIndex] || 1;
            return splatColor.w >= minAlpha;
        });
        console.timeEnd('SplatTree build');

        let leavesWithVertices = 0;
        let avgSplatCount = 0;
        let maxSplatCount = 0;
        let nodeCount = 0;

        splatTree.visitLeaves((node) => {
            const nodeSplatCount = node.data.indexes.length;
            if (nodeSplatCount > 0) {
                avgSplatCount += nodeSplatCount;
                maxSplatCount = Math.max(maxSplatCount, nodeSplatCount);
                nodeCount++;
                leavesWithVertices++;
            }
        });
        console.log(`SplatTree leaves: ${splatTree.countLeaves()}`);
        console.log(`SplatTree leaves with splats:${leavesWithVertices}`);
        avgSplatCount = avgSplatCount / nodeCount;
        console.log(`Avg splat count per node: ${avgSplatCount}`);
        console.log(`Total splat count: ${splatMesh.getSplatCount()}`);
        return splatTree;
    }

    /**
     * Construct this instance of SplatMesh.
     * @param {Array<SplatBuffer>} splatBuffers The base splat data, instances of SplatBuffer
     * @param {Array<object>} sceneOptions Dynamic options for each scene {
     *
     *         splatAlphaRemovalThreshold: Ignore any splats with an alpha less than the specified
     *                                     value (valid range: 0 - 255), defaults to 1
     *
     *         position (Array<number>):   Position of the scene, acts as an offset from its default position, defaults to [0, 0, 0]
     *
     *         rotation (Array<number>):   Rotation of the scene represented as a quaternion, defaults to [0, 0, 0, 1]
     *
     *         scale (Array<number>):      Scene's scale, defaults to [1, 1, 1]
     *
     * }
     * @param {Boolean} keepSceneTransforms For a scene that already exists and is being overwritten, this flag
     *                                      says to keep the transform from the existing scene.
     */
    build(splatBuffers, sceneOptions, keepSceneTransforms = true) {
        this.disposeMeshData();
        const totalSplatCount = SplatMesh.getTotalSplatCountForSplatBuffers(splatBuffers);

        const newScenes = SplatMesh.buildScenes(splatBuffers, sceneOptions);
        if (keepSceneTransforms) {
            for (let i = 0; i < this.scenes.length && i < newScenes.length; i++) {
                const newScene = newScenes[i];
                const existingScene = this.getScene(i);
                newScene.copyTransformData(existingScene);
            }
        }
        this.scenes = newScenes;

        this.geometry = SplatMesh.buildGeomtery(totalSplatCount);
        this.material = SplatMesh.buildMaterial(this.dynamicMode);
        const indexMaps = SplatMesh.buildSplatIndexMaps(splatBuffers);
        this.globalSplatIndexToLocalSplatIndexMap = indexMaps.localSplatIndexMap;
        this.globalSplatIndexToSceneIndexMap = indexMaps.sceneIndexMap;
        this.splatTree = SplatMesh.buildSplatTree(this, sceneOptions.map(options => options.splatAlphaRemovalThreshold || 1));

        if (this.enableDistancesComputationOnGPU) this.setupDistancesComputationTransformFeedback();
        this.resetDataFromSplatBuffers();
    }

    /**
     * Dispose all resources held by the splat mesh
     */
    dispose() {
        this.disposeMeshData();
        if (this.enableDistancesComputationOnGPU) {
            this.disposeDistancesComputationGPUResources();
        }
    }

    /**
     * Dispose of only the Three.js mesh resources (geometry, material, and texture)
     */
    disposeMeshData() {
        if (this.geometry && this.geometry !== dummyGeometry) {
            this.geometry.dispose();
            this.geometry = null;
        }
        for (let textureKey in this.splatDataTextures) {
            if (this.splatDataTextures.hasOwnProperty(textureKey)) {
                const textureContainer = this.splatDataTextures[textureKey];
                if (textureContainer.texture) {
                    textureContainer.texture.dispose();
                    textureContainer.texture = null;
                }
            }
        }
        this.splatDataTextures = null;
        if (this.material) {
            this.material.dispose();
            this.material = null;
        }
        this.splatTree = null;
    }

    getSplatTree() {
        return this.splatTree;
    }

    /**
     * Refresh data textures and GPU buffers for splat distance pre-computation with data from the splat buffers for this mesh.
     */
    resetDataFromSplatBuffers() {
        this.uploadSplatDataToTextures();
        if (this.enableDistancesComputationOnGPU) {
            this.updateGPUCentersBufferForDistancesComputation();
            this.updateGPUTransformIndexesBufferForDistancesComputation();
        }
    }

    /**
     * Refresh data textures with data from the splat buffers for this mesh.
     */
    uploadSplatDataToTextures() {

        const splatCount = this.getSplatCount();

        const covariances = new Float32Array(splatCount * 6);
        const centers = new Float32Array(splatCount * 3);
        const colors = new Uint8Array(splatCount * 4);
        this.fillSplatDataArrays(covariances, centers, colors);

        const COVARIANCES_ELEMENTS_PER_TEXEL = 2;
        const CENTER_COLORS_ELEMENTS_PER_TEXEL = 4;
        const TRANSFORM_INDEXES_ELEMENTS_PER_TEXEL = 1;

        const covariancesTextureSize = new THREE.Vector2(4096, 1024);
        while (covariancesTextureSize.x * covariancesTextureSize.y * COVARIANCES_ELEMENTS_PER_TEXEL < splatCount * 6) {
            covariancesTextureSize.y *= 2;
        }

        const centersColorsTextureSize = new THREE.Vector2(4096, 1024);
        while (centersColorsTextureSize.x * centersColorsTextureSize.y * CENTER_COLORS_ELEMENTS_PER_TEXEL < splatCount * 4) {
            centersColorsTextureSize.y *= 2;
        }

        let covariancesTexture;
        let paddedCovariances;
        if (this.halfPrecisionCovariancesOnGPU) {
            paddedCovariances = new Uint16Array(covariancesTextureSize.x * covariancesTextureSize.y * COVARIANCES_ELEMENTS_PER_TEXEL);
            for (let i = 0; i < covariances.length; i++) {
                paddedCovariances[i] = THREE.DataUtils.toHalfFloat(covariances[i]);
            }
            covariancesTexture = new THREE.DataTexture(paddedCovariances, covariancesTextureSize.x,
                                                       covariancesTextureSize.y, THREE.RGFormat, THREE.HalfFloatType);
        } else {
            paddedCovariances = new Float32Array(covariancesTextureSize.x * covariancesTextureSize.y * COVARIANCES_ELEMENTS_PER_TEXEL);
            paddedCovariances.set(covariances);
            covariancesTexture = new THREE.DataTexture(paddedCovariances, covariancesTextureSize.x,
                                                       covariancesTextureSize.y, THREE.RGFormat, THREE.FloatType);
        }
        covariancesTexture.needsUpdate = true;
        this.material.uniforms.covariancesTexture.value = covariancesTexture;
        this.material.uniforms.covariancesTextureSize.value.copy(covariancesTextureSize);

        const paddedCenterColors = new Uint32Array(centersColorsTextureSize.x *
                                                   centersColorsTextureSize.y * CENTER_COLORS_ELEMENTS_PER_TEXEL);
        for (let c = 0; c < splatCount; c++) {
            const colorsBase = c * 4;
            const centersBase = c * 3;
            const centerColorsBase = c * 4;
            paddedCenterColors[centerColorsBase] = rgbaToInteger(colors[colorsBase], colors[colorsBase + 1],
                                                                 colors[colorsBase + 2], colors[colorsBase + 3]);
            paddedCenterColors[centerColorsBase + 1] = uintEncodedFloat(centers[centersBase]);
            paddedCenterColors[centerColorsBase + 2] = uintEncodedFloat(centers[centersBase + 1]);
            paddedCenterColors[centerColorsBase + 3] = uintEncodedFloat(centers[centersBase + 2]);
        }
        const centersColorsTexture = new THREE.DataTexture(paddedCenterColors, centersColorsTextureSize.x,
                                                           centersColorsTextureSize.y, THREE.RGBAIntegerFormat, THREE.UnsignedIntType);
        centersColorsTexture.internalFormat = 'RGBA32UI';
        centersColorsTexture.needsUpdate = true;
        this.material.uniforms.centersColorsTexture.value = centersColorsTexture;
        this.material.uniforms.centersColorsTextureSize.value.copy(centersColorsTextureSize);
        this.material.uniformsNeedUpdate = true;

        this.splatDataTextures = {
            'covariances': {
                'data': paddedCovariances,
                'texture': covariancesTexture,
                'size': covariancesTextureSize
            },
            'centerColors': {
                'data': paddedCenterColors,
                'texture': centersColorsTexture,
                'size': centersColorsTextureSize
            }
        };

        if (this.dynamicMode) {
            const transformIndexesTextureSize = new THREE.Vector2(4096, 1024);
            while (transformIndexesTextureSize.x * transformIndexesTextureSize.y * TRANSFORM_INDEXES_ELEMENTS_PER_TEXEL < splatCount) {
                transformIndexesTextureSize.y *= 2;
            }

            const paddedTransformIndexes = new Uint32Array(transformIndexesTextureSize.x *
                                                           transformIndexesTextureSize.y * TRANSFORM_INDEXES_ELEMENTS_PER_TEXEL);
            for (let c = 0; c < splatCount; c++) {
                paddedTransformIndexes[c] = this.globalSplatIndexToSceneIndexMap[c];
            }
            const transformIndexesTexture = new THREE.DataTexture(paddedTransformIndexes, transformIndexesTextureSize.x,
                                                                  transformIndexesTextureSize.y, THREE.RedIntegerFormat,
                                                                  THREE.UnsignedIntType);
            transformIndexesTexture.internalFormat = 'R32UI';
            transformIndexesTexture.needsUpdate = true;
            this.material.uniforms.transformIndexesTexture.value = transformIndexesTexture;
            this.material.uniforms.transformIndexesTextureSize.value.copy(transformIndexesTextureSize);
            this.material.uniformsNeedUpdate = true;
            this.splatDataTextures['tansformIndexes'] = {
                'data': paddedTransformIndexes,
                'texture': transformIndexesTexture,
                'size': transformIndexesTextureSize
            };
        }
    }

    /**
     * Set the indexes of splats that should be rendered; should be sorted in desired render order.
     * @param {Uint32Array} globalIndexes Sorted index list of splats to be rendered
     * @param {number} renderSplatCount Total number of splats to be rendered. Necessary because we may not want to render
     *                                  every splat.
     */
    updateRenderIndexes(globalIndexes, renderSplatCount) {
        const geometry = this.geometry;
        geometry.attributes.splatIndex.set(globalIndexes);
        geometry.attributes.splatIndex.needsUpdate = true;
        geometry.instanceCount = renderSplatCount;
    }

    /**
     * Update the transforms for each scene in this splat mesh from their individual components (position,
     * quaternion, and scale)
     */
    updateTransforms() {
        for (let i = 0; i < this.scenes.length; i++) {
            const scene = this.getScene(i);
            scene.updateTransform();
        }
    }

    updateUniforms = function() {

        const viewport = new THREE.Vector2();

        return function(renderDimensions, cameraFocalLengthX, cameraFocalLengthY) {
            const splatCount = this.getSplatCount();
            if (splatCount > 0) {
                viewport.set(renderDimensions.x * this.devicePixelRatio,
                             renderDimensions.y * this.devicePixelRatio);
                this.material.uniforms.viewport.value.copy(viewport);
                this.material.uniforms.basisViewport.value.set(1.0 / viewport.x, 1.0 / viewport.y);
                this.material.uniforms.focal.value.set(cameraFocalLengthX, cameraFocalLengthY);
                if (this.dynamicMode) {
                    for (let i = 0; i < this.scenes.length; i++) {
                        this.material.uniforms.transforms.value[i].copy(this.getScene(i).transform);
                    }
                }
                this.material.uniformsNeedUpdate = true;
            }
        };

    }();

    getSplatDataTextures() {
        return this.splatDataTextures;
    }

    getSplatCount() {
        return SplatMesh.getTotalSplatCountForScenes(this.scenes);
    }

    static getTotalSplatCountForScenes(scenes) {
        let totalSplatCount = 0;
        for (let scene of scenes) {
            if (scene && scene.splatBuffer) totalSplatCount += scene.splatBuffer.getSplatCount();
        }
        return totalSplatCount;
    }

    static getTotalSplatCountForSplatBuffers(splatBuffers) {
        let totalSplatCount = 0;
        for (let splatBuffer of splatBuffers) totalSplatCount += splatBuffer.getSplatCount();
        return totalSplatCount;
    }

    disposeDistancesComputationGPUResources() {

        if (!this.renderer) return;

        const gl = this.renderer.getContext();

        if (this.distancesTransformFeedback.vao) {
            gl.deleteVertexArray(this.distancesTransformFeedback.vao);
            this.distancesTransformFeedback.vao = null;
        }
        if (this.distancesTransformFeedback.program) {
            gl.deleteProgram(this.distancesTransformFeedback.program);
            gl.deleteShader(this.distancesTransformFeedback.vertexShader);
            gl.deleteShader(this.distancesTransformFeedback.fragmentShader);
            this.distancesTransformFeedback.program = null;
            this.distancesTransformFeedback.vertexShader = null;
            this.distancesTransformFeedback.fragmentShader = null;
        }
        this.disposeDistancesComputationGPUBufferResources();
        if (this.distancesTransformFeedback.id) {
            gl.deleteTransformFeedback(this.distancesTransformFeedback.id);
            this.distancesTransformFeedback.id = null;
        }
    }

    disposeDistancesComputationGPUBufferResources() {

        if (!this.renderer) return;

        const gl = this.renderer.getContext();

        if (this.distancesTransformFeedback.centersBuffer) {
            this.distancesTransformFeedback.centersBuffer = null;
            gl.deleteBuffer(this.distancesTransformFeedback.centersBuffer);
        }
        if (this.distancesTransformFeedback.outDistancesBuffer) {
            gl.deleteBuffer(this.distancesTransformFeedback.outDistancesBuffer);
            this.distancesTransformFeedback.outDistancesBuffer = null;
        }
    }

    /**
     * Set the Three.js renderer used by this splat mesh
     * @param {THREE.WebGLRenderer} renderer Instance of THREE.WebGLRenderer
     */
    setRenderer(renderer) {
        if (renderer !== this.renderer) {
            this.renderer = renderer;
            if (this.enableDistancesComputationOnGPU && this.getSplatCount() > 0) {
                this.setupDistancesComputationTransformFeedback();
                this.updateGPUCentersBufferForDistancesComputation();
                this.updateGPUTransformIndexesBufferForDistancesComputation();
            }
        }
    }

    setupDistancesComputationTransformFeedback = function() {

        let currentRenderer;
        let currentSplatCount;

        return function() {
            const splatCount = this.getSplatCount();

            if (!this.renderer || (currentRenderer === this.renderer && currentSplatCount === splatCount)) return;
            const rebuildGPUObjects = (currentRenderer !== this.renderer);
            const rebuildBuffers = currentSplatCount !== splatCount;
            if (rebuildGPUObjects) {
                this.disposeDistancesComputationGPUResources();
            } else if (rebuildBuffers) {
                this.disposeDistancesComputationGPUBufferResources();
            }

            const gl = this.renderer.getContext();

            const createShader = (gl, type, source) => {
                const shader = gl.createShader(type);
                if (!shader) {
                    console.error('Fatal error: gl could not create a shader object.');
                    return null;
                }

                gl.shaderSource(shader, source);
                gl.compileShader(shader);

                const compiled = gl.getShaderParameter(shader, gl.COMPILE_STATUS);
                if (!compiled) {
                    let typeName = 'unknown';
                    if (type === gl.VERTEX_SHADER) typeName = 'vertex shader';
                    else if (type === gl.FRAGMENT_SHADER) typeName = 'fragement shader';
                    const errors = gl.getShaderInfoLog(shader);
                    console.error('Failed to compile ' + typeName + ' with these errors:' + errors);
                    gl.deleteShader(shader);
                    return null;
                }

                return shader;
            };

            let vsSource;
            if (this.integerBasedDistancesComputation) {
                vsSource =
                `#version 300 es
                in ivec4 center;
                flat out int distance;`;
                if (this.dynamicMode) {
                    vsSource += `
                        in uint transformIndex;
                        uniform ivec4 transforms[${Constants.MaxScenes}];
                        void main(void) {
                            ivec4 transform = transforms[transformIndex];
                            distance = center.x * transform.x + center.y * transform.y + center.z * transform.z + transform.w * center.w;
                        }
                    `;
                } else {
                    vsSource += `
                        uniform ivec3 modelViewProj;
                        void main(void) {
                            distance = center.x * modelViewProj.x + center.y * modelViewProj.y + center.z * modelViewProj.z;
                        }
                    `;
                }
            } else {
                vsSource =
                `#version 300 es
                in vec3 center;
                flat out float distance;`;
                if (this.dynamicMode) {
                    vsSource += `
                        in uint transformIndex;
                        uniform mat4 transforms[${Constants.MaxScenes}];
                        void main(void) {
                            vec4 transformedCenter = transforms[transformIndex] * vec4(center, 1.0);
                            distance = transformedCenter.z;
                        }
                    `;
                } else {
                    vsSource += `
                        uniform vec3 modelViewProj;
                        void main(void) {
                            distance = center.x * modelViewProj.x + center.y * modelViewProj.y + center.z * modelViewProj.z;
                        }
                    `;
                }
            }

            const fsSource =
            `#version 300 es
                precision lowp float;
                out vec4 fragColor;
                void main(){}
            `;

            const currentVao = gl.getParameter(gl.VERTEX_ARRAY_BINDING);
            const currentProgram = gl.getParameter(gl.CURRENT_PROGRAM);

            if (rebuildGPUObjects) {
                this.distancesTransformFeedback.vao = gl.createVertexArray();
            }

            gl.bindVertexArray(this.distancesTransformFeedback.vao);

            if (rebuildGPUObjects) {
                const program = gl.createProgram();
                const vertexShader = createShader(gl, gl.VERTEX_SHADER, vsSource);
                const fragmentShader = createShader(gl, gl.FRAGMENT_SHADER, fsSource);
                if (!vertexShader || !fragmentShader) {
                    throw new Error('Could not compile shaders for distances computation on GPU.');
                }
                gl.attachShader(program, vertexShader);
                gl.attachShader(program, fragmentShader);
                gl.transformFeedbackVaryings(program, ['distance'], gl.SEPARATE_ATTRIBS);
                gl.linkProgram(program);

                const linked = gl.getProgramParameter(program, gl.LINK_STATUS);
                if (!linked) {
                    const error = gl.getProgramInfoLog(program);
                    console.error('Fatal error: Failed to link program: ' + error);
                    gl.deleteProgram(program);
                    gl.deleteShader(fragmentShader);
                    gl.deleteShader(vertexShader);
                    throw new Error('Could not link shaders for distances computation on GPU.');
                }

                this.distancesTransformFeedback.program = program;
                this.distancesTransformFeedback.vertexShader = vertexShader;
                this.distancesTransformFeedback.vertexShader = fragmentShader;
            }

            gl.useProgram(this.distancesTransformFeedback.program);

            this.distancesTransformFeedback.centersLoc =
                gl.getAttribLocation(this.distancesTransformFeedback.program, 'center');
            if (this.dynamicMode) {
                this.distancesTransformFeedback.transformIndexesLoc =
                    gl.getAttribLocation(this.distancesTransformFeedback.program, 'transformIndex');
                for (let i = 0; i < this.scenes.length; i++) {
                    this.distancesTransformFeedback.transformsLocs[i] =
                        gl.getUniformLocation(this.distancesTransformFeedback.program, `transforms[${i}]`);
                }
            } else {
                this.distancesTransformFeedback.modelViewProjLoc =
                    gl.getUniformLocation(this.distancesTransformFeedback.program, 'modelViewProj');
            }

            if (rebuildGPUObjects || rebuildBuffers) {
                this.distancesTransformFeedback.centersBuffer = gl.createBuffer();
                gl.bindBuffer(gl.ARRAY_BUFFER, this.distancesTransformFeedback.centersBuffer);
                gl.enableVertexAttribArray(this.distancesTransformFeedback.centersLoc);
                if (this.integerBasedDistancesComputation) {
                    gl.vertexAttribIPointer(this.distancesTransformFeedback.centersLoc, 4, gl.INT, 0, 0);
                } else {
                    gl.vertexAttribPointer(this.distancesTransformFeedback.centersLoc, 3, gl.FLOAT, false, 0, 0);
                }

                if (this.dynamicMode) {
                    this.distancesTransformFeedback.transformIndexesBuffer = gl.createBuffer();
                    gl.bindBuffer(gl.ARRAY_BUFFER, this.distancesTransformFeedback.transformIndexesBuffer);
                    gl.enableVertexAttribArray(this.distancesTransformFeedback.transformIndexesLoc);
                    gl.vertexAttribIPointer(this.distancesTransformFeedback.transformIndexesLoc, 1, gl.UNSIGNED_INT, 0, 0);
                }
            }

            if (rebuildGPUObjects || rebuildBuffers) {
                this.distancesTransformFeedback.outDistancesBuffer = gl.createBuffer();
            }
            gl.bindBuffer(gl.ARRAY_BUFFER, this.distancesTransformFeedback.outDistancesBuffer);
            gl.bufferData(gl.ARRAY_BUFFER, splatCount * 4, gl.STATIC_READ);

            if (rebuildGPUObjects) {
                this.distancesTransformFeedback.id = gl.createTransformFeedback();
            }
            gl.bindTransformFeedback(gl.TRANSFORM_FEEDBACK, this.distancesTransformFeedback.id);
            gl.bindBufferBase(gl.TRANSFORM_FEEDBACK_BUFFER, 0, this.distancesTransformFeedback.outDistancesBuffer);

            if (currentProgram) gl.useProgram(currentProgram);
            if (currentVao) gl.bindVertexArray(currentVao);

            currentRenderer = this.renderer;
            currentSplatCount = splatCount;
        };

    }();

    /**
     * Refresh GPU buffers used for computing splat distances with centers data from the scenes for this mesh.
     */
    updateGPUCentersBufferForDistancesComputation() {

        if (!this.renderer) return;

        const gl = this.renderer.getContext();

        const currentVao = gl.getParameter(gl.VERTEX_ARRAY_BINDING);
        gl.bindVertexArray(this.distancesTransformFeedback.vao);

        gl.bindBuffer(gl.ARRAY_BUFFER, this.distancesTransformFeedback.centersBuffer);
        if (this.integerBasedDistancesComputation) {
            const intCenters = this.getIntegerCenters(true);
            gl.bufferData(gl.ARRAY_BUFFER, intCenters, gl.STATIC_DRAW);
        } else {
            const floatCenters = this.getFloatCenters(false);
            gl.bufferData(gl.ARRAY_BUFFER, floatCenters, gl.STATIC_DRAW);
        }
        gl.bindBuffer(gl.ARRAY_BUFFER, null);

        if (currentVao) gl.bindVertexArray(currentVao);
    }

    /**
     * Refresh GPU buffers used for pre-computing splat distances with centers data from the scenes for this mesh.
     */
    updateGPUTransformIndexesBufferForDistancesComputation() {

        if (!this.renderer || !this.dynamicMode) return;

        const gl = this.renderer.getContext();

        const currentVao = gl.getParameter(gl.VERTEX_ARRAY_BINDING);
        gl.bindVertexArray(this.distancesTransformFeedback.vao);

        gl.bindBuffer(gl.ARRAY_BUFFER, this.distancesTransformFeedback.transformIndexesBuffer);
        gl.bufferData(gl.ARRAY_BUFFER, this.getTransformIndexes(), gl.STATIC_DRAW);
        gl.bindBuffer(gl.ARRAY_BUFFER, null);

        if (currentVao) gl.bindVertexArray(currentVao);
    }

    /**
     * Get a typed array containing a mapping from global splat indexes to their scene index.
     * @return {Uint32Array}
     */
    getTransformIndexes() {
        const transformIndexes = new Uint32Array(this.globalSplatIndexToSceneIndexMap.length);
        transformIndexes.set(this.globalSplatIndexToSceneIndexMap);
        return transformIndexes;
    }

    /**
     * Fill 'array' with the transforms for each scene in this splat mesh.
     * @param {Array} array Empty array to be filled with scene transforms. If not empty, contents will be overwritten.
     */
    fillTransformsArray = function() {

        const tempArray = [];

        return function(array) {
            if (tempArray.length !== array.length) tempArray.length = array.length;
            for (let i = 0; i < this.scenes.length; i++) {
                const sceneTransform = this.getScene(i).transform;
                const sceneTransformElements = sceneTransform.elements;
                for (let j = 0; j < 16; j++) {
                    tempArray[i * 16 + j] = sceneTransformElements[j];
                }
            }
            array.set(tempArray);
        };

    }();

    computeDistancesOnGPU = function() {

        const tempMatrix = new THREE.Matrix4();

        return function(modelViewProjMatrix, outComputedDistances) {
            if (!this.renderer) return;

            // console.time("gpu_compute_distances");
            const gl = this.renderer.getContext();

            const currentVao = gl.getParameter(gl.VERTEX_ARRAY_BINDING);
            const currentProgram = gl.getParameter(gl.CURRENT_PROGRAM);

            gl.bindVertexArray(this.distancesTransformFeedback.vao);
            gl.useProgram(this.distancesTransformFeedback.program);

            gl.enable(gl.RASTERIZER_DISCARD);

            if (this.dynamicMode) {
                for (let i = 0; i < this.scenes.length; i++) {
                    tempMatrix.copy(this.getScene(i).transform);
                    tempMatrix.premultiply(modelViewProjMatrix);

                    if (this.integerBasedDistancesComputation) {
                        const iTempMatrix = SplatMesh.getIntegerMatrixArray(tempMatrix);
                        const iTransform = [iTempMatrix[2], iTempMatrix[6], iTempMatrix[10], iTempMatrix[14]];
                        gl.uniform4i(this.distancesTransformFeedback.transformsLocs[i], iTransform[0], iTransform[1],
                                                                                        iTransform[2], iTransform[3]);
                    } else {
                        gl.uniformMatrix4fv(this.distancesTransformFeedback.transformsLocs[i], false, tempMatrix.elements);
                    }
                }
            } else {
                if (this.integerBasedDistancesComputation) {
                    const iViewProjMatrix = SplatMesh.getIntegerMatrixArray(modelViewProjMatrix);
                    const iViewProj = [iViewProjMatrix[2], iViewProjMatrix[6], iViewProjMatrix[10]];
                    gl.uniform3i(this.distancesTransformFeedback.modelViewProjLoc, iViewProj[0], iViewProj[1], iViewProj[2]);
                } else {
                    const viewProj = [modelViewProjMatrix.elements[2], modelViewProjMatrix.elements[6], modelViewProjMatrix.elements[10]];
                    gl.uniform3f(this.distancesTransformFeedback.modelViewProjLoc, viewProj[0], viewProj[1], viewProj[2]);
                }
            }

            gl.bindBuffer(gl.ARRAY_BUFFER, this.distancesTransformFeedback.centersBuffer);
            gl.enableVertexAttribArray(this.distancesTransformFeedback.centersLoc);
            if (this.integerBasedDistancesComputation) {
                gl.vertexAttribIPointer(this.distancesTransformFeedback.centersLoc, 4, gl.INT, 0, 0);
            } else {
                gl.vertexAttribPointer(this.distancesTransformFeedback.centersLoc, 3, gl.FLOAT, false, 0, 0);
            }

            if (this.dynamicMode) {
                gl.bindBuffer(gl.ARRAY_BUFFER, this.distancesTransformFeedback.transformIndexesBuffer);
                gl.enableVertexAttribArray(this.distancesTransformFeedback.transformIndexesLoc);
                gl.vertexAttribIPointer(this.distancesTransformFeedback.transformIndexesLoc, 1, gl.UNSIGNED_INT, 0, 0);
            }

            gl.bindTransformFeedback(gl.TRANSFORM_FEEDBACK, this.distancesTransformFeedback.id);
            gl.bindBufferBase(gl.TRANSFORM_FEEDBACK_BUFFER, 0, this.distancesTransformFeedback.outDistancesBuffer);

            gl.beginTransformFeedback(gl.POINTS);
            gl.drawArrays(gl.POINTS, 0, this.getSplatCount());
            gl.endTransformFeedback();

            gl.bindBufferBase(gl.TRANSFORM_FEEDBACK_BUFFER, 0, null);
            gl.bindTransformFeedback(gl.TRANSFORM_FEEDBACK, null);

            gl.disable(gl.RASTERIZER_DISCARD);

            const sync = gl.fenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0);
            gl.flush();

            const promise = new Promise((resolve) => {
                const checkSync = () => {
                    const timeout = 0;
                    const bitflags = 0;
                    const status = gl.clientWaitSync(sync, bitflags, timeout);
                    switch (status) {
                        case gl.TIMEOUT_EXPIRED:
                            return setTimeout(checkSync);
                        case gl.WAIT_FAILED:
                            throw new Error('should never get here');
                        default:
                            gl.deleteSync(sync);
                            const currentVao = gl.getParameter(gl.VERTEX_ARRAY_BINDING);
                            gl.bindVertexArray(this.distancesTransformFeedback.vao);
                            gl.bindBuffer(gl.ARRAY_BUFFER, this.distancesTransformFeedback.outDistancesBuffer);
                            gl.getBufferSubData(gl.ARRAY_BUFFER, 0, outComputedDistances);
                            gl.bindBuffer(gl.ARRAY_BUFFER, null);

                            if (currentVao) gl.bindVertexArray(currentVao);

                            // console.timeEnd("gpu_compute_distances");

                            resolve();
                    }
                };
                setTimeout(checkSync);
            });

            if (currentProgram) gl.useProgram(currentProgram);
            if (currentVao) gl.bindVertexArray(currentVao);

            return promise;
        };

    }();

    /**
     * Given a global splat index, return corresponding local data (splat buffer, index of splat in that splat
     * buffer, and the corresponding transform)
     * @param {number} globalIndex Global splat index
     * @param {object} paramsObj Object in which to store local data
     * @param {boolean} returnSceneTransform By default, the transform of the scene to which the splat at 'globalIndex' belongs will be
     *                                       returned via the 'sceneTransform' property of 'paramsObj' only if the splat mesh is static.
     *                                       If 'returnSceneTransform' is true, the 'sceneTransform' property will always contain the scene
     *                                       transform, and if 'returnSceneTransform' is false, the 'sceneTransform' property will always
     *                                       be null.
     */
    getLocalSplatParameters(globalIndex, paramsObj, returnSceneTransform) {
        if (returnSceneTransform === undefined || returnSceneTransform === null) {
            returnSceneTransform = this.dynamicMode ? false : true;
        }
        paramsObj.splatBuffer = this.getSplatBufferForSplat(globalIndex);
        paramsObj.localIndex = this.getSplatLocalIndex(globalIndex);
        paramsObj.sceneTransform = returnSceneTransform ? this.getSceneTransformForSplat(globalIndex) : null;
    }

    /**
     * Fill arrays with splat data and apply transforms if appropriate. Each array is optional.
     * @param {Float32Array} covariances Target storage for splat covariances
     * @param {Float32Array} centers Target storage for splat centers
     * @param {Uint8Array} colors Target storage for splat colors
     * @param {boolean} applySceneTransform By default, scene transforms are applied to relevant splat data only if the splat mesh is
     *                                      static. If 'applySceneTransform' is true, scene transforms will always be applied and if
     *                                      it is false, they will never be applied. If undefined, the default behavior will apply.
     */
    fillSplatDataArrays(covariances, centers, colors, applySceneTransform) {
        let offset = 0;
        for (let i = 0; i < this.scenes.length; i++) {
            if (applySceneTransform === undefined || applySceneTransform === null) {
                applySceneTransform = this.dynamicMode ? false : true;
            }
            const scene = this.getScene(i);
            const splatBuffer = scene.splatBuffer;
            const sceneTransform = applySceneTransform ? scene.transform : null;
            if (covariances) splatBuffer.fillSplatCovarianceArray(covariances, offset, sceneTransform);
            if (centers) splatBuffer.fillSplatCenterArray(centers, offset, sceneTransform);
            if (colors) splatBuffer.fillSplatColorArray(colors, offset, sceneTransform);
            offset += splatBuffer.getSplatCount();
        }
    }

    /**
     * Convert splat centers, which are floating point values, to an array of integers and multiply
     * each by 1000. Centers will get transformed as appropriate before conversion to integer.
     * @param {number} padFour Enforce alignement of 4 by inserting a 1000 after every 3 values
     * @return {Int32Array}
     */
    getIntegerCenters(padFour) {
        const splatCount = this.getSplatCount();
        const floatCenters = new Float32Array(splatCount * 3);
        this.fillSplatDataArrays(null, floatCenters, null);
        let intCenters;
        let componentCount = padFour ? 4 : 3;
        intCenters = new Int32Array(splatCount * componentCount);
        for (let i = 0; i < splatCount; i++) {
            for (let t = 0; t < 3; t++) {
                intCenters[i * componentCount + t] = Math.round(floatCenters[i * 3 + t] * 1000.0);
            }
            if (padFour) intCenters[i * componentCount + 3] = 1000;
        }
        return intCenters;
    }


    /**
     * Returns an array of splat centers, transformed as appropriate, optionally padded.
     * @param {number} padFour Enforce alignement of 4 by inserting a 1 after every 3 values
     * @return {Float32Array}
     */
    getFloatCenters(padFour) {
        const splatCount = this.getSplatCount();
        const floatCenters = new Float32Array(splatCount * 3);
        this.fillSplatDataArrays(null, floatCenters, null);
        if (!padFour) return floatCenters;
        let paddedFloatCenters = new Float32Array(splatCount * 4);
        for (let i = 0; i < splatCount; i++) {
            for (let t = 0; t < 3; t++) {
                paddedFloatCenters[i * 4 + t] = floatCenters[i * 3 + t];
            }
            paddedFloatCenters[i * 4 + 3] = 1;
        }
        return paddedFloatCenters;
    }

    /**
     * Get the center for a splat, transformed as appropriate.
     * @param {number} globalIndex Global index of splat
     * @param {THREE.Vector3} outCenter THREE.Vector3 instance in which to store splat center
     * @param {boolean} applySceneTransform By default, if the splat mesh is static, the transform of the scene to which the splat at
     *                                      'globalIndex' belongs will be applied to the splat center. If 'applySceneTransform' is true,
     *                                      the scene transform will always be applied and if 'applySceneTransform' is false, the
     *                                      scene transform will never be applied. If undefined, the default behavior will apply.
     */
    getSplatCenter = function() {

        const paramsObj = {};

        return function(globalIndex, outCenter, applySceneTransform) {
            this.getLocalSplatParameters(globalIndex, paramsObj, applySceneTransform);
            paramsObj.splatBuffer.getSplatCenter(paramsObj.localIndex, outCenter, paramsObj.sceneTransform);
        };

    }();

    /**
     * Get the scale and rotation for a splat, transformed as appropriate.
     * @param {number} globalIndex Global index of splat
     * @param {THREE.Vector3} outScale THREE.Vector3 instance in which to store splat scale
     * @param {THREE.Quaternion} outRotation THREE.Quaternion instance in which to store splat rotation
     * @param {boolean} applySceneTransform By default, if the splat mesh is static, the transform of the scene to which the splat at
     *                                      'globalIndex' belongs will be applied to the splat scale and rotation. If
     *                                      'applySceneTransform' is true, the scene transform will always be applied and if
     *                                      'applySceneTransform' is false, the scene transform will never be applied. If undefined,
     *                                      the default behavior will apply.
     */
    getSplatScaleAndRotation = function() {

        const paramsObj = {};

        return function(globalIndex, outScale, outRotation, applySceneTransform) {
            this.getLocalSplatParameters(globalIndex, paramsObj, applySceneTransform);
            paramsObj.splatBuffer.getSplatScaleAndRotation(paramsObj.localIndex, outScale, outRotation, paramsObj.sceneTransform);
        };

    }();

    /**
     * Get the color for a splat.
     * @param {number} globalIndex Global index of splat
     * @param {THREE.Vector4} outColor THREE.Vector4 instance in which to store splat color
     */
    getSplatColor = function() {

        const paramsObj = {};

        return function(globalIndex, outColor) {
            this.getLocalSplatParameters(globalIndex, paramsObj);
            paramsObj.splatBuffer.getSplatColor(paramsObj.localIndex, outColor, paramsObj.sceneTransform);
        };

    }();

    /**
     * Store the transform of the scene at 'sceneIndex' in 'outTransform'.
     * @param {number} sceneIndex Index of the desired scene
     * @param {THREE.Matrix4} outTransform Instance of THREE.Matrix4 in which to store the scene's transform
     */
    getSceneTransform(sceneIndex, outTransform) {
        const scene = this.getScene(sceneIndex);
        scene.updateTransform();
        outTransform.copy(scene.transform);
    }

    /**
     * Get the scene at 'sceneIndex'.
     * @param {number} sceneIndex Index of the desired scene
     * @return {SplatScene}
     */
    getScene(sceneIndex) {
        if (sceneIndex < 0 || sceneIndex >= this.scenes.length) {
            throw new Error('SplatMesh::getScene() -> Invalid scene index.');
        }
        return this.scenes[sceneIndex];
    }

    getSplatBufferForSplat(globalIndex) {
        return this.getScene(this.globalSplatIndexToSceneIndexMap[globalIndex]).splatBuffer;
    }

    getSceneIndexForSplat(globalIndex) {
        return this.globalSplatIndexToSceneIndexMap[globalIndex];
    }

    getSceneTransformForSplat(globalIndex) {
        return this.getScene(this.globalSplatIndexToSceneIndexMap[globalIndex]).transform;
    }

    getSplatLocalIndex(globalIndex) {
        return this.globalSplatIndexToLocalSplatIndexMap[globalIndex];
    }

    static getIntegerMatrixArray(matrix) {
        const matrixElements = matrix.elements;
        const intMatrixArray = [];
        for (let i = 0; i < 16; i++) {
            intMatrixArray[i] = Math.round(matrixElements[i] * 1000.0);
        }
        return intMatrixArray;
    }
}

var SorterWasm = "AGFzbQEAAAAADAZkeWxpbmsAAAAAAAEbA2AAAGAQf39/f39/f39/f39/f39/fwBgAAF/AhIBA2VudgZtZW1vcnkCAwCAgAQDBAMAAQIHOQMRX193YXNtX2NhbGxfY3RvcnMAAAtzb3J0SW5kZXhlcwABE2Vtc2NyaXB0ZW5fdGxzX2luaXQAAgrHEAMDAAELuxAFAXwDewJ/A30CfiALIAprIQwCQCAOBEAgDQRAQfj///8HIQ5BiICAgHghDSALIAxNDQIgDCEBA0AgAyABQQJ0IgVqIAIgACAFaigCAEECdGooAgAiBTYCACAFIA4gBSAOSBshDiAFIA0gBSANShshDSABQQFqIgEgC0cNAAsMAgsgDwRAQfj///8HIQ5BiICAgHghDSALIAxNDQJBfyEPIAwhAgNAIA8gByAAIAJBAnQiFGooAgAiFUECdGooAgAiCkcEQAJ+IAX9CQIIIAggCkEGdGoiD/0JAgAgDyoCEP0gASAPKgIg/SACIA8qAjD9IAP95gEgBf0JAhggD/0JAgQgDyoCFP0gASAPKgIk/SACIA8qAjT9IAP95gH95AEgBf0JAiggD/0JAgggDyoCGP0gASAPKgIo/SACIA8qAjj9IAP95gH95AEgBf0JAjggD/0JAgwgDyoCHP0gASAPKgIs/SACIA8qAjz9IAP95gH95AEiEf0fArv9FCAR/R8Du/0iAf0MAAAAAABAj0AAAAAAAECPQCIS/fIBIhP9IQEiEJlEAAAAAAAA4ENjBEAgELAMAQtCgICAgICAgICAfwshGQJ+IBP9IQAiEJlEAAAAAAAA4ENjBEAgELAMAQtCgICAgICAgICAfwv9EiETAn4gEf0fALv9FCAR/R8Bu/0iASAS/fIBIhH9IQEiEJlEAAAAAAAA4ENjBEAgELAMAQtCgICAgICAgICAfwshGiATIBn9HgEhEgJ+IBH9IQAiEJlEAAAAAAAA4ENjBEAgELAMAQtCgICAgICAgICAfwv9EiAa/R4BIBL9DQABAgMICQoLEBESExgZGhshEiAKIQ8LIAMgFGogASAVQQR0av0AAAAgEv21ASIR/RsAIBH9GwFqIBH9GwJqIBH9GwNqIgo2AgAgCiAOIAogDkgbIQ4gCiANIAogDUobIQ0gAkEBaiICIAtHDQALDAILAn8gBSoCGLtEAAAAAABAj0CiIhCZRAAAAAAAAOBBYwRAIBCqDAELQYCAgIB4CyEKAn8gBSoCCLtEAAAAAABAj0CiIhCZRAAAAAAAAOBBYwRAIBCqDAELQYCAgIB4CyECAn8gBSoCKLtEAAAAAABAj0CiIhCZRAAAAAAAAOBBYwRAIBCqDAELQYCAgIB4CyEFQfj///8HIQ5BiICAgHghDSALIAxNDQEgAv0RIAr9HAEgBf0cAiESIAwhBQNAIAMgBUECdCICaiABIAAgAmooAgBBBHRq/QAAACAS/bUBIhH9GwAgEf0bAWogEf0bAmoiAjYCACACIA4gAiAOSBshDiACIA0gAiANShshDSAFQQFqIgUgC0cNAAsMAQsgDQRAQfj///8HIQ5BiICAgHghDSALIAxNDQEgDCEBA0AgAyABQQJ0IgVqAn8gAiAAIAVqKAIAQQJ0aioCALtEAAAAAAAAsECiIhCZRAAAAAAAAOBBYwRAIBCqDAELQYCAgIB4CyIKNgIAIAogDiAKIA5IGyEOIAogDSAKIA1KGyENIAFBAWoiASALRw0ACwwBCwJAIA9FBEAgCyAMSw0BQYiAgIB4IQ1B+P///wchDgwCC0H4////ByEOQYiAgIB4IQ0gCyAMTQ0BQX8hDyAMIQIDQCAPIAcgACACQQJ0IhRqKAIAQQJ0IhVqKAIAIgpHBEAgBf0JAgggCCAKQQZ0aiIP/QkCACAPKgIQ/SABIA8qAiD9IAIgDyoCMP0gA/3mASAF/QkCGCAP/QkCBCAPKgIU/SABIA8qAiT9IAIgDyoCNP0gA/3mAf3kASAF/QkCKCAP/QkCCCAPKgIY/SABIA8qAij9IAIgDyoCOP0gA/3mAf3kASAF/QkCOCAP/QkCDCAPKgIc/SABIA8qAiz9IAIgDyoCPP0gA/3mAf3kASERIAohDwsgAyAUagJ/IBEgASAVQQJ0IgpqKQIA/RL95gEiEv0fACAS/R8BkiARIBH9DQgJCgsMDQ4PAAAAAAAAAAAgASAKQQhyaikCAP0S/eYBIhL9HwCSIBL9HwGSu0QAAAAAAACwQKIiEJlEAAAAAAAA4EFjBEAgEKoMAQtBgICAgHgLIgo2AgAgCiAOIAogDkgbIQ4gCiANIAogDUobIQ0gAkEBaiICIAtHDQALDAELIAUqAighFiAFKgIYIRcgBSoCCCEYQfj///8HIQ5BiICAgHghDSAMIQUDQAJ/IBggASAAIAVBAnQiB2ooAgBBBHRqIgIqAgCUIBcgAioCBJSSIBYgAioCCJSSu0QAAAAAAACwQKIiEJlEAAAAAAAA4EFjBEAgEKoMAQtBgICAgHgLIQogAyAHaiAKNgIAIAogDiAKIA5IGyEOIAogDSAKIA1KGyENIAVBAWoiBSALRw0ACwsgCyAMSwRAIAlBAWuzIA2yIA6yk5UhFiAMIQ0DQAJ/IBYgAyANQQJ0aiIBKAIAIA5rspQiF4tDAAAAT10EQCAXqAwBC0GAgICAeAshCiABIAo2AgAgBCAKQQJ0aiIBIAEoAgBBAWo2AgAgDUEBaiINIAtHDQALCyAJQQJPBEAgBCgCACENQQEhDgNAIAQgDkECdGoiASABKAIAIA1qIg02AgAgDkEBaiIOIAlHDQALCyAMQQBKBEAgDCEOA0AgBiAOQQFrIgFBAnQiAmogACACaigCADYCACAOQQFKIQIgASEOIAINAAsLIAsgDEoEQCALIQ4DQCAGIAsgBCADIA5BAWsiDkECdCIBaigCAEECdGoiAigCACIFa0ECdGogACABaigCADYCACACIAVBAWs2AgAgDCAOSA0ACwsLBABBAAs=";

function sortWorker(self) {

    let wasmInstance;
    let wasmMemory;
    let useSharedMemory;
    let integerBasedSort;
    let dynamicMode;
    let splatCount;
    let indexesToSortOffset;
    let sortedIndexesOffset;
    let transformIndexesOffset;
    let transformsOffset;
    let precomputedDistancesOffset;
    let mappedDistancesOffset;
    let frequenciesOffset;
    let centersOffset;
    let modelViewProjOffset;
    let countsZero;

    let Constants;

    function sort(splatSortCount, splatRenderCount, modelViewProj,
                  usePrecomputedDistances, copyIndexesToSort, copyPrecomputedDistances, copyTransforms) {
        const sortStartTime = performance.now();

        if (!useSharedMemory) {
            const indexesToSort = new Uint32Array(wasmMemory, indexesToSortOffset, copyIndexesToSort.byteLength / Constants.BytesPerInt);
            indexesToSort.set(copyIndexesToSort);
            const transforms = new Float32Array(wasmMemory, transformsOffset, copyTransforms.byteLength / Constants.BytesPerFloat);
            transforms.set(copyTransforms);
            if (usePrecomputedDistances) {
                let precomputedDistances;
                if (integerBasedSort) {
                    precomputedDistances = new Int32Array(wasmMemory, precomputedDistancesOffset,
                                                          copyPrecomputedDistances.byteLength / Constants.BytesPerInt);
                } else {
                    precomputedDistances = new Float32Array(wasmMemory, precomputedDistancesOffset,
                                                            copyPrecomputedDistances.byteLength / Constants.BytesPerFloat);
                }
                precomputedDistances.set(copyPrecomputedDistances);
            }
        }

        if (!countsZero) countsZero = new Uint32Array(Constants.DepthMapRange);
        new Float32Array(wasmMemory, modelViewProjOffset, 16).set(modelViewProj);
        new Uint32Array(wasmMemory, frequenciesOffset, Constants.DepthMapRange).set(countsZero);
        wasmInstance.exports.sortIndexes(indexesToSortOffset, centersOffset, precomputedDistancesOffset,
                                         mappedDistancesOffset, frequenciesOffset, modelViewProjOffset,
                                         sortedIndexesOffset, transformIndexesOffset, transformsOffset, Constants.DepthMapRange,
                                         splatSortCount, splatRenderCount, splatCount, usePrecomputedDistances, integerBasedSort,
                                         dynamicMode);

        const sortMessage = {
            'sortDone': true,
            'splatSortCount': splatSortCount,
            'splatRenderCount': splatRenderCount,
            'sortTime': 0
        };
        const transferBuffers = [];
        if (!useSharedMemory) {
            const sortedIndexes = new Uint32Array(wasmMemory, sortedIndexesOffset, splatRenderCount);
            const sortedIndexesOut = new Uint32Array(splatRenderCount);
            sortedIndexesOut.set(sortedIndexes);
            sortMessage.sortedIndexes = sortedIndexesOut.buffer;
            transferBuffers.push(sortedIndexesOut.buffer);
        }
        const sortEndTime = performance.now();

        sortMessage.sortTime = sortEndTime - sortStartTime;

        self.postMessage(sortMessage, transferBuffers);
    }

    self.onmessage = (e) => {
        if (e.data.centers) {
            centers = e.data.centers;
            transformIndexes = e.data.transformIndexes;
            if (integerBasedSort) {
                new Int32Array(wasmMemory, centersOffset, splatCount * 4).set(new Int32Array(centers));
            } else {
                new Float32Array(wasmMemory, centersOffset, splatCount * 4).set(new Float32Array(centers));
            }
            if (dynamicMode) {
                new Uint32Array(wasmMemory, transformIndexesOffset, splatCount).set(new Uint32Array(transformIndexes));
            }
            self.postMessage({
                'sortSetupComplete': true,
            });
        } else if (e.data.sort) {
            const renderCount = e.data.sort.splatRenderCount || 0;
            const sortCount = e.data.sort.splatSortCount || 0;
            const usePrecomputedDistances = e.data.sort.usePrecomputedDistances;

            let copyIndexesToSort;
            let copyPrecomputedDistances;
            let copyTransforms;
            if (!useSharedMemory) {
                copyIndexesToSort = e.data.sort.indexesToSort;
                copyTransforms = e.data.sort.transforms;
                if (usePrecomputedDistances) copyPrecomputedDistances = e.data.sort.precomputedDistances;
            }
            sort(sortCount, renderCount, e.data.sort.modelViewProj, usePrecomputedDistances,
                 copyIndexesToSort, copyPrecomputedDistances, copyTransforms);
        } else if (e.data.init) {
            // Yep, this is super hacky and gross :(
            Constants = e.data.init.Constants;

            splatCount = e.data.init.splatCount;
            useSharedMemory = e.data.init.useSharedMemory;
            integerBasedSort = e.data.init.integerBasedSort;
            dynamicMode = e.data.init.dynamicMode;

            const CENTERS_BYTES_PER_ENTRY = integerBasedSort ? (Constants.BytesPerInt * 4) : (Constants.BytesPerFloat * 4);

            const sorterWasmBytes = new Uint8Array(e.data.init.sorterWasmBytes);

            const matrixSize = 16 * Constants.BytesPerFloat;
            const memoryRequiredForIndexesToSort = splatCount * Constants.BytesPerInt;
            const memoryRequiredForCenters = splatCount * CENTERS_BYTES_PER_ENTRY;
            const memoryRequiredForModelViewProjectionMatrix = matrixSize;
            const memoryRequiredForPrecomputedDistances = integerBasedSort ?
                                                          (splatCount * Constants.BytesPerInt) : (splatCount * Constants.BytesPerFloat);
            const memoryRequiredForMappedDistances = splatCount * Constants.BytesPerInt;
            const memoryRequiredForSortedIndexes = splatCount * Constants.BytesPerInt;
            const memoryRequiredForIntermediateSortBuffers = Constants.DepthMapRange * Constants.BytesPerInt * 2;
            const memoryRequiredforTransformIndexes = dynamicMode ? (splatCount * Constants.BytesPerInt) : 0;
            const memoryRequiredforTransforms = dynamicMode ? (Constants.MaxScenes * matrixSize) : 0;
            const extraMemory = Constants.MemoryPageSize * 32;

            const totalRequiredMemory = memoryRequiredForIndexesToSort +
                                        memoryRequiredForCenters +
                                        memoryRequiredForModelViewProjectionMatrix +
                                        memoryRequiredForPrecomputedDistances +
                                        memoryRequiredForMappedDistances +
                                        memoryRequiredForIntermediateSortBuffers +
                                        memoryRequiredForSortedIndexes +
                                        memoryRequiredforTransformIndexes +
                                        memoryRequiredforTransforms +
                                        extraMemory;
            const totalPagesRequired = Math.floor(totalRequiredMemory / Constants.MemoryPageSize ) + 1;
            const sorterWasmImport = {
                module: {},
                env: {
                    memory: new WebAssembly.Memory({
                        initial: totalPagesRequired * 2,
                        maximum: totalPagesRequired * 4,
                        shared: true,
                    }),
                }
            };
            WebAssembly.compile(sorterWasmBytes)
            .then((wasmModule) => {
                return WebAssembly.instantiate(wasmModule, sorterWasmImport);
            })
            .then((instance) => {
                wasmInstance = instance;
                indexesToSortOffset = 0;
                centersOffset = indexesToSortOffset + memoryRequiredForIndexesToSort;
                modelViewProjOffset = centersOffset + memoryRequiredForCenters;
                precomputedDistancesOffset = modelViewProjOffset + memoryRequiredForModelViewProjectionMatrix;
                mappedDistancesOffset = precomputedDistancesOffset + memoryRequiredForPrecomputedDistances;
                frequenciesOffset = mappedDistancesOffset + memoryRequiredForMappedDistances;
                sortedIndexesOffset = frequenciesOffset + memoryRequiredForIntermediateSortBuffers;
                transformIndexesOffset = sortedIndexesOffset + memoryRequiredForSortedIndexes;
                transformsOffset = transformIndexesOffset + memoryRequiredforTransformIndexes;
                wasmMemory = sorterWasmImport.env.memory.buffer;
                if (useSharedMemory) {
                    self.postMessage({
                        'sortSetupPhase1Complete': true,
                        'indexesToSortBuffer': wasmMemory,
                        'indexesToSortOffset': indexesToSortOffset,
                        'sortedIndexesBuffer': wasmMemory,
                        'sortedIndexesOffset': sortedIndexesOffset,
                        'precomputedDistancesBuffer': wasmMemory,
                        'precomputedDistancesOffset': precomputedDistancesOffset,
                        'transformsBuffer': wasmMemory,
                        'transformsOffset': transformsOffset
                    });
                } else {
                    self.postMessage({
                        'sortSetupPhase1Complete': true
                    });
                }
            });
        }
    };
}

function createSortWorker(splatCount, useSharedMemory, integerBasedSort, dynamicMode) {
    const worker = new Worker(
        URL.createObjectURL(
            new Blob(['(', sortWorker.toString(), ')(self)'], {
                type: 'application/javascript',
            }),
        ),
    );

    const sorterWasmBinaryString = atob(SorterWasm);
    const sorterWasmBytes = new Uint8Array(sorterWasmBinaryString.length);
    for (let i = 0; i < sorterWasmBinaryString.length; i++) {
        sorterWasmBytes[i] = sorterWasmBinaryString.charCodeAt(i);
    }

    worker.postMessage({
        'init': {
            'sorterWasmBytes': sorterWasmBytes.buffer,
            'splatCount': splatCount,
            'useSharedMemory': useSharedMemory,
            'integerBasedSort': integerBasedSort,
            'dynamicMode': dynamicMode,
            // Super hacky
            'Constants': {
                'BytesPerFloat': Constants.BytesPerFloat,
                'BytesPerInt': Constants.BytesPerInt,
                'DepthMapRange': Constants.DepthMapRange,
                'MemoryPageSize': Constants.MemoryPageSize,
                'MaxScenes': Constants.MaxScenes
            }
        }
    });
    return worker;
}

const WebXRMode = {
    None: 0,
    VR: 1,
    AR: 2
};

/*
Copyright © 2010-2024 three.js authors & Mark Kellogg

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
*/

class VRButton {

    static createButton( renderer ) {

        const button = document.createElement( 'button' );

        function showEnterVR( /* device */ ) {

            let currentSession = null;

            async function onSessionStarted( session ) {

                session.addEventListener( 'end', onSessionEnded );

                await renderer.xr.setSession( session );
                button.textContent = 'EXIT VR';

                currentSession = session;

            }

            function onSessionEnded( /* event */ ) {

                currentSession.removeEventListener( 'end', onSessionEnded );

                button.textContent = 'ENTER VR';

                currentSession = null;

            }

            //

            button.style.display = '';

            button.style.cursor = 'pointer';
            button.style.left = 'calc(50% - 50px)';
            button.style.width = '100px';

            button.textContent = 'ENTER VR';

            // WebXR's requestReferenceSpace only works if the corresponding feature
            // was requested at session creation time. For simplicity, just ask for
            // the interesting ones as optional features, but be aware that the
            // requestReferenceSpace call will fail if it turns out to be unavailable.
            // ('local' is always available for immersive sessions and doesn't need to
            // be requested separately.)

            const sessionInit = { optionalFeatures: ['local-floor', 'bounded-floor', 'hand-tracking', 'layers'] };

            button.onmouseenter = function() {

                button.style.opacity = '1.0';

            };

            button.onmouseleave = function() {

                button.style.opacity = '0.5';

            };

            button.onclick = function() {

                if ( currentSession === null ) {

                    navigator.xr.requestSession( 'immersive-vr', sessionInit ).then( onSessionStarted );

                } else {

                    currentSession.end();

                    if ( navigator.xr.offerSession !== undefined ) {

                        navigator.xr.offerSession( 'immersive-vr', sessionInit )
                            .then( onSessionStarted )
                            .catch( ( err ) => {

                                console.warn( err );

                            } );

                    }

                }

            };

            if ( navigator.xr.offerSession !== undefined ) {

                navigator.xr.offerSession( 'immersive-vr', sessionInit )
                    .then( onSessionStarted )
                    .catch( ( err ) => {

                        console.warn( err );

                    } );

            }

        }

        function disableButton() {

            button.style.display = '';

            button.style.cursor = 'auto';
            button.style.left = 'calc(50% - 75px)';
            button.style.width = '150px';

            button.onmouseenter = null;
            button.onmouseleave = null;

            button.onclick = null;

        }

        function showWebXRNotFound() {

            disableButton();

            button.textContent = 'VR NOT SUPPORTED';

        }

        function showVRNotAllowed( exception ) {

            disableButton();

            console.warn( 'Exception when trying to call xr.isSessionSupported', exception );

            button.textContent = 'VR NOT ALLOWED';

        }

        function stylizeElement( element ) {

            element.style.position = 'absolute';
            element.style.bottom = '20px';
            element.style.padding = '12px 6px';
            element.style.border = '1px solid #fff';
            element.style.borderRadius = '4px';
            element.style.background = 'rgba(0,0,0,0.1)';
            element.style.color = '#fff';
            element.style.font = 'normal 13px sans-serif';
            element.style.textAlign = 'center';
            element.style.opacity = '0.5';
            element.style.outline = 'none';
            element.style.zIndex = '999';

        }

        if ( 'xr' in navigator ) {

            button.id = 'VRButton';
            button.style.display = 'none';

            stylizeElement( button );

            navigator.xr.isSessionSupported( 'immersive-vr' ).then( function( supported ) {

                supported ? showEnterVR() : showWebXRNotFound();

                if ( supported && VRButton.xrSessionIsGranted ) {

                    button.click();

                }

            } ).catch( showVRNotAllowed );

            return button;

        } else {

            const message = document.createElement( 'a' );

            if ( window.isSecureContext === false ) {

                message.href = document.location.href.replace( /^http:/, 'https:' );
                message.innerHTML = 'WEBXR NEEDS HTTPS'; // TODO Improve message

            } else {

                message.href = 'https://immersiveweb.dev/';
                message.innerHTML = 'WEBXR NOT AVAILABLE';

            }

            message.style.left = 'calc(50% - 90px)';
            message.style.width = '180px';
            message.style.textDecoration = 'none';

            stylizeElement( message );

            return message;

        }

    }

    static registerSessionGrantedListener() {

        if ( typeof navigator !== 'undefined' && 'xr' in navigator ) {

            // WebXRViewer (based on Firefox) has a bug where addEventListener
            // throws a silent exception and aborts execution entirely.
            if ( /WebXRViewer\//i.test( navigator.userAgent ) ) return;

            navigator.xr.addEventListener( 'sessiongranted', () => {

                VRButton.xrSessionIsGranted = true;

            } );

        }

    }

}

VRButton.xrSessionIsGranted = false;
VRButton.registerSessionGrantedListener();

/*
Copyright © 2010-2024 three.js authors & Mark Kellogg

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
*/

class ARButton {

    static createButton( renderer, sessionInit = {} ) {

        const button = document.createElement( 'button' );

        function showStartAR( /* device */ ) {

            if ( sessionInit.domOverlay === undefined ) {

                const overlay = document.createElement( 'div' );
                overlay.style.display = 'none';
                document.body.appendChild( overlay );

                const svg = document.createElementNS( 'http://www.w3.org/2000/svg', 'svg' );
                svg.setAttribute( 'width', 38 );
                svg.setAttribute( 'height', 38 );
                svg.style.position = 'absolute';
                svg.style.right = '20px';
                svg.style.top = '20px';
                svg.addEventListener( 'click', function() {

                    currentSession.end();

                } );
                overlay.appendChild( svg );

                const path = document.createElementNS( 'http://www.w3.org/2000/svg', 'path' );
                path.setAttribute( 'd', 'M 12,12 L 28,28 M 28,12 12,28' );
                path.setAttribute( 'stroke', '#fff' );
                path.setAttribute( 'stroke-width', 2 );
                svg.appendChild( path );

                if ( sessionInit.optionalFeatures === undefined ) {

                    sessionInit.optionalFeatures = [];

                }

                sessionInit.optionalFeatures.push( 'dom-overlay' );
                sessionInit.domOverlay = { root: overlay };

            }

            //

            let currentSession = null;

            async function onSessionStarted( session ) {

                session.addEventListener( 'end', onSessionEnded );

                renderer.xr.setReferenceSpaceType( 'local' );

                await renderer.xr.setSession( session );

                button.textContent = 'STOP AR';
                sessionInit.domOverlay.root.style.display = '';

                currentSession = session;

            }

            function onSessionEnded( /* event */ ) {

                currentSession.removeEventListener( 'end', onSessionEnded );

                button.textContent = 'START AR';
                sessionInit.domOverlay.root.style.display = 'none';

                currentSession = null;

            }

            //

            button.style.display = '';

            button.style.cursor = 'pointer';
            button.style.left = 'calc(50% - 50px)';
            button.style.width = '100px';

            button.textContent = 'START AR';

            button.onmouseenter = function() {

                button.style.opacity = '1.0';

            };

            button.onmouseleave = function() {

                button.style.opacity = '0.5';

            };

            button.onclick = function() {

                if ( currentSession === null ) {

                    navigator.xr.requestSession( 'immersive-ar', sessionInit ).then( onSessionStarted );

                } else {

                    currentSession.end();

                    if ( navigator.xr.offerSession !== undefined ) {

                        navigator.xr.offerSession( 'immersive-ar', sessionInit )
                            .then( onSessionStarted )
                            .catch( ( err ) => {

                                console.warn( err );

                            } );

                    }

                }

            };

            if ( navigator.xr.offerSession !== undefined ) {

                navigator.xr.offerSession( 'immersive-ar', sessionInit )
                    .then( onSessionStarted )
                    .catch( ( err ) => {

                        console.warn( err );

                    } );

            }

        }

        function disableButton() {

            button.style.display = '';

            button.style.cursor = 'auto';
            button.style.left = 'calc(50% - 75px)';
            button.style.width = '150px';

            button.onmouseenter = null;
            button.onmouseleave = null;

            button.onclick = null;

        }

        function showARNotSupported() {

            disableButton();

            button.textContent = 'AR NOT SUPPORTED';

        }

        function showARNotAllowed( exception ) {

            disableButton();

            console.warn( 'Exception when trying to call xr.isSessionSupported', exception );

            button.textContent = 'AR NOT ALLOWED';

        }

        function stylizeElement( element ) {

            element.style.position = 'absolute';
            element.style.bottom = '20px';
            element.style.padding = '12px 6px';
            element.style.border = '1px solid #fff';
            element.style.borderRadius = '4px';
            element.style.background = 'rgba(0,0,0,0.1)';
            element.style.color = '#fff';
            element.style.font = 'normal 13px sans-serif';
            element.style.textAlign = 'center';
            element.style.opacity = '0.5';
            element.style.outline = 'none';
            element.style.zIndex = '999';

        }

        if ( 'xr' in navigator ) {

            button.id = 'ARButton';
            button.style.display = 'none';

            stylizeElement( button );

            navigator.xr.isSessionSupported( 'immersive-ar' ).then( function( supported ) {

                supported ? showStartAR() : showARNotSupported();

            } ).catch( showARNotAllowed );

            return button;

        } else {

            const message = document.createElement( 'a' );

            if ( window.isSecureContext === false ) {

                message.href = document.location.href.replace( /^http:/, 'https:' );
                message.innerHTML = 'WEBXR NEEDS HTTPS'; // TODO Improve message

            } else {

                message.href = 'https://immersiveweb.dev/';
                message.innerHTML = 'WEBXR NOT AVAILABLE';

            }

            message.style.left = 'calc(50% - 90px)';
            message.style.width = '180px';
            message.style.textDecoration = 'none';

            stylizeElement( message );

            return message;

        }

    }

}

const THREE_CAMERA_FOV = 50;
const MINIMUM_DISTANCE_TO_NEW_FOCAL_POINT = .75;

/**
 * Viewer: Manages the rendering of splat scenes. Manages an instance of SplatMesh as well as a web worker
 * that performs the sort for its splats.
 */
class Viewer {

    constructor(options = {}) {

        // The natural 'up' vector for viewing the scene (only has an effect when used with orbit controls and
        // when the viewer uses its own camera).
        if (!options.cameraUp) options.cameraUp = [0, 1, 0];
        this.cameraUp = new THREE.Vector3().fromArray(options.cameraUp);

        // The camera's initial position (only used when the viewer uses its own camera).
        if (!options.initialCameraPosition) options.initialCameraPosition = [0, 10, 15];
        this.initialCameraPosition = new THREE.Vector3().fromArray(options.initialCameraPosition);

        // The initial focal point of the camera and center of the camera's orbit (only used when the viewer uses its own camera).
        if (!options.initialCameraLookAt) options.initialCameraLookAt = [0, 0, 0];
        this.initialCameraLookAt = new THREE.Vector3().fromArray(options.initialCameraLookAt);

        // 'dropInMode' is a flag that is used internally to support the usage of the viewer as a Three.js scene object
        this.dropInMode = options.dropInMode || false;

        // If 'selfDrivenMode' is true, the viewer manages its own update/animation loop via requestAnimationFrame()
        if (options.selfDrivenMode === undefined || options.selfDrivenMode === null) options.selfDrivenMode = true;
        this.selfDrivenMode = options.selfDrivenMode && !this.dropInMode;
        this.selfDrivenUpdateFunc = this.selfDrivenUpdate.bind(this);

        // If 'useBuiltInControls' is true, the viewer will create its own instance of OrbitControls and attach to the camera
        if (options.useBuiltInControls === undefined) options.useBuiltInControls = true;
        this.useBuiltInControls = options.useBuiltInControls;

        // parent element of the Three.js renderer canvas
        this.rootElement = options.rootElement;

        // Tells the viewer to pretend the device pixel ratio is 1, which can boost performance on devices where it is larger,
        // at a small cost to visual quality
        this.ignoreDevicePixelRatio = options.ignoreDevicePixelRatio || false;
        this.devicePixelRatio = this.ignoreDevicePixelRatio ? 1 : window.devicePixelRatio;

        // Tells the viewer to use 16-bit floating point values when storing splat covariance data in textures, instead of 32-bit
        if (options.halfPrecisionCovariancesOnGPU === undefined || options.halfPrecisionCovariancesOnGPU === null) {
            options.halfPrecisionCovariancesOnGPU = true;
        }
        this.halfPrecisionCovariancesOnGPU = options.halfPrecisionCovariancesOnGPU;

        // If 'threeScene' is valid, it will be rendered by the viewer along with the splat mesh
        this.threeScene = options.threeScene;
        // Allows for usage of an external Three.js renderer
        this.renderer = options.renderer;
        // Allows for usage of an external Three.js camera
        this.camera = options.camera;

        // If 'gpuAcceleratedSort' is true, a partially GPU-accelerated approach to sorting splats will be used.
        // Currently this means pre-computing splat distances from the camera on the GPU
        this.gpuAcceleratedSort = options.gpuAcceleratedSort;
        if (this.gpuAcceleratedSort !== true && this.gpuAcceleratedSort !== false) {
            if (this.isMobile()) this.gpuAcceleratedSort = false;
            else this.gpuAcceleratedSort = true;
        }

        // if 'integerBasedSort' is true, the integer version of splat centers as well as other values used to calculate
        // splat distances are used instead of the float version. This speeds up computation, but introduces the possibility of
        // overflow in larger scenes.
        if (options.integerBasedSort === undefined || options.integerBasedSort === null) {
            options.integerBasedSort = true;
        }
        this.integerBasedSort = options.integerBasedSort;

        // If 'sharedMemoryForWorkers' is true, a SharedArrayBuffer will be used to communicate with web workers. This method
        // is faster than copying memory to or from web workers, but comes with security implications as outlined here:
        // https://web.dev/articles/cross-origin-isolation-guide
        // If enabled, it requires specific CORS headers to be present in the response from the server that is sent when
        // loading the application. More information is available in the README.
        if (options.sharedMemoryForWorkers === undefined || options.sharedMemoryForWorkers === null) options.sharedMemoryForWorkers = true;
        this.sharedMemoryForWorkers = options.sharedMemoryForWorkers;

        // if 'dynamicScene' is true, it tells the viewer to assume scene elements are not stationary or that the number of splats in the
        // scene may change. This prevents optimizations that depend on a static scene from being made. Additionally, if 'dynamicScene' is
        // true it tells the splat mesh to not apply scene tranforms to splat data that is returned by functions like
        // SplatMesh.getSplatCenter() by default.
        const dynamicScene = !!options.dynamicScene;
        this.splatMesh = new SplatMesh(dynamicScene, this.halfPrecisionCovariancesOnGPU, this.devicePixelRatio,
                                       this.gpuAcceleratedSort, this.integerBasedSort);


        this.webXRMode = options.webXRMode || WebXRMode.None;

        this.controls = null;

        this.showMeshCursor = false;
        this.showControlPlane = false;
        this.showInfo = false;

        this.sceneHelper = null;

        this.sortWorker = null;
        this.sortRunning = false;
        this.splatRenderCount = 0;
        this.sortWorkerIndexesToSort = null;
        this.sortWorkerSortedIndexes = null;
        this.sortWorkerPrecomputedDistances = null;
        this.sortWorkerTransforms = null;

        this.selfDrivenModeRunning = false;
        this.splatRenderingInitialized = false;

        this.raycaster = new Raycaster();

        this.infoPanel = null;
        this.infoPanelCells = {};

        this.currentFPS = 0;
        this.lastSortTime = 0;

        this.previousCameraTarget = new THREE.Vector3();
        this.nextCameraTarget = new THREE.Vector3();

        this.mousePosition = new THREE.Vector2();
        this.mouseDownPosition = new THREE.Vector2();
        this.mouseDownTime = null;

        this.resizeObserver = null;
        this.mouseMoveListener = null;
        this.mouseDownListener = null;
        this.mouseUpListener = null;
        this.keyDownListener = null;

        this.sortPromise = null;
        this.sortPromiseResolver = null;

        this.loadingSpinner = new LoadingSpinner(null, this.rootElement || document.body);
        this.loadingSpinner.hide();

        this.usingExternalCamera = (this.dropInMode || this.camera) ? true : false;
        this.usingExternalRenderer = (this.dropInMode || this.renderer) ? true : false;

        this.initialized = false;
        if (!this.dropInMode) this.init();
    }

    init() {

        if (this.initialized) return;

        if (!this.rootElement) {
            if (!this.usingExternalRenderer) {
                this.rootElement = document.createElement('div');
                this.rootElement.style.width = '100%';
                this.rootElement.style.height = '100%';
                this.rootElement.style.position = 'absolute';
                document.body.appendChild(this.rootElement);
            } else {
                this.rootElement = this.renderer.domElement.parentElement || document.body;
            }
        }

        const renderDimensions = new THREE.Vector2();
        this.getRenderDimensions(renderDimensions);

        if (!this.usingExternalCamera) {
            this.camera = new THREE.PerspectiveCamera(THREE_CAMERA_FOV, renderDimensions.x / renderDimensions.y, 0.1, 500);
            this.camera.position.copy(this.initialCameraPosition);
            this.camera.up.copy(this.cameraUp).normalize();
            this.camera.lookAt(this.initialCameraLookAt);
        }

        if (!this.usingExternalRenderer) {
            this.renderer = new THREE.WebGLRenderer({
                antialias: false,
                precision: 'highp'
            });
            this.renderer.setPixelRatio(this.devicePixelRatio);
            this.renderer.autoClear = true;
            this.renderer.setClearColor(new THREE.Color( 0x000000 ), 0.0);
            this.renderer.setSize(renderDimensions.x, renderDimensions.y);

            this.resizeObserver = new ResizeObserver(() => {
                this.getRenderDimensions(renderDimensions);
                this.renderer.setSize(renderDimensions.x, renderDimensions.y);
            });
            this.resizeObserver.observe(this.rootElement);
            this.rootElement.appendChild(this.renderer.domElement);
        }

        if (this.webXRMode) {
            if (this.webXRMode === WebXRMode.VR) {
                this.rootElement.appendChild(VRButton.createButton(this.renderer));
            } else if (this.webXRMode === WebXRMode.AR) {
                this.rootElement.appendChild(ARButton.createButton(this.renderer));
            }
            this.renderer.xr.enabled = true;
            this.camera.up.copy(this.cameraUp).normalize();
            this.camera.lookAt(this.initialCameraLookAt);
        }

        this.threeScene = this.threeScene || new THREE.Scene();
        this.sceneHelper = new SceneHelper(this.threeScene);
        this.sceneHelper.setupMeshCursor();
        this.sceneHelper.setupFocusMarker();
        this.sceneHelper.setupControlPlane();

        if (this.useBuiltInControls && this.webXRMode === WebXRMode.None) {
            this.controls = new OrbitControls(this.camera, this.renderer.domElement);
            this.controls.listenToKeyEvents(window);
            this.controls.rotateSpeed = 0.5;
            this.controls.maxPolarAngle = Math.PI * .75;
            this.controls.minPolarAngle = 0.1;
            this.controls.enableDamping = true;
            this.controls.dampingFactor = 0.05;
            this.controls.target.copy(this.initialCameraLookAt);
            this.mouseMoveListener = this.onMouseMove.bind(this);
            this.rootElement.addEventListener('pointermove', this.mouseMoveListener, false);
            this.mouseDownListener = this.onMouseDown.bind(this);
            this.rootElement.addEventListener('pointerdown', this.mouseDownListener, false);
            this.mouseUpListener = this.onMouseUp.bind(this);
            this.rootElement.addEventListener('pointerup', this.mouseUpListener, false);
            this.keyDownListener = this.onKeyDown.bind(this);
            window.addEventListener('keydown', this.keyDownListener, false);
        }

        this.setupInfoPanel();
        this.loadingSpinner.setContainer(this.rootElement);

        this.initialized = true;
    }

    removeEventHandlers() {
        if (this.useBuiltInControls) {
            this.rootElement.removeEventListener('pointermove', this.mouseMoveListener);
            this.mouseMoveListener = null;
            this.rootElement.removeEventListener('pointerdown', this.mouseDownListener);
            this.mouseDownListener = null;
            this.rootElement.removeEventListener('pointerup', this.mouseUpListener);
            this.mouseUpListener = null;
            window.removeEventListener('keydown', this.keyDownListener);
            this.keyDownListener = null;
        }
    }

    onKeyDown = function() {

        const forward = new THREE.Vector3();
        const tempMatrixLeft = new THREE.Matrix4();
        const tempMatrixRight = new THREE.Matrix4();

        return function(e) {
            forward.set(0, 0, -1);
            forward.transformDirection(this.camera.matrixWorld);
            tempMatrixLeft.makeRotationAxis(forward, Math.PI / 128);
            tempMatrixRight.makeRotationAxis(forward, -Math.PI / 128);
            switch (e.code) {
                case 'ArrowLeft':
                    this.camera.up.transformDirection(tempMatrixLeft);
                break;
                case 'ArrowRight':
                    this.camera.up.transformDirection(tempMatrixRight);
                break;
                case 'KeyC':
                    this.showMeshCursor = !this.showMeshCursor;
                break;
                case 'KeyP':
                    this.showControlPlane = !this.showControlPlane;
                break;
                case 'KeyI':
                    this.showInfo = !this.showInfo;
                    if (this.showInfo) {
                        this.infoPanel.style.display = 'block';
                    } else {
                        this.infoPanel.style.display = 'none';
                    }
                break;
            }
        };

    }();

    onMouseMove(mouse) {
        this.mousePosition.set(mouse.offsetX, mouse.offsetY);
    }

    onMouseDown() {
        this.mouseDownPosition.copy(this.mousePosition);
        this.mouseDownTime = getCurrentTime();
    }

    onMouseUp = function() {

        const clickOffset = new THREE.Vector2();

        return function(mouse) {
            clickOffset.copy(this.mousePosition).sub(this.mouseDownPosition);
            const mouseUpTime = getCurrentTime();
            const wasClick = mouseUpTime - this.mouseDownTime < 0.5 && clickOffset.length() < 2;
            if (wasClick) {
                this.onMouseClick(mouse);
            }
        };

    }();

    onMouseClick(mouse) {
        this.mousePosition.set(mouse.offsetX, mouse.offsetY);
        this.checkForFocalPointChange();
    }

    checkForFocalPointChange = function() {

        const renderDimensions = new THREE.Vector2();
        const toNewFocalPoint = new THREE.Vector3();
        const outHits = [];

        return function() {
            if (!this.transitioningCameraTarget) {
                this.getRenderDimensions(renderDimensions);
                outHits.length = 0;
                this.raycaster.setFromCameraAndScreenPosition(this.camera, this.mousePosition, renderDimensions);
                this.raycaster.intersectSplatMesh(this.splatMesh, outHits);
                if (outHits.length > 0) {
                    const hit = outHits[0];
                    const intersectionPoint = hit.origin;
                    toNewFocalPoint.copy(intersectionPoint).sub(this.camera.position);
                    if (toNewFocalPoint.length() > MINIMUM_DISTANCE_TO_NEW_FOCAL_POINT) {
                        this.previousCameraTarget.copy(this.controls.target);
                        this.nextCameraTarget.copy(intersectionPoint);
                        this.transitioningCameraTarget = true;
                        this.transitioningCameraTargetStartTime = getCurrentTime();
                    }
                }
            }
        };

    }();

    getRenderDimensions(outDimensions) {
        if (this.rootElement) {
            outDimensions.x = this.rootElement.offsetWidth;
            outDimensions.y = this.rootElement.offsetHeight;
        } else {
            this.renderer.getSize(outDimensions);
        }
    }

    setupInfoPanel() {
        this.infoPanel = document.createElement('div');
        this.infoPanel.style.position = 'absolute';
        this.infoPanel.style.padding = '10px';
        this.infoPanel.style.backgroundColor = '#cccccc';
        this.infoPanel.style.border = '#aaaaaa 1px solid';
        this.infoPanel.style.zIndex = 100;
        this.infoPanel.style.width = '375px';
        this.infoPanel.style.fontFamily = 'arial';
        this.infoPanel.style.fontSize = '10pt';
        this.infoPanel.style.textAlign = 'left';

        const layout = [
            ['Camera position', 'cameraPosition'],
            ['Camera look-at', 'cameraLookAt'],
            ['Camera up', 'cameraUp'],
            ['Cursor position', 'cursorPosition'],
            ['FPS', 'fps'],
            ['Render window', 'renderWindow'],
            ['Rendering:', 'renderSplatCount'],
            ['Sort time', 'sortTime']
        ];

        const infoTable = document.createElement('div');
        infoTable.style.display = 'table';

        for (let layoutEntry of layout) {
            const row = document.createElement('div');
            row.style.display = 'table-row';

            const labelCell = document.createElement('div');
            labelCell.style.display = 'table-cell';
            labelCell.style.width = '110px';
            labelCell.innerHTML = `${layoutEntry[0]}: `;

            const spacerCell = document.createElement('div');
            spacerCell.style.display = 'table-cell';
            spacerCell.style.width = '10px';
            spacerCell.innerHTML = ' ';

            const infoCell = document.createElement('div');
            infoCell.style.display = 'table-cell';
            infoCell.innerHTML = '';

            this.infoPanelCells[layoutEntry[1]] = infoCell;

            row.appendChild(labelCell);
            row.appendChild(spacerCell);
            row.appendChild(infoCell);

            infoTable.appendChild(row);
        }

        this.infoPanel.appendChild(infoTable);
        this.infoPanel.style.display = 'none';
        this.renderer.domElement.parentElement.prepend(this.infoPanel);
    }

    updateSplatMesh = function() {

        const renderDimensions = new THREE.Vector2();

        return function() {
            if (!this.splatMesh) return;
            const splatCount = this.splatMesh.getSplatCount();
            if (splatCount > 0) {
                this.splatMesh.updateTransforms();
                this.getRenderDimensions(renderDimensions);
                this.cameraFocalLengthX = this.camera.projectionMatrix.elements[0] *
                                          this.devicePixelRatio * renderDimensions.x * 0.45;
                                          this.cameraFocalLengthY = this.camera.projectionMatrix.elements[5] *
                                          this.devicePixelRatio * renderDimensions.y * 0.45;
                this.splatMesh.updateUniforms(renderDimensions, this.cameraFocalLengthX, this.cameraFocalLengthY);
            }
        };

    }();

    /**
     * Add a splat scene to the viewer.
     * @param {string} path Path to splat scene to be loaded
     * @param {object} options {
     *
     *         splatAlphaRemovalThreshold: Ignore any splats with an alpha less than the specified
     *                                     value (valid range: 0 - 255), defaults to 1
     *
     *         showLoadingSpinner:         Display a loading spinner while the scene is loading, defaults to true
     *
     *         position (Array<number>):   Position of the scene, acts as an offset from its default position, defaults to [0, 0, 0]
     *
     *         rotation (Array<number>):   Rotation of the scene represented as a quaternion, defaults to [0, 0, 0, 1]
     *
     *         scale (Array<number>):      Scene's scale, defaults to [1, 1, 1]
     *
     *         onProgress:                 Function to be called as file data are received
     *
     * }
     * @return {AbortablePromise}
     */
    addSplatScene(path, options = {}) {
        if (options.showLoadingSpinner !== false) options.showLoadingSpinner = true;
        if (options.showLoadingSpinner) this.loadingSpinner.show();
        const downloadProgress = (percent, percentLabel) => {
            if (options.showLoadingSpinner) {
                if (percent == 100) {
                    this.loadingSpinner.setMessage(`Download complete!`);
                } else {
                    const suffix = percentLabel ? `: ${percentLabel}` : `...`;
                    this.loadingSpinner.setMessage(`Downloading${suffix}`);
                }
            }
            if (options.onProgress) options.onProgress(percent, percentLabel, 'downloading');
        };
        const loadPromise = this.loadFileToSplatBuffer(path, options.splatAlphaRemovalThreshold, downloadProgress, options.format);
        return new AbortablePromise((resolve, reject) => {
            loadPromise.then((splatBuffer) => {
                if (options.showLoadingSpinner) this.loadingSpinner.hide();
                if (options.onProgress) options.onProgress(0, '0%', 'processing');
                const splatBufferOptions = {
                    'rotation': options.rotation || options.orientation,
                    'position': options.position,
                    'scale': options.scale,
                    'splatAlphaRemovalThreshold': options.splatAlphaRemovalThreshold,
                };
                this.addSplatBuffers([splatBuffer], [splatBufferOptions], options.showLoadingSpinner).then(() => {
                    if (options.onProgress) options.onProgress(100, '100%', 'processing');
                    resolve();
                });
            })
            .catch(() => {
                if (options.showLoadingSpinner) this.loadingSpinner.hide();
                reject(new Error(`Viewer::addSplatScene -> Could not load file ${path}`));
            });
        }, loadPromise.abortHandler);
    }

    /**
     * Add multiple splat scenes to the viewer.
     * @param {Array<object>} sceneOptions Array of per-scene options: {
     *
     *         path: Path to splat scene to be loaded
     *
     *         splatAlphaRemovalThreshold: Ignore any splats with an alpha less than the specified
     *                                     value (valid range: 0 - 255), defaults to 1
     *
     *         position (Array<number>):   Position of the scene, acts as an offset from its default position, defaults to [0, 0, 0]
     *
     *         rotation (Array<number>):   Rotation of the scene represented as a quaternion, defaults to [0, 0, 0, 1]
     *
     *         scale (Array<number>):      Scene's scale, defaults to [1, 1, 1]
     * }
     * @param {boolean} showLoadingSpinner Display a loading spinner while the scene is loading, defaults to true
     * @param {function} onProgress Function to be called as file data are received
     * @return {AbortablePromise}
     */
    addSplatScenes(sceneOptions, showLoadingSpinner = true, onProgress = undefined) {
        const fileCount = sceneOptions.length;
        const percentComplete = [];
        if (showLoadingSpinner) this.loadingSpinner.show();
        const downloadProgress = (fileIndex, percent, percentLabel) => {
            percentComplete[fileIndex] = percent;
            let totalPercent = 0;
            for (let i = 0; i < fileCount; i++) totalPercent += percentComplete[i] || 0;
            totalPercent = totalPercent / fileCount;
            percentLabel = `${totalPercent.toFixed(2)}%`;
            if (showLoadingSpinner) {
                if (totalPercent == 100) {
                    this.loadingSpinner.setMessage(`Download complete!`);
                } else {
                    this.loadingSpinner.setMessage(`Downloading: ${percentLabel}`);
                }
            }
            if (onProgress) onProgress(totalPercent, percentLabel, 'downloading');
        };

        const loadPromises = [];
        const abortHandlers = [];
        for (let i = 0; i < sceneOptions.length; i++) {
            const loadPromise = this.loadFileToSplatBuffer(sceneOptions[i].path, sceneOptions[i].splatAlphaRemovalThreshold,
                                                           downloadProgress.bind(this, i), sceneOptions.format);
            abortHandlers.push(loadPromise.abortHandler);
            loadPromises.push(loadPromise.promise);
        }
        const abortHandler = () => {
            for (let abortHandler of abortHandlers) {
                abortHandler();
            }
        };
        return new AbortablePromise((resolve, reject) => {
            Promise.all(loadPromises)
            .then((splatBuffers) => {
                if (showLoadingSpinner) this.loadingSpinner.hide();
                if (onProgress) options.onProgress(0, '0%', 'processing');
                this.addSplatBuffers(splatBuffers, sceneOptions, showLoadingSpinner).then(() => {
                    if (onProgress) onProgress(100, '100%', 'processing');
                    resolve();
                });
            })
            .catch(() => {
                if (showLoadingSpinner) this.loadingSpinner.hide();
                reject(new Error(`Viewer::addSplatScenes -> Could not load one or more splat scenes.`));
            });
        }, abortHandler);
    }

    /**
     *
     * @param {string} path Path to splat scene to be loaded
     * @param {number} splatAlphaRemovalThreshold Ignore any splats with an alpha less than the specified
     *                                            value (valid range: 0 - 255), defaults to 1
     *
     * @param {function} onProgress Function to be called as file data are received
     * @param {string} format Optional format specifier, if not specified the format will be inferred from the file extension
     * @return {AbortablePromise}
     */
    loadFileToSplatBuffer(path, splatAlphaRemovalThreshold = 1, onProgress = undefined, format = undefined) {
        const downloadProgress = (percent, percentLabel) => {
            if (onProgress) onProgress(percent, percentLabel, 'downloading');
        };
        if (format !== undefined && format !== null) {
            if (format === SceneFormat.Splat || format === SceneFormat.KSplat) {
                return new SplatLoader().loadFromURL(path, downloadProgress, 0, splatAlphaRemovalThreshold, undefined, undefined, format);
            } else if (format === SceneFormat.Ply) {
                return new PlyLoader().loadFromURL(path, downloadProgress, 0, splatAlphaRemovalThreshold);
            }
        } else {
            if (SplatLoader.isFileSplatFormat(path)) {
                return new SplatLoader().loadFromURL(path, downloadProgress, 0, splatAlphaRemovalThreshold);
            } else if (path.endsWith('.ply')) {
                return new PlyLoader().loadFromURL(path, downloadProgress, 0, splatAlphaRemovalThreshold);
            }
        }

        return AbortablePromise.reject(new Error(`Viewer::loadFileToSplatBuffer -> File format not supported: ${path}`));
    }

    /**
     * Add one or more instances of SplatBuffer to the SplatMesh instance managed by the viewer and set up the sorting web worker.
     * This function will terminate the existing sort worker (if there is one).
     */
    addSplatBuffers = function() {

        let loadPromise;
        let loadCount = 0;

        return function(splatBuffers, splatBufferOptions = [], showLoadingSpinner = true) {
            this.splatRenderingInitialized = false;
            loadCount++;
            const performLoad = () => {
                return new Promise((resolve) => {
                    if (showLoadingSpinner) {
                        this.loadingSpinner.show();
                        this.loadingSpinner.setMessage(`Processing splats...`);
                    }
                    window.setTimeout(() => {
                        this.disposeSortWorker();
                        this.addSplatBuffersToMesh(splatBuffers, splatBufferOptions);
                        this.setupSortWorker(this.splatMesh).then(() => {
                            loadCount--;
                            if (loadCount === 0) {
                                if (showLoadingSpinner) this.loadingSpinner.hide();
                                this.splatRenderingInitialized = true;
                            }
                            resolve();
                        });
                    }, 1);
                });
            };
            if (!loadPromise) {
                loadPromise = performLoad();
            } else {
                loadPromise = loadPromise.then(() => {
                    return performLoad();
                });
            }
            return loadPromise;
        };

    }();

    disposeSortWorker() {
        if (this.sortWorker) this.sortWorker.terminate();
        this.sortWorker = null;
        this.sortRunning = false;
    }

    /**
     * Add one or more instances of SplatBuffer to the SplatMesh instance managed by the viewer. This function is additive; all splat
     * buffers contained by the viewer's splat mesh before calling this function will be preserved.
     * @param {Array<SplatBuffer>} splatBuffers SplatBuffer instances
     * @param {Array<object>} splatBufferOptions Array of options objects: {
     *
     *         splatAlphaRemovalThreshold: Ignore any splats with an alpha less than the specified
     *                                     value (valid range: 0 - 255), defaults to 1
     *
     *         position (Array<number>):   Position of the scene, acts as an offset from its default position, defaults to [0, 0, 0]
     *
     *         rotation (Array<number>):   Rotation of the scene represented as a quaternion, defaults to [0, 0, 0, 1]
     *
     *         scale (Array<number>):      Scene's scale, defaults to [1, 1, 1]
     * }
     */
    addSplatBuffersToMesh(splatBuffers, splatBufferOptions) {
        const allSplatBuffers = this.splatMesh.splatBuffers || [];
        const allSplatBufferOptions = this.splatMesh.splatBufferOptions || [];
        allSplatBuffers.push(...splatBuffers);
        allSplatBufferOptions.push(...splatBufferOptions);
        this.splatMesh.build(allSplatBuffers, allSplatBufferOptions, true);
        if (this.renderer) this.splatMesh.setRenderer(this.renderer);
        this.splatMesh.frustumCulled = false;
    }

    /**
     * Set up the splat sorting web worker.
     * @param {SplatMesh} splatMesh SplatMesh instance that contains the splats to be sorted
     * @return {Promise}
     */
    setupSortWorker(splatMesh) {
        return new Promise((resolve) => {
            const DistancesArrayType = this.integerBasedSort ? Int32Array : Float32Array;
            const splatCount = splatMesh.getSplatCount();
            const sortWorker = createSortWorker(splatCount, this.sharedMemoryForWorkers, this.integerBasedSort, this.splatMesh.dynamicMode);
            sortWorker.onmessage = (e) => {
                if (e.data.sortDone) {
                    this.sortRunning = false;
                    if (this.sharedMemoryForWorkers) {
                        this.splatMesh.updateRenderIndexes(this.sortWorkerSortedIndexes, e.data.splatRenderCount);
                    } else {
                        const sortedIndexes = new Uint32Array(e.data.sortedIndexes, 0, e.data.splatRenderCount);
                        this.splatMesh.updateRenderIndexes(sortedIndexes, e.data.splatRenderCount);
                    }
                    this.lastSortTime = e.data.sortTime;
                    this.sortPromiseResolver();
                    this.sortPromise = null;
                    this.sortPromiseResolver = null;
                } else if (e.data.sortCanceled) {
                    this.sortRunning = false;
                } else if (e.data.sortSetupPhase1Complete) {
                    console.log('Sorting web worker WASM setup complete.');
                    const centers = this.integerBasedSort ? this.splatMesh.getIntegerCenters(true) : this.splatMesh.getFloatCenters(true);
                    const transformIndexes = this.splatMesh.getTransformIndexes();
                    sortWorker.postMessage({
                        'centers': centers.buffer,
                        'transformIndexes': transformIndexes.buffer
                    });
                    if (this.sharedMemoryForWorkers) {
                        this.sortWorkerSortedIndexes = new Uint32Array(e.data.sortedIndexesBuffer,
                                                                       e.data.sortedIndexesOffset, splatCount);
                        this.sortWorkerIndexesToSort = new Uint32Array(e.data.indexesToSortBuffer,
                                                                       e.data.indexesToSortOffset, splatCount);
                        this.sortWorkerPrecomputedDistances = new DistancesArrayType(e.data.precomputedDistancesBuffer,
                                                                                     e.data.precomputedDistancesOffset,
                                                                                     splatCount);
                         this.sortWorkerTransforms = new Float32Array(e.data.transformsBuffer,
                                                                      e.data.transformsOffset, Constants.MaxScenes * 16);
                    } else {
                        this.sortWorkerIndexesToSort = new Uint32Array(splatCount);
                        this.sortWorkerPrecomputedDistances = new DistancesArrayType(splatCount);
                        this.sortWorkerTransforms = new Float32Array(Constants.MaxScenes * 16);
                    }
                    for (let i = 0; i < splatCount; i++) this.sortWorkerIndexesToSort[i] = i;
                } else if (e.data.sortSetupComplete) {
                    console.log('Sorting web worker ready.');
                    const splatDataTextures = this.splatMesh.getSplatDataTextures();
                    const covariancesTextureSize = splatDataTextures.covariances.size;
                    const centersColorsTextureSize = splatDataTextures.centerColors.size;
                    console.log('Covariances texture size: ' + covariancesTextureSize.x + ' x ' + covariancesTextureSize.y);
                    console.log('Centers/colors texture size: ' + centersColorsTextureSize.x + ' x ' + centersColorsTextureSize.y);
                    this.sortWorker = sortWorker;
                    resolve();
                }
            };
        });
    }

    /**
     * Start self-driven mode
     */
    start() {
        if (this.selfDrivenMode) {
            if (this.webXRMode) {
                this.renderer.setAnimationLoop(this.selfDrivenUpdateFunc);
            } else {
                this.requestFrameId = requestAnimationFrame(this.selfDrivenUpdateFunc);
            }
            this.selfDrivenModeRunning = true;
        } else {
            throw new Error('Cannot start viewer unless it is in self driven mode.');
        }
    }

    /**
     * Stop self-driven mode
     */
    stop() {
        if (this.selfDrivenMode && this.selfDrivenModeRunning) {
            if (!this.webXRMode) {
                cancelAnimationFrame(this.requestFrameId);
            }
            this.selfDrivenModeRunning = false;
        }
    }

    /**
     * Dispose of all resources held directly and indirectly by this viewer.
     */
    async dispose() {
        if (this.sortPromise) {
            await this.sortPromise;
        }
        this.stop();
        if (this.controls) {
            this.controls.dispose();
            this.controls = null;
        }
        if (this.splatMesh) {
            this.splatMesh.dispose();
            this.splatMesh = null;
        }
        if (this.sceneHelper) {
            this.sceneHelper.dispose();
            this.sceneHelper = null;
        }
        if (this.resizeObserver) {
            this.resizeObserver.unobserve(this.rootElement);
            this.resizeObserver = null;
        }
        if (this.renderer) {
            if (!this.usingExternalRenderer) this.renderer.dispose();
            this.renderer = null;
        }
        this.disposeSortWorker();
        this.removeEventHandlers();
        this.camera = null;
        this.threeScene = null;
        this.splatRenderingInitialized = false;
        this.initialized = false;
    }

    selfDrivenUpdate() {
        if (this.selfDrivenMode && !this.webXRMode) {
            this.requestFrameId = requestAnimationFrame(this.selfDrivenUpdateFunc);
        }
        this.update();
        this.render();
    }

    render = function() {

        return function() {
            if (!this.initialized || !this.splatRenderingInitialized) return;
            const hasRenderables = (threeScene) => {
                for (let child of threeScene.children) {
                    if (child.visible) return true;
                }
                return false;
            };
            const savedAuoClear = this.renderer.autoClear;
            this.renderer.autoClear = false;
            if (hasRenderables(this.threeScene)) this.renderer.render(this.threeScene, this.camera);
            this.renderer.render(this.splatMesh, this.camera);
            if (this.sceneHelper.getFocusMarkerOpacity() > 0.0) this.renderer.render(this.sceneHelper.focusMarker, this.camera);
            if (this.showControlPlane) this.renderer.render(this.sceneHelper.controlPlane, this.camera);
            this.renderer.autoClear = savedAuoClear;
        };

    }();

    update(renderer, camera) {
        if (this.dropInMode) this.updateForDropInMode(renderer, camera);
        if (!this.initialized || !this.splatRenderingInitialized) return;
        if (this.controls) this.controls.update();
        this.updateSplatSort();
        this.updateForRendererSizeChanges();
        this.updateSplatMesh();
        this.updateMeshCursor();
        this.updateFPS();
        this.timingSensitiveUpdates();
        this.updateInfoPanel();
        this.updateControlPlane();
    }

    updateForDropInMode(renderer, camera) {
        this.renderer = renderer;
        if (this.splatMesh) this.splatMesh.setRenderer(this.renderer);
        this.camera = camera;
        if (this.controls) this.controls.object = camera;
        this.init();
    }

    updateFPS = function() {

        let lastCalcTime = getCurrentTime();
        let frameCount = 0;

        return function() {
            const currentTime = getCurrentTime();
            const calcDelta = currentTime - lastCalcTime;
            if (calcDelta >= 1.0) {
                this.currentFPS = frameCount;
                frameCount = 0;
                lastCalcTime = currentTime;
            } else {
                frameCount++;
            }
        };

    }();

    updateForRendererSizeChanges = function() {

        const lastRendererSize = new THREE.Vector2();
        const currentRendererSize = new THREE.Vector2();

        return function() {
            this.renderer.getSize(currentRendererSize);
            if (currentRendererSize.x !== lastRendererSize.x || currentRendererSize.y !== lastRendererSize.y) {
                if (!this.usingExternalCamera) {
                    this.camera.aspect = currentRendererSize.x / currentRendererSize.y;
                    this.camera.updateProjectionMatrix();
                }
                lastRendererSize.copy(currentRendererSize);
            }
        };

    }();

    timingSensitiveUpdates = function() {

        let lastUpdateTime;

        return function() {
            const currentTime = getCurrentTime();
            if (!lastUpdateTime) lastUpdateTime = currentTime;
            const timeDelta = currentTime - lastUpdateTime;

            this.updateCameraTransition(currentTime);
            this.updateFocusMarker(timeDelta);

            lastUpdateTime = currentTime;
        };

    }();

    updateCameraTransition = function() {

        let tempCameraTarget = new THREE.Vector3();
        let toPreviousTarget = new THREE.Vector3();
        let toNextTarget = new THREE.Vector3();

        return function(currentTime) {
            if (this.transitioningCameraTarget) {
                toPreviousTarget.copy(this.previousCameraTarget).sub(this.camera.position).normalize();
                toNextTarget.copy(this.nextCameraTarget).sub(this.camera.position).normalize();
                const rotationAngle = Math.acos(toPreviousTarget.dot(toNextTarget));
                const rotationSpeed = rotationAngle / (Math.PI / 3) * .65 + .3;
                const t = (rotationSpeed / rotationAngle * (currentTime - this.transitioningCameraTargetStartTime));
                tempCameraTarget.copy(this.previousCameraTarget).lerp(this.nextCameraTarget, t);
                this.camera.lookAt(tempCameraTarget);
                this.controls.target.copy(tempCameraTarget);
                if (t >= 1.0) {
                    this.transitioningCameraTarget = false;
                }
            }
        };

    }();

    updateFocusMarker = function() {

        const renderDimensions = new THREE.Vector2();
        let wasTransitioning = false;

        return function(timeDelta) {
            this.getRenderDimensions(renderDimensions);
            const fadeInSpeed = 10.0;
            const fadeOutSpeed = 2.5;
            if (this.transitioningCameraTarget) {
                this.sceneHelper.setFocusMarkerVisibility(true);
                const currentFocusMarkerOpacity = Math.max(this.sceneHelper.getFocusMarkerOpacity(), 0.0);
                let newFocusMarkerOpacity = Math.min(currentFocusMarkerOpacity + fadeInSpeed * timeDelta, 1.0);
                this.sceneHelper.setFocusMarkerOpacity(newFocusMarkerOpacity);
                this.sceneHelper.updateFocusMarker(this.nextCameraTarget, this.camera, renderDimensions);
                wasTransitioning = true;
            } else {
                let currentFocusMarkerOpacity;
                if (wasTransitioning) currentFocusMarkerOpacity = 1.0;
                else currentFocusMarkerOpacity = Math.min(this.sceneHelper.getFocusMarkerOpacity(), 1.0);
                if (currentFocusMarkerOpacity > 0) {
                    this.sceneHelper.updateFocusMarker(this.nextCameraTarget, this.camera, renderDimensions);
                    let newFocusMarkerOpacity = Math.max(currentFocusMarkerOpacity - fadeOutSpeed * timeDelta, 0.0);
                    this.sceneHelper.setFocusMarkerOpacity(newFocusMarkerOpacity);
                    if (newFocusMarkerOpacity === 0.0) this.sceneHelper.setFocusMarkerVisibility(false);
                }
                wasTransitioning = false;
            }
        };

    }();

    updateMeshCursor = function() {

        const outHits = [];
        const renderDimensions = new THREE.Vector2();

        return function() {
            if (this.showMeshCursor) {
                this.getRenderDimensions(renderDimensions);
                outHits.length = 0;
                this.raycaster.setFromCameraAndScreenPosition(this.camera, this.mousePosition, renderDimensions);
                this.raycaster.intersectSplatMesh(this.splatMesh, outHits);
                if (outHits.length > 0) {
                    this.sceneHelper.setMeshCursorVisibility(true);
                    this.sceneHelper.positionAndOrientMeshCursor(outHits[0].origin, this.camera);
                } else {
                    this.sceneHelper.setMeshCursorVisibility(false);
                }
            } else {
                this.sceneHelper.setMeshCursorVisibility(false);
            }
        };

    }();

    updateInfoPanel = function() {

        const renderDimensions = new THREE.Vector2();

        return function() {
            if (!this.showInfo) return;
            const splatCount = this.splatMesh.getSplatCount();
            this.getRenderDimensions(renderDimensions);

            const cameraPos = this.camera.position;
            const cameraPosString = `[${cameraPos.x.toFixed(5)}, ${cameraPos.y.toFixed(5)}, ${cameraPos.z.toFixed(5)}]`;
            this.infoPanelCells.cameraPosition.innerHTML = cameraPosString;

            if (this.controls) {
            const cameraLookAt = this.controls.target;
            const cameraLookAtString = `[${cameraLookAt.x.toFixed(5)}, ${cameraLookAt.y.toFixed(5)}, ${cameraLookAt.z.toFixed(5)}]`;
            this.infoPanelCells.cameraLookAt.innerHTML = cameraLookAtString;
            }

            const cameraUp = this.camera.up;
            const cameraUpString = `[${cameraUp.x.toFixed(5)}, ${cameraUp.y.toFixed(5)}, ${cameraUp.z.toFixed(5)}]`;
            this.infoPanelCells.cameraUp.innerHTML = cameraUpString;

            if (this.showMeshCursor) {
                const cursorPos = this.sceneHelper.meshCursor.position;
                const cursorPosString = `[${cursorPos.x.toFixed(5)}, ${cursorPos.y.toFixed(5)}, ${cursorPos.z.toFixed(5)}]`;
                this.infoPanelCells.cursorPosition.innerHTML = cursorPosString;
            } else {
                this.infoPanelCells.cursorPosition.innerHTML = 'N/A';
            }

            this.infoPanelCells.fps.innerHTML = this.currentFPS;
            this.infoPanelCells.renderWindow.innerHTML = `${renderDimensions.x} x ${renderDimensions.y}`;

            const renderPct = this.splatRenderCount / splatCount * 100;
            this.infoPanelCells.renderSplatCount.innerHTML =
                `${this.splatRenderCount} splats out of ${splatCount} (${renderPct.toFixed(2)}%)`;

            this.infoPanelCells.sortTime.innerHTML = `${this.lastSortTime.toFixed(3)} ms`;
        };

    }();

    updateControlPlane() {
        if (this.showControlPlane) {
            this.sceneHelper.setControlPlaneVisibility(true);
            this.sceneHelper.positionAndOrientControlPlane(this.controls.target, this.camera.up);
        } else {
            this.sceneHelper.setControlPlaneVisibility(false);
        }
    }

    updateSplatSort = function() {

        const mvpMatrix = new THREE.Matrix4();
        const cameraPositionArray = [];
        const lastSortViewDir = new THREE.Vector3(0, 0, -1);
        const sortViewDir = new THREE.Vector3(0, 0, -1);
        const lastSortViewPos = new THREE.Vector3();
        const sortViewOffset = new THREE.Vector3();
        const queuedSorts = [];
        let runCount = 0;

        const partialSorts = [
            {
                'angleThreshold': 0.55,
                'sortFractions': [0.125, 0.33333, 0.75]
            },
            {
                'angleThreshold': 0.65,
                'sortFractions': [0.33333, 0.66667]
            },
            {
                'angleThreshold': 0.8,
                'sortFractions': [0.5]
            }
        ];

        return async function(force = false, gatherAllNodes = false) {
            if (this.sortRunning) return;
            if (!this.initialized || !this.splatRenderingInitialized) return;

            let angleDiff = 0;
            let positionDiff = 0;
            let needsRefreshForRotation = false;
            let needsRefreshForPosition = false;

            sortViewDir.set(0, 0, -1).applyQuaternion(this.camera.quaternion);
            angleDiff = sortViewDir.dot(lastSortViewDir);
            positionDiff = sortViewOffset.copy(this.camera.position).sub(lastSortViewPos).length();

            if (!force && !this.splatMesh.dynamicMode && queuedSorts.length === 0 && runCount > 0) {
                if (angleDiff <= 0.95) needsRefreshForRotation = true;
                if (positionDiff >= 1.0) needsRefreshForPosition = true;
                if (!needsRefreshForRotation && !needsRefreshForPosition) return;
            }

            this.sortRunning = true;
            this.splatRenderCount = this.gatherSceneNodesForSort(gatherAllNodes);
            this.sortPromise = new Promise((resolve) => {
                this.sortPromiseResolver = resolve;
            });

            mvpMatrix.copy(this.camera.matrixWorld).invert();
            mvpMatrix.premultiply(this.camera.projectionMatrix);
            mvpMatrix.multiply(this.splatMesh.matrixWorld);

            if (this.gpuAcceleratedSort && (queuedSorts.length <= 1 || queuedSorts.length % 2 === 0)) {
                await this.splatMesh.computeDistancesOnGPU(mvpMatrix, this.sortWorkerPrecomputedDistances);
            }

            if (this.splatMesh.dynamicMode) {
                queuedSorts.push(this.splatRenderCount);
            } else {
                if (queuedSorts.length === 0) {
                    for (let partialSort of partialSorts) {
                        if (angleDiff < partialSort.angleThreshold) {
                            for (let sortFraction of partialSort.sortFractions) {
                                queuedSorts.push(Math.floor(this.splatRenderCount * sortFraction));
                            }
                            break;
                        }
                    }
                    queuedSorts.push(this.splatRenderCount);
                }
            }
            const sortCount = Math.min(queuedSorts.shift(), this.splatRenderCount);

            cameraPositionArray[0] = this.camera.position.x;
            cameraPositionArray[1] = this.camera.position.y;
            cameraPositionArray[2] = this.camera.position.z;

            const sortMessage = {
                'modelViewProj': mvpMatrix.elements,
                'cameraPosition': cameraPositionArray,
                'splatRenderCount': this.splatRenderCount,
                'splatSortCount': sortCount,
                'usePrecomputedDistances': this.gpuAcceleratedSort
            };
            if (this.splatMesh.dynamicMode) {
                this.splatMesh.fillTransformsArray(this.sortWorkerTransforms);
            }
            if (!this.sharedMemoryForWorkers) {
                sortMessage.indexesToSort = this.sortWorkerIndexesToSort;
                sortMessage.transforms = this.sortWorkerTransforms;
                if (this.gpuAcceleratedSort) {
                    sortMessage.precomputedDistances = this.sortWorkerPrecomputedDistances;
                }
            }
            this.sortWorker.postMessage({
                'sort': sortMessage
            });

            if (queuedSorts.length === 0) {
                lastSortViewPos.copy(this.camera.position);
                lastSortViewDir.copy(sortViewDir);
            }
            runCount++;
        };

    }();

    /**
     * Determine which splats to render by checking which are inside or close to the view frustum
     */
    gatherSceneNodesForSort = function() {

        const nodeRenderList = [];
        const tempVectorYZ = new THREE.Vector3();
        const tempVectorXZ = new THREE.Vector3();
        const tempVector = new THREE.Vector3();
        const modelView = new THREE.Matrix4();
        const baseModelView = new THREE.Matrix4();
        const sceneTransform = new THREE.Matrix4();
        const renderDimensions = new THREE.Vector3();
        const forward = new THREE.Vector3(0, 0, -1);

        const tempMax = new THREE.Vector3();
        const nodeSize = (node) => {
            return tempMax.copy(node.max).sub(node.min).length();
        };

        const MaximumDistanceToRender = 125;

        return function(gatherAllNodes) {

            this.getRenderDimensions(renderDimensions);
            const cameraFocalLength = (renderDimensions.y / 2.0) / Math.tan(this.camera.fov / 2.0 * THREE.MathUtils.DEG2RAD);
            const fovXOver2 = Math.atan(renderDimensions.x / 2.0 / cameraFocalLength);
            const fovYOver2 = Math.atan(renderDimensions.y / 2.0 / cameraFocalLength);
            const cosFovXOver2 = Math.cos(fovXOver2);
            const cosFovYOver2 = Math.cos(fovYOver2);

            const splatTree = this.splatMesh.getSplatTree();
            baseModelView.copy(this.camera.matrixWorld).invert();
            baseModelView.multiply(this.splatMesh.matrixWorld);

            let nodeRenderCount = 0;
            let splatRenderCount = 0;

            for (let s = 0; s < splatTree.subTrees.length; s++) {
                const subTree = splatTree.subTrees[s];
                modelView.copy(baseModelView);
                if (this.splatMesh.dynamicMode) {
                    this.splatMesh.getSceneTransform(s, sceneTransform);
                    modelView.multiply(sceneTransform);
                }
                const nodeCount = subTree.nodesWithIndexes.length;
                for (let i = 0; i < nodeCount; i++) {
                    const node = subTree.nodesWithIndexes[i];
                    tempVector.copy(node.center).applyMatrix4(modelView);

                    const distanceToNode = tempVector.length();
                    tempVector.normalize();

                    tempVectorYZ.copy(tempVector).setX(0).normalize();
                    tempVectorXZ.copy(tempVector).setY(0).normalize();

                    const cameraAngleXZDot = forward.dot(tempVectorXZ);
                    const cameraAngleYZDot = forward.dot(tempVectorYZ);

                    const ns = nodeSize(node);
                    const outOfFovY = cameraAngleYZDot < (cosFovYOver2 - .6);
                    const outOfFovX = cameraAngleXZDot < (cosFovXOver2 - .6);
                    if (!gatherAllNodes && ((outOfFovX || outOfFovY || distanceToNode > MaximumDistanceToRender) && distanceToNode > ns)) {
                        continue;
                    }
                    splatRenderCount += node.data.indexes.length;
                    nodeRenderList[nodeRenderCount] = node;
                    node.data.distanceToNode = distanceToNode;
                    nodeRenderCount++;
                }
            }

            nodeRenderList.length = nodeRenderCount;
            nodeRenderList.sort((a, b) => {
                if (a.data.distanceToNode < b.data.distanceToNode) return -1;
                else return 1;
            });

            let currentByteOffset = splatRenderCount * Constants.BytesPerInt;
            for (let i = 0; i < nodeRenderCount; i++) {
                const node = nodeRenderList[i];
                const windowSizeInts = node.data.indexes.length;
                const windowSizeBytes = windowSizeInts * Constants.BytesPerInt;
                let destView = new Uint32Array(this.sortWorkerIndexesToSort.buffer, currentByteOffset - windowSizeBytes, windowSizeInts);
                destView.set(node.data.indexes);
                currentByteOffset -= windowSizeBytes;
            }

            return splatRenderCount;
        };

    }();

    getSplatMesh() {
        return this.splatMesh;
    }

    /**
     * Get a reference to a splat scene.
     * @param {number} sceneIndex The index of the scene to which the reference will be returned
     * @return {SplatScene}
     */
    getSplatScene(sceneIndex) {
        return this.splatMesh.getScene(sceneIndex);
    }

    isMobile() {
        return navigator.userAgent.includes('Mobi');
    }
}

/**
 * DropInViewer: Wrapper for a Viewer instance that enables it to be added to a Three.js scene like
 * any other Three.js scene object (Mesh, Object3D, etc.)
 */
class DropInViewer extends THREE.Group {

    constructor(options = {}) {
        super();

        options.selfDrivenMode = false;
        options.useBuiltInControls = false;
        options.rootElement = null;
        options.ignoreDevicePixelRatio = false;
        options.dropInMode = true;
        options.camera = undefined;
        options.renderer = undefined;

        this.viewer = new Viewer(options);

        this.callbackMesh = DropInViewer.createCallbackMesh();
        this.add(this.callbackMesh);
        this.callbackMesh.onBeforeRender = DropInViewer.onBeforeRender.bind(this, this.viewer);

    }

    /**
     * Add a single splat scene to the viewer.
     * @param {string} path Path to splat scene to be loaded
     * @param {object} options {
     *
     *         splatAlphaRemovalThreshold: Ignore any splats with an alpha less than the specified
     *                                     value (valid range: 0 - 255), defaults to 1
     *
     *         showLoadingSpinner:         Display a loading spinner while the scene is loading, defaults to true
     *
     *         position (Array<number>):   Position of the scene, acts as an offset from its default position, defaults to [0, 0, 0]
     *
     *         rotation (Array<number>):   Rotation of the scene represented as a quaternion, defaults to [0, 0, 0, 1]
     *
     *         scale (Array<number>):      Scene's scale, defaults to [1, 1, 1]
     *
     *         onProgress:                 Function to be called as file data are received
     *
     * }
     * @return {AbortablePromise}
     */
    addSplatScene(path, options = {}) {
        if (options.showLoadingSpinner !== false) options.showLoadingSpinner = true;
        const loadPromise = this.viewer.addSplatScene(path, options);
        loadPromise.then(() => {
            this.add(this.viewer.splatMesh);
        });
        return loadPromise;
    }

    /**
     * Add multiple splat scenes to the viewer.
     * @param {Array<object>} sceneOptions Array of per-scene options: {
     *
     *         path: Path to splat scene to be loaded
     *
     *         splatAlphaRemovalThreshold: Ignore any splats with an alpha less than the specified
     *                                     value (valid range: 0 - 255), defaults to 1
     *
     *         position (Array<number>):   Position of the scene, acts as an offset from its default position, defaults to [0, 0, 0]
     *
     *         rotation (Array<number>):   Rotation of the scene represented as a quaternion, defaults to [0, 0, 0, 1]
     *
     *         scale (Array<number>):      Scene's scale, defaults to [1, 1, 1]
     * }
     * @param {boolean} showLoadingSpinner Display a loading spinner while the scene is loading, defaults to true
     * @return {AbortablePromise}
     */
    addSplatScenes(sceneOptions, showLoadingSpinner) {
        if (showLoadingSpinner !== false) showLoadingSpinner = true;
        const loadPromise = this.viewer.addSplatScenes(sceneOptions, showLoadingSpinner);
        loadPromise.then(() => {
            this.add(this.viewer.splatMesh);
        });
        return loadPromise;
    }

    /**
     * Get a reference to a splat scene.
     * @param {number} sceneIndex The index of the scene to which the reference will be returned
     * @return {SplatScene}
     */
    getSplatScene(sceneIndex) {
        return this.viewer.getSplatScene(sceneIndex);
    }

    dispose() {
        return this.viewer.dispose();
    }

    static onBeforeRender(viewer, renderer, threeScene, camera) {
        viewer.update(renderer, camera);
    }

    static createCallbackMesh() {
        const geometry = new THREE.SphereGeometry(1, 8, 8);
        const material = new THREE.MeshBasicMaterial();
        material.colorWrite = false;
        material.depthWrite = false;
        const mesh = new THREE.Mesh(geometry, material);
        mesh.frustumCulled = false;
        return mesh;
    }

}

export { AbortablePromise, DropInViewer, OrbitControls, PlyLoader, PlyParser, SceneFormat, SplatBuffer, SplatCompressor, SplatLoader, Viewer, WebXRMode };
//# sourceMappingURL=gaussian-splats-3d.module.js.map
