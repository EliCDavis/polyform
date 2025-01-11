
export class NoteManager {

    constructor(requestManager, flowGraph) {
        this.requestManager = requestManager;
        this.flowGraph = flowGraph;

        flowGraph.addNoteAddedListener(this.noteAdded.bind(this));
    }

    noteAdded(addedNote /*FlowNote*/) {
        addedNote.addWidthChangeListener(this.noteWidthChange.bind(this));
        addedNote.addContentChangeListener(this.noteContentChange.bind(this));
    }

    noteWidthChange(node /*FlowNote*/, newWidth /*number*/) {

    }

    noteContentChange(node /*FlowNote*/, newContents /*number*/) {

    }
}