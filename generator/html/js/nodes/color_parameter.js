
function BuildColorParameterNode(app) {
    const node = LiteGraph.createNode("polyform/color");
    console.log(node)
    app.LightGraph.add(node);
    return node;
}

export class ColorParameter {

    constructor(nodeManager, id, parameterData, app) {
        this.lightNode = BuildColorParameterNode(app);
        this.lightNode.title = parameterData.name;

        this.lightNode.widgets[0].value = parameterData.currentValue;
        this.lightNode.widgets[0].mouse = (event, pos, node) => {
            if (event.type !== "mouseup") {
                return;
            }
            app.ColorSelector.show(this.lightNode.widgets[0].value, (newColor) => {
                console.log(newColor);
                this.lightNode.widgets[0].value = newColor;
                app.LightGraph.dirty_canvas = true;
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: newColor
                });
            })
            console.log("clicked!", event)
        }
        // this.lightNode.setSize(this.lightNode.computeSize());

        // this.lightNode.onDropFile = (file) => {
        //     // console.log(file)
        //     var reader = new FileReader();
        //     reader.onload = (evt) => {
        //         console.log(evt.target.result)
        //         nodeManager.nodeParameterChanged({
        //             id: id,
        //             data: evt.target.result,
        //             binary: true
        //         });
        //     }
        //     reader.readAsArrayBuffer(file);

        //     const url = URL.createObjectURL(file);
        //     // this.loadImage(this._url, function (img) {
        //     //     that.size[1] = (img.height / img.width) * that.size[0];
        //     // });
        //     const img = document.createElement("img");
        //     img.src = url;
        //     img.onload = () => {
        //         // if (callback) {
        //         // callback(this);
        //         // }
        //         // console.log("Image loaded, size: " + img.width + "x" + img.height);
        //         // this.dirty = true;
        //         // that.boxcolor = "#9F9";
        //         // that.setDirtyCanvas(true);
        //         this.lightNode.widgets[0].image = img
        //         this.lightNode.setSize(this.lightNode.computeSize());
        //     };
        //     img.onerror = () => {
        //         console.log("error loading the image:" + url);
        //     }
        // }
    }

    update(parameterData) {
        this.lightNode.widgets[0].value = parameterData.currentValue;
    }

}     