import { RequestManager } from "./requests";
import { NodeFlowGraph, FlowNote } from "@elicdavis/node-flow";
import { GraphInstance } from "./schema";
import { getScopedNotes } from "./graphScope";
import { ROOT_SCOPE, type GraphScope } from "./portTypes";

const ID_PROPERTY: string = "id";

export class NoteManager {

    requestManager: RequestManager;

    flowGraph: NodeFlowGraph;

    updating: boolean;

    notes: Map<string, FlowNote>;

    graphScope: GraphScope = ROOT_SCOPE;

    constructor(requestManager: RequestManager, flowGraph: NodeFlowGraph) {
        this.requestManager = requestManager;
        this.flowGraph = flowGraph;

        this.notes = new Map<string, FlowNote>();
        this.updating = false;

        flowGraph.addNoteAddedListener(this.noteAdded.bind(this));
        flowGraph.addNoteRemovedListener(this.noteRemoved.bind(this));
        flowGraph.addNoteDragStopListener(this.noteDragStopped.bind(this));
    }

    generateID() {
        let maxID = 0;
        this.notes.forEach((note, ID) => {
            maxID = Math.max(maxID, parseInt(ID))
        })
        return "" + (maxID + 1);
    }

    clearCanvasNotes(): void {
        this.updating = true;
        this.notes.forEach((note) => {
            this.flowGraph.removeNote(note);
        });
        this.notes.clear();
        this.updating = false;
    }

    switchGraphScope(scope: GraphScope, schema: GraphInstance): void {
        this.graphScope = scope;
        this.clearCanvasNotes();
        this.schemaUpdate(schema);
    }

    /** Keep in-memory note layout in sync before a tab/scope swap. */
    syncLivePositionsIntoSchema(schema: GraphInstance): void {
        const scopedNotes = getScopedNotes(schema, this.graphScope);
        if (!scopedNotes) {
            return;
        }
        this.notes.forEach((note, noteId) => {
            const noteData = scopedNotes[noteId];
            if (!noteData) {
                return;
            }
            const pos = note.position();
            noteData.position = {
                x: Math.round(pos.x),
                y: Math.round(pos.y),
            };
        });
    }

    schemaUpdate(newSchema: GraphInstance) {
        const schemaNotes = getScopedNotes(newSchema, this.graphScope) ?? {};

        this.updating = true;

        for (const noteID of [...this.notes.keys()]) {
            if (!(noteID in schemaNotes)) {
                this.flowGraph.removeNote(this.notes.get(noteID)!);
                this.notes.delete(noteID);
            }
        }

        for (const noteID in schemaNotes) {
            const noteData = schemaNotes[noteID];

            if (this.notes.has(noteID)) {

                // Update the note
                const noteToUpdate = this.notes.get(noteID)

                noteToUpdate.setPosition(noteData.position);
                noteToUpdate.setText(noteData.text);
                noteToUpdate.setWidth(noteData.width);

            } else {

                // We gotta create the note
                const note = new FlowNote({
                    text: noteData.text,
                    width: noteData.width,
                    position: noteData.position
                });
                note.setMetadataProperty(ID_PROPERTY, noteID);
                this.flowGraph.addNote(note);
                this.setupNote(note);
                this.notes.set(noteID, note);
            }
        }

        this.updating = false;
    }

    noteAdded(addedNote: FlowNote): void {
        if (this.updating) {
            return;
        }
        const id = this.generateID();
        addedNote.setMetadataProperty(ID_PROPERTY, id);
        this.requestManager.createNote(
            id,
            {
                "position": addedNote.position(),
                "text": addedNote.text(),
                "width": addedNote.width()
            },
            () => { }
        );

        this.notes.set(id, addedNote);
        this.setupNote(addedNote);
    }

    noteRemoved(removedNode: FlowNote): void {
        if (this.updating) {
            return;
        }
        this.requestManager.deleteMetadata(`notes/${removedNode.getMetadataProperty(ID_PROPERTY)}`)
    }

    noteDragStopped(noteDragged: FlowNote): void {
        if (this.updating) {
            return;
        }
        
        // Round to decrease file size of json. Precision isn't needed
        const pos = noteDragged.position();
        pos.x = Math.round(pos.x);
        pos.y = Math.round(pos.y);

        this.requestManager.setNoteMetadata(noteDragged.getMetadataProperty(ID_PROPERTY), "position", pos, () => { });
    }

    setupNote(note: FlowNote): void {
        note.addWidthChangeListener(this.noteWidthChange.bind(this));
        note.addContentChangeListener(this.noteContentChange.bind(this));
    }

    noteWidthChange(node: FlowNote, newWidth: number): void {
        if (this.updating) {
            return;
        }
        this.requestManager.setNoteMetadata(node.getMetadataProperty(ID_PROPERTY), "width", newWidth, () => { });
    }

    noteContentChange(node: FlowNote, newContents: string): void {
        if (this.updating) {
            return;
        }
        this.requestManager.setNoteMetadata(node.getMetadataProperty(ID_PROPERTY), "text", newContents, () => { });
    }
}
