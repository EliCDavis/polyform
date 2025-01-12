
export class NoteManager {

    constructor(requestManager, flowGraph) {
        this.requestManager = requestManager;
        this.flowGraph = flowGraph;

        this.notes = new Map();
        this.updating = false;

        flowGraph.addNoteAddedListener(this.noteAdded.bind(this));
        flowGraph.addNoteDragStopListener(this.noteDragStopped.bind(this));
    }

    schemaUpdate(newSchema) {
        const schemaNotes = newSchema.notes;
        if (!schemaNotes) {
            return;
        }

        this.updating = true;

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
                note.id = noteID;
                this.flowGraph.addNote(note);
                this.setupNote(note);
                this.notes.set(noteID, note);
            }
        }

        // TODO: Delete notes that aren't apart of the schema

        this.updating = false;
    }

    noteAdded(addedNote /*FlowNote*/) {
        if (this.updating) {
            return;
        }
        addedNote.id = "" + this.notes.size;
        this.requestManager.createNote(
            addedNote.id,
            {
                "position": addedNote.position(),
                "text": addedNote.text(),
                "width": addedNote.width()
            },
            () => { }
        );

        this.notes.set(addedNote.id, addedNote);
        this.setupNote(addedNote);
    }

    noteDragStopped(noteDragged /*FlowNote*/) {
        if (this.updating) {
            return;
        }
        this.requestManager.setNoteMetadata(noteDragged.id, "position", noteDragged.position(), () => { });
    }

    setupNote(note) {
        note.addWidthChangeListener(this.noteWidthChange.bind(this));
        note.addContentChangeListener(this.noteContentChange.bind(this));
    }

    noteWidthChange(node /*FlowNote*/, newWidth /*number*/) {
        if (this.updating) {
            return;
        }
        this.requestManager.setNoteMetadata(node.id, "width", newWidth, () => { });
    }

    noteContentChange(node /*FlowNote*/, newContents /*string*/) {
        if (this.updating) {
            return;
        }
        this.requestManager.setNoteMetadata(node.id, "text", newContents, () => { });
    }
}