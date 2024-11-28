/**
 * Get the file extension if any.
 * 
 * @param {string} path 
 * @returns string
 */
export function getFileExtension(path) {
    return path.split('.').pop().toLowerCase()
}

export function getLastSegmentOfURL(url) {
    const parts = url.split('/');
    return parts.pop() || parts.pop();  // handle potential trailing slash
}