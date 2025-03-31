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