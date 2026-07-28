import { TemperatureSelection } from './temperature-selection';

describe('TemperatureSelection', () => {
    it('starts empty, caps selection at twenty, and preserves existing selections', () => {
        const selection = new TemperatureSelection();

        for (let index = 0; index < 21; index++) {
            selection.toggle(`device-${index}`);
        }

        expect(selection.ids.length).toBe(20);
        expect(selection.ids).toContain('device-0');
        expect(selection.ids).not.toContain('device-20');
    });

    it('removes an already selected device', () => {
        const selection = new TemperatureSelection();
        selection.toggle('device-1');

        selection.toggle('device-1');

        expect(selection.ids).toEqual([]);
    });
});
