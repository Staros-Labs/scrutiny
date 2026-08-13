package database

import (
	"fmt"
	"sort"
	"testing"

	"github.com/analogj/scrutiny/webapp/backend/pkg"
	"github.com/analogj/scrutiny/webapp/backend/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestCompareDeviceSummariesKeepsHostsTogetherBeforeDeviceSort(t *testing.T) {
	summaries := []*models.DeviceSummary{
		{Device: models.Device{DeviceID: "b-2", HostId: "beta", DeviceName: "sdb"}},
		{Device: models.Device{DeviceID: "a-2", HostId: "alpha", DeviceName: "sdb"}},
		{Device: models.Device{DeviceID: "a-1", HostId: "alpha", DeviceName: "sda"}},
	}

	sort.SliceStable(summaries, func(i, j int) bool {
		return compareDeviceSummaries(summaries[i], summaries[j], "title_asc", "name") < 0
	})

	require.Equal(t, []string{"a-1", "a-2", "b-2"}, []string{
		summaries[0].Device.DeviceID,
		summaries[1].Device.DeviceID,
		summaries[2].Device.DeviceID,
	})
}

func TestCompareDeviceSummariesPreservesStatusSortSemantics(t *testing.T) {
	passed := &models.DeviceSummary{
		Device:       models.Device{DeviceID: "passed"},
		SmartResults: &models.SmartSummary{},
	}
	failed := &models.DeviceSummary{
		Device:       models.Device{DeviceID: "failed", DeviceStatus: pkg.DeviceStatusFailedScrutiny},
		SmartResults: &models.SmartSummary{},
	}

	require.Greater(t, compareDeviceSummaries(failed, passed, "status_desc", "name"), 0)
	require.Less(t, compareDeviceSummaries(failed, passed, "status_asc", "name"), 0)
}

func TestDeviceTitleForTypeFallsBackToDeviceName(t *testing.T) {
	device := models.Device{DeviceName: "sda", DeviceType: "sat", ModelName: "Example"}

	require.Equal(t, "/dev/sda - sat - Example", deviceTitleForType(&device, "label"))
}

func TestBuildSummaryPageFiltersSlicesAndCountsAttention(t *testing.T) {
	summaries := make(map[string]*models.DeviceSummary)
	for index := 1; index <= 60; index++ {
		deviceID := fmt.Sprintf("device-%03d", index)
		summaries[deviceID] = &models.DeviceSummary{
			Device: models.Device{
				DeviceID:     deviceID,
				DeviceName:   fmt.Sprintf("sd%03d", index),
				DeviceStatus: pkg.DeviceStatusPassed,
			},
		}
	}
	summaries["device-001"].Device.DeviceStatus = pkg.DeviceStatusFailedSmart
	summaries["archived"] = &models.DeviceSummary{
		Device: models.Device{DeviceID: "archived", DeviceName: "archived", Archived: true},
	}

	page := buildSummaryPage(summaries, models.DeviceSummaryPageOptions{
		Page:     2,
		PageSize: 25,
		Sort:     "title_asc",
		Display:  "name",
	})

	require.Len(t, page.Summary, 25)
	require.Equal(t, 60, page.Pagination.TotalItems)
	require.Equal(t, 3, page.Pagination.TotalPages)
	require.Equal(t, 1, page.Pagination.AttentionCount)
	require.NotContains(t, page.Summary, "archived")
	require.Contains(t, page.Summary, "device-026")
	require.Contains(t, page.Summary, "device-050")
}

func TestBuildSummaryPageFiltersHostsBeforePagination(t *testing.T) {
	summaries := map[string]*models.DeviceSummary{
		"alpha-1": {Device: models.Device{DeviceID: "alpha-1", HostId: "Alpha"}},
		"alpha-2": {Device: models.Device{DeviceID: "alpha-2", HostId: "Alpha"}},
		"beta-1":  {Device: models.Device{DeviceID: "beta-1", HostId: "Beta"}},
	}

	page := buildSummaryPage(summaries, models.DeviceSummaryPageOptions{
		Page:       1,
		PageSize:   25,
		HostSearch: " alp ",
	})

	require.Equal(t, 2, page.Pagination.TotalItems)
	require.Contains(t, page.Summary, "alpha-1")
	require.Contains(t, page.Summary, "alpha-2")
	require.NotContains(t, page.Summary, "beta-1")
}

func TestBuildSummaryPagePaginatesHostsWithoutSplittingDevices(t *testing.T) {
	summaries := map[string]*models.DeviceSummary{
		"host-1-a": {Device: models.Device{DeviceID: "host-1-a", HostId: "host-1"}},
		"host-1-b": {Device: models.Device{DeviceID: "host-1-b", HostId: "host-1"}},
		"host-2":   {Device: models.Device{DeviceID: "host-2", HostId: "host-2"}},
		"host-3":   {Device: models.Device{DeviceID: "host-3", HostId: "host-3"}},
		"host-4":   {Device: models.Device{DeviceID: "host-4", HostId: "host-4"}},
		"host-5":   {Device: models.Device{DeviceID: "host-5", HostId: "host-5"}},
		"host-6":   {Device: models.Device{DeviceID: "host-6", HostId: "host-6"}},
	}

	firstPage := buildSummaryPage(summaries, models.DeviceSummaryPageOptions{
		Page:        1,
		PageSize:    5,
		GroupByHost: true,
	})

	require.Len(t, firstPage.Summary, 6)
	require.Equal(t, 6, firstPage.Pagination.TotalItems)
	require.Equal(t, 2, firstPage.Pagination.TotalPages)
	require.Contains(t, firstPage.Summary, "host-1-a")
	require.Contains(t, firstPage.Summary, "host-1-b")
	require.NotContains(t, firstPage.Summary, "host-6")

	secondPage := buildSummaryPage(summaries, models.DeviceSummaryPageOptions{
		Page:        2,
		PageSize:    5,
		GroupByHost: true,
	})
	require.Equal(t, map[string]*models.DeviceSummary{"host-6": summaries["host-6"]}, secondPage.Summary)
}
