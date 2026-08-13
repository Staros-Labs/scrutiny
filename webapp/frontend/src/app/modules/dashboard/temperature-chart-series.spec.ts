import { alignTemperatureChartSeries } from './temperature-chart-series';

describe('alignTemperatureChartSeries', () => {
    it('aligns irregular series to one timestamp index without moving readings', () => {
        const midnight = new Date('2026-08-01T00:00:00Z');
        const oneAm = new Date('2026-08-01T01:00:00Z');
        const twoAm = new Date('2026-08-01T02:00:00Z');

        const result = alignTemperatureChartSeries([
            {
                name: 'Steady drive',
                data: [
                    { x: midnight, y: 40 },
                    { x: oneAm, y: 45 },
                    { x: twoAm, y: 40 },
                ],
            },
            {
                name: 'Spike drive',
                data: [
                    { x: midnight, y: 35 },
                    { x: twoAm, y: 70 },
                ],
            },
        ]);

        expect(result.map((series) => series.data.map((point) => point.x.toISOString()))).toEqual([
            [midnight, oneAm, twoAm].map((date) => date.toISOString()),
            [midnight, oneAm, twoAm].map((date) => date.toISOString()),
        ]);
        expect(result[0].data.map((point) => point.y)).toEqual([40, 45, 40]);
        expect(result[1].data.map((point) => point.y)).toEqual([35, null, 70]);
    });

    it('sorts timestamps and preserves empty visible series', () => {
        const earlier = new Date('2026-08-01T00:00:00Z');
        const later = new Date('2026-08-01T01:00:00Z');

        const result = alignTemperatureChartSeries([
            {
                name: 'Out of order drive',
                data: [
                    { x: later, y: 50 },
                    { x: earlier, y: 40 },
                ],
            },
            { name: 'No readings drive', data: [] },
        ]);

        expect(result[0].data.map((point) => point.y)).toEqual([40, 50]);
        expect(result[1].data.map((point) => point.y)).toEqual([null, null]);
    });

    it('snaps nearby raw collection times to one shared hover index', () => {
        const timestamp = (minute: number) => new Date(Date.UTC(2026, 7, 1, 0, minute));

        const result = alignTemperatureChartSeries([
            {
                name: 'First collector',
                data: [
                    { x: timestamp(0), y: 40 },
                    { x: timestamp(10), y: 41 },
                    { x: timestamp(20), y: 42 },
                ],
            },
            {
                name: 'Second collector',
                data: [
                    { x: timestamp(1), y: 35 },
                    { x: timestamp(11), y: 36 },
                    { x: timestamp(21), y: 70 },
                ],
            },
        ]);

        expect(result[0].data.map((point) => point.x.toISOString())).toEqual([timestamp(0), timestamp(10), timestamp(20)].map((date) => date.toISOString()));
        expect(result[0].data.map((point) => point.y)).toEqual([40, 41, 42]);
        expect(result[1].data.map((point) => point.y)).toEqual([35, 36, 70]);
    });

    it('does not snap sparse readings that are hours apart', () => {
        const timestamp = (hour: number) => new Date(Date.UTC(2026, 7, 1, hour));

        const result = alignTemperatureChartSeries([
            {
                name: 'Sparse drive',
                data: [
                    { x: timestamp(0), y: 40 },
                    { x: timestamp(24), y: 42 },
                ],
            },
            {
                name: 'Other sparse drive',
                data: [{ x: timestamp(12), y: 50 }],
            },
        ]);

        expect(result[0].data.map((point) => point.y)).toEqual([40, null, 42]);
        expect(result[1].data.map((point) => point.y)).toEqual([null, 50, null]);
    });
});
