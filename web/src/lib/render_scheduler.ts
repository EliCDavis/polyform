let pending = true;
let holds = 0;

export function requestRender(): void {
    pending = true;
}

export function renderContinuously(): () => void {
    holds++;

    let released = false;
    return () => {
        if (released) {
            return;
        }
        released = true;
        holds--;
        requestRender();
    };
}

export function consumeRenderRequest(): boolean {
    if (holds > 0) {
        pending = false;
        return true;
    }

    if (!pending) {
        return false;
    }

    pending = false;
    return true;
}
