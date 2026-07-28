import { SmartTemperatureModel } from './measurements/smart-temperature-model';

export interface DeviceSummaryTempResponseWrapper {
    success: boolean;
    errors: any[];
    data: {
        temp_history: { [key: string]: SmartTemperatureModel[] };
    };
}

export interface TemperatureDeviceOption {
    device_id: string;
    host_id: string;
    label: string;
    device_name: string;
    model_name: string;
    serial_number: string;
}

export interface TemperatureDeviceOptionsResponseWrapper {
    success: boolean;
    errors?: any[];
    data: {
        devices: TemperatureDeviceOption[];
    };
}
