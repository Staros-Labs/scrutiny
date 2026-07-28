export interface HostSummaryModel {
    host_id: string;
    active_devices: number;
    archived_devices: number;
    total_devices: number;
}

export interface HostActionResultModel {
    host_id: string;
    success: boolean;
    device_count: number;
    error?: string;
}

export interface HostSummaryResponse {
    success: boolean;
    data: HostSummaryModel[];
}

export interface HostActionResponse {
    success: boolean;
    data: HostActionResultModel[];
    error?: string;
}
