
export class Observable<T> {

    private currentValue: T;

    constructor(startValue: T) {
        this.currentValue = startValue;
    }

    set(value: T): void {
        this.currentValue = value;
    }

    value(): T {
        return this.currentValue;
    }

}