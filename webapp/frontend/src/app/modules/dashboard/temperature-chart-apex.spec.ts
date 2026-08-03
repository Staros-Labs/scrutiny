import ApexCharts from 'apexcharts';
import { alignTemperatureChartSeries } from './temperature-chart-series';

describe('dashboard temperature ApexCharts integration', () => {
    let chart: ApexCharts;
    let host: HTMLDivElement;

    beforeEach(() => {
        host = document.createElement('div');
        host.style.width = '700px';
        host.style.height = '320px';
        document.body.appendChild(host);
    });

    afterEach(() => {
        chart?.destroy();
        host.remove();
    });

    it('shows aligned values and markers without moving the fixed tooltip', async () => {
        const timestamp = (hour: number) => new Date(Date.UTC(2026, 7, 1, hour));
        const series = alignTemperatureChartSeries([
            {
                name: 'Steady drive',
                data: [
                    { x: timestamp(0), y: 40 },
                    { x: timestamp(1), y: 45 },
                    { x: timestamp(2), y: 40 },
                ],
            },
            {
                name: 'Spike drive',
                data: [
                    { x: timestamp(0), y: 35 },
                    { x: timestamp(2), y: 70 },
                ],
            },
        ]);

        chart = new ApexCharts(host, {
            chart: {
                type: 'area',
                width: 700,
                height: 320,
                animations: { enabled: false },
                toolbar: { show: false },
            },
            series,
            markers: {
                size: 0,
                hover: { sizeOffset: 4 },
            },
            tooltip: {
                shared: true,
                intersect: false,
                fixed: {
                    enabled: true,
                    position: 'topLeft',
                    offsetX: 12,
                    offsetY: 12,
                },
            },
            xaxis: { type: 'datetime' },
        });
        await chart.render();

        await hoverPoint(chart, host, 0, 0);
        const tooltip = host.querySelector<HTMLElement>('.apexcharts-tooltip');
        const initialPosition = { left: tooltip?.style.left, top: tooltip?.style.top };

        await hoverPoint(chart, host, 1, 2);

        const activeRows = Array.from(host.querySelectorAll<HTMLElement>('.apexcharts-tooltip-series-group.apexcharts-active'));
        expect(activeRows.map((row) => row.textContent)).toEqual([jasmine.stringMatching(/Steady drive:\s*40/), jasmine.stringMatching(/Spike drive:\s*70/)]);
        expect(dynamicMarkerPaths(host).every((path) => path.getAttribute('d'))).toBeTrue();
        expect({ left: tooltip?.style.left, top: tooltip?.style.top }).toEqual(initialPosition);
    });
});

async function hoverPoint(chart: ApexCharts, host: HTMLElement, seriesIndex: number, dataPointIndex: number): Promise<void> {
    await new Promise<void>((resolve) => setTimeout(resolve, 25));

    const chartContext = chart as any;
    const [x, y] = chartContext.w.globals.pointsArray[seriesIndex][dataPointIndex];
    const gridBounds = host.querySelector<SVGElement>('.apexcharts-grid')?.getBoundingClientRect();
    const svg = host.querySelector<SVGElement>('.apexcharts-svg');

    svg?.dispatchEvent(
        new MouseEvent('mousemove', {
            bubbles: true,
            clientX: (gridBounds?.left || 0) + x,
            clientY: (gridBounds?.top || 0) + y,
        })
    );

    await new Promise<void>((resolve) => setTimeout(resolve, 25));
}

function dynamicMarkerPaths(host: HTMLElement): SVGPathElement[] {
    return Array.from(host.querySelectorAll<SVGPathElement>('.apexcharts-series-markers-wrap > .apexcharts-series-markers:last-child .apexcharts-marker'));
}
