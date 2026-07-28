package models

type HostSummary struct {
	HostID          string `json:"host_id"`
	ActiveDevices   int64  `json:"active_devices"`
	ArchivedDevices int64  `json:"archived_devices"`
	TotalDevices    int64  `json:"total_devices"`
}

type HostActionResult struct {
	HostID      string `json:"host_id"`
	Success     bool   `json:"success"`
	DeviceCount int64  `json:"device_count"`
	Error       string `json:"error,omitempty"`
}
