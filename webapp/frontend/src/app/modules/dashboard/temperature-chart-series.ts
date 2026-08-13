export interface TemperatureChartPoint {
    x: Date;
    y: number | null;
}

export interface TemperatureChartSeries {
    name: string;
    data: TemperatureChartPoint[];
}

const MAX_SNAP_TOLERANCE_MS = 15 * 60 * 1000;

interface IndexedTemperaturePoint {
    seriesIndex: number;
    timestamp: number;
    value: number | null;
}

interface TemperatureTimestampCluster {
    timestamps: number[];
    valuesBySeries: Map<number, number | null>;
}

function median(values: number[]): number {
    const sorted = [...values].sort((left, right) => left - right);
    return sorted[Math.floor((sorted.length - 1) / 2)];
}

function snapTolerance(series: TemperatureChartSeries[]): number {
    const intervals: number[] = [];

    for (const entry of series) {
        const timestamps = entry.data
            .map((point) => point.x.getTime())
            .filter(Number.isFinite)
            .sort((left, right) => left - right);

        for (let index = 1; index < timestamps.length; index++) {
            const interval = timestamps[index] - timestamps[index - 1];
            if (interval > 0) {
                intervals.push(interval);
            }
        }
    }

    return intervals.length > 0 ? Math.min(median(intervals) / 2, MAX_SNAP_TOLERANCE_MS) : 0;
}

/**
 * ApexCharts shared datetime tooltips resolve every series with one data-point index.
 * Irregular arrays can therefore show a value and marker from the wrong timestamp.
 * Keep series indexes aligned by clustering nearby collection times, padding absent
 * readings with null, and limiting nearest-X snapping to 15 minutes. Do not restore
 * the old tooltipUtil monkey patch: it claimed overlap without aligning the indexes.
 */
export function alignTemperatureChartSeries(series: TemperatureChartSeries[]): TemperatureChartSeries[] {
    const points: IndexedTemperaturePoint[] = [];

    for (let seriesIndex = 0; seriesIndex < series.length; seriesIndex++) {
        const entry = series[seriesIndex];
        for (const point of entry.data) {
            const timestamp = point.x.getTime();
            if (Number.isFinite(timestamp)) {
                points.push({ seriesIndex, timestamp, value: point.y });
            }
        }
    }

    points.sort((left, right) => left.timestamp - right.timestamp || left.seriesIndex - right.seriesIndex);

    const tolerance = snapTolerance(series);
    const clusters: TemperatureTimestampCluster[] = [];

    for (const point of points) {
        const cluster = clusters[clusters.length - 1];
        const clusterCenter = cluster ? cluster.timestamps.reduce((sum, timestamp) => sum + timestamp, 0) / cluster.timestamps.length : 0;
        const canJoinCluster = cluster && !cluster.valuesBySeries.has(point.seriesIndex) && Math.abs(point.timestamp - clusterCenter) <= tolerance;

        if (canJoinCluster) {
            cluster.timestamps.push(point.timestamp);
            cluster.valuesBySeries.set(point.seriesIndex, point.value);
        } else {
            clusters.push({
                timestamps: [point.timestamp],
                valuesBySeries: new Map([[point.seriesIndex, point.value]]),
            });
        }
    }

    return series.map((entry, seriesIndex) => {
        return {
            name: entry.name,
            data: clusters.map((cluster) => ({
                x: new Date(median(cluster.timestamps)),
                y: cluster.valuesBySeries.get(seriesIndex) ?? null,
            })),
        };
    });
}
