import { DeviceSummaryModel } from 'app/core/models/device-summary-model';

export interface DeviceSummaryPagination {
    page: number;
    page_size: number;
    total_items: number;
    total_pages: number;
    attention_count: number;
}

export interface DeviceSummaryPage {
    summary: { [key: string]: DeviceSummaryModel };
    pagination: DeviceSummaryPagination;
}

// maps to webapp/backend/pkg/models/device_summary.go
export interface DeviceSummaryResponseWrapper {
    success: boolean;
    errors: any[];
    data: {
        summary: { [key: string]: DeviceSummaryModel };
        pagination?: DeviceSummaryPagination;
    };
}
