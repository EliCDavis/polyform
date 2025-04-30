/**
 * Get the file extension if any.
 * 
 * @param {string} path 
 * @returns string
 */
export function getFileExtension(path: string): string {
    if (!path) {
        return "";
    }

    const split = path.split('.')
    if (split.length === 0) {
        return "";
    }

    return split[split.length - 1].toLowerCase()
}

export function getLastSegmentOfURL(url: string): string {
    const parts = url.split('/');
    
      // handle potential trailing slash
    const popped = parts.pop() || parts.pop();
    
    if (popped) {
        return popped;
    }
    return "";
}

export const Compress = async (str: string, encoding = 'gzip' as CompressionFormat): Promise<ArrayBuffer> => {
    const byteArray = new TextEncoder().encode(str)
    const cs = new CompressionStream(encoding)
    const writer = cs.writable.getWriter()
    writer.write(byteArray)
    writer.close()
    return new Response(cs.readable).arrayBuffer()
}

export const Decompress = async (byteArray: BufferSource, encoding = 'gzip' as CompressionFormat): Promise<string> => {
    const cs = new DecompressionStream(encoding)
    const writer = cs.writable.getWriter()
    writer.write(byteArray)
    writer.close()
    const arrayBuffer = await new Response(cs.readable).arrayBuffer()
    return new TextDecoder().decode(arrayBuffer)
}

export async function CopyToClipboard(text: string) {
    try {
        await navigator.clipboard.writeText(text);
        console.log('Text copied to clipboard');
    } catch (err) {
        console.error('Failed to copy text: ', err);
    }
}

export function ArrayBufferToBase64(buffer: ArrayBuffer) {
    return new Promise((resolve, reject) => {
        let blob = new Blob([buffer]);
        let reader = new FileReader();
        reader.onloadend = () => resolve(reader.result);
        reader.onerror = reject;
        reader.readAsDataURL(blob);
    });
}
