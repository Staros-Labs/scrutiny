package database

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/analogj/scrutiny/webapp/backend/pkg/models"
	"gorm.io/gorm"
)

func (sr *scrutinyRepository) GetHosts(ctx context.Context) ([]models.HostSummary, error) {
	hosts := []models.HostSummary{}
	err := sr.gormClient.WithContext(ctx).
		Model(&models.Device{}).
		Select(`
			host_id,
			SUM(CASE WHEN archived = false THEN 1 ELSE 0 END) AS active_devices,
			SUM(CASE WHEN archived = true THEN 1 ELSE 0 END) AS archived_devices,
			COUNT(*) AS total_devices
		`).
		Where("TRIM(host_id) <> ''").
		Group("host_id").
		Order("LOWER(host_id) ASC, host_id ASC").
		Scan(&hosts).Error
	if err != nil {
		return nil, fmt.Errorf("could not list SMART hosts: %w", err)
	}
	return hosts, nil
}

func (sr *scrutinyRepository) UpdateHostArchived(ctx context.Context, hostID string, archived bool) (int64, error) {
	result := sr.gormClient.WithContext(ctx).
		Model(&models.Device{}).
		Where("host_id = ?", hostID).
		Update("archived", archived)
	if result.Error != nil {
		return 0, fmt.Errorf("could not update host %q: %w", hostID, result.Error)
	}
	return result.RowsAffected, nil
}

func (sr *scrutinyRepository) PurgeHosts(ctx context.Context, hostIDs []string) ([]models.HostActionResult, error) {
	var devices []models.Device
	if err := sr.gormClient.WithContext(ctx).
		Where("TRIM(host_id) <> ''").
		Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("could not load SMART hosts for purge: %w", err)
	}

	selectedHosts := make(map[string]struct{}, len(hostIDs))
	devicesByHost := make(map[string][]models.Device, len(hostIDs))
	for _, hostID := range hostIDs {
		selectedHosts[hostID] = struct{}{}
	}
	for i := range devices {
		if _, selected := selectedHosts[devices[i].HostId]; selected {
			devicesByHost[devices[i].HostId] = append(devicesByHost[devices[i].HostId], devices[i])
		}
	}

	blockedHosts := findHostsWithExternallySharedWWNs(devices, selectedHosts)
	results := make([]models.HostActionResult, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		hostDevices := devicesByHost[hostID]
		result := models.HostActionResult{
			HostID:      hostID,
			DeviceCount: int64(len(hostDevices)),
		}
		if len(hostDevices) == 0 {
			result.Error = "host not found"
			results = append(results, result)
			continue
		}
		if sharedWWNs := blockedHosts[hostID]; len(sharedWWNs) > 0 {
			result.Error = fmt.Sprintf(
				"history deletion blocked because WWN %q is also used outside selected hosts",
				sharedWWNs[0],
			)
			results = append(results, result)
			continue
		}

		if err := sr.deleteHostInfluxHistory(ctx, hostID, hostDevices); err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		if err := sr.gormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return tx.Where("host_id = ?", hostID).Delete(&models.Device{}).Error
		}); err != nil {
			result.Error = fmt.Sprintf("could not delete host devices: %v", err)
			results = append(results, result)
			continue
		}

		result.Success = true
		results = append(results, result)
	}
	return results, nil
}

func findHostsWithExternallySharedWWNs(devices []models.Device, selectedHosts map[string]struct{}) map[string][]string {
	hostsByWWN := make(map[string]map[string]struct{})
	for i := range devices {
		wwn := devices[i].WWN
		if strings.TrimSpace(wwn) == "" {
			continue
		}
		if hostsByWWN[wwn] == nil {
			hostsByWWN[wwn] = make(map[string]struct{})
		}
		hostsByWWN[wwn][devices[i].HostId] = struct{}{}
	}

	blocked := make(map[string][]string)
	for wwn, hosts := range hostsByWWN {
		hasOutsideHost := false
		for hostID := range hosts {
			if _, selected := selectedHosts[hostID]; !selected {
				hasOutsideHost = true
				break
			}
		}
		if !hasOutsideHost {
			continue
		}
		for hostID := range hosts {
			if _, selected := selectedHosts[hostID]; selected {
				blocked[hostID] = append(blocked[hostID], wwn)
			}
		}
	}
	for hostID := range blocked {
		sort.Strings(blocked[hostID])
	}
	return blocked
}

func (sr *scrutinyRepository) deleteHostInfluxHistory(ctx context.Context, hostID string, devices []models.Device) error {
	wwns := make(map[string]struct{}, len(devices))
	for i := range devices {
		if wwn := devices[i].WWN; strings.TrimSpace(wwn) != "" {
			wwns[wwn] = struct{}{}
		}
	}
	if len(wwns) == 0 {
		return nil
	}

	buckets := []string{
		sr.appConfig.GetString(cfgInfluxDBBucket),
		fmt.Sprintf("%s_weekly", sr.appConfig.GetString(cfgInfluxDBBucket)),
		fmt.Sprintf("%s_monthly", sr.appConfig.GetString(cfgInfluxDBBucket)),
		fmt.Sprintf("%s_yearly", sr.appConfig.GetString(cfgInfluxDBBucket)),
	}
	for wwn := range wwns {
		for _, bucket := range buckets {
			sr.logger.Infof("Purging SMART host %s history for WWN %s in bucket %s", hostID, wwn, bucket)
			if err := sr.influxClient.DeleteAPI().DeleteWithName(
				ctx,
				sr.appConfig.GetString(cfgInfluxDBOrg),
				bucket,
				time.Now().AddDate(-10, 0, 0),
				time.Now(),
				fmt.Sprintf("device_wwn=%q", wwn),
			); err != nil {
				return fmt.Errorf("could not delete history for WWN %q from bucket %q: %w", wwn, bucket, err)
			}
		}
	}
	return nil
}
