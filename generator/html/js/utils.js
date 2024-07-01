
export function addRenderableImageWidget(node) {
    const H = LiteGraph.NODE_WIDGET_HEIGHT;

    const imgWidget = node.addWidget("image", "Image", true, { property: "surname" }); //this will modify the node.properties
    node.imgWidget = imgWidget;
    const margin = 15;
    node.imgWidget.draw = (ctx, node, widget_width, y, H) => {
        if (!imgWidget.image) {
            return;
        }

        const adjustedWidth = widget_width - margin * 2
        ctx.drawImage(
            imgWidget.image,
            margin,
            y,
            adjustedWidth,
            (adjustedWidth / imgWidget.image.width) * imgWidget.image.height
        );
    }

    node.imgWidget.computeSize = (width) => {
        if (!!imgWidget.image) {
            const adjustedWidth = width - margin * 2
            const newH = (adjustedWidth / imgWidget.image.width) * imgWidget.image.height;
            return [width, newH]
        }
        return [width, 0];
    }
}

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