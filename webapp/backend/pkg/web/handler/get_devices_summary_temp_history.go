package handler

import (
	"net/http"
	"sort"

	"github.com/analogj/scrutiny/webapp/backend/pkg/database"
	"github.com/analogj/scrutiny/webapp/backend/pkg/models"
	"github.com/analogj/scrutiny/webapp/backend/pkg/models/measurements"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetDevicesSummaryTempHistory(c *gin.Context) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	durationKey, exists := c.GetQuery("duration_key")
	if !exists {
		durationKey = "week"
	}

	deviceIDs := c.QueryArray("device_id")
	if len(deviceIDs) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "at most 20 device_id values are allowed"})
		return
	}
	var tempHistory map[string][]measurements.SmartTemperature
	var err error
	if len(deviceIDs) > 0 {
		tempHistory, err = deviceRepo.GetSmartTemperatureHistoryForDevices(c, durationKey, deviceIDs)
	} else {
		tempHistory, err = deviceRepo.GetSmartTemperatureHistory(c, durationKey)
	}
	if err != nil {
		logger.Errorln("An error occurred while retrieving summary/temp history", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"temp_history": tempHistory,
		},
	})
}

func GetTemperatureDeviceOptions(c *gin.Context) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	devices, err := deviceRepo.GetDevices(c)
	if err != nil {
		logger.Errorln("An error occurred while retrieving temperature device options", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	options := make([]models.TemperatureDeviceOption, 0, len(devices))
	for i := range devices {
		if devices[i].Archived {
			continue
		}
		options = append(options, models.TemperatureDeviceOption{
			DeviceID:     devices[i].DeviceID,
			HostID:       devices[i].HostId,
			Label:        devices[i].Label,
			DeviceName:   devices[i].DeviceName,
			ModelName:    devices[i].ModelName,
			SerialNumber: devices[i].SerialNumber,
		})
	}
	sort.Slice(options, func(i, j int) bool {
		if options[i].HostID != options[j].HostID {
			return options[i].HostID < options[j].HostID
		}
		return options[i].DeviceID < options[j].DeviceID
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"devices": options,
		},
	})
}
