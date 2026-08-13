export class TemperatureSelection {
    readonly maxSelected = 20;
    private readonly selected = new Set<string>();

    get ids(): string[] {
        return Array.from(this.selected);
    }

    has(deviceID: string): boolean {
        return this.selected.has(deviceID);
    }

    toggle(deviceID: string): void {
        if (this.selected.delete(deviceID)) {
            return;
        }
        if (this.selected.size < this.maxSelected) {
            this.selected.add(deviceID);
        }
    }
}
