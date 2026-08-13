package handler

import (
	"github.com/analogj/scrutiny/webapp/backend/pkg/database"
	"github.com/analogj/scrutiny/webapp/backend/pkg/models"
	"github.com/analogj/scrutiny/webapp/backend/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func SaveSettings(c *gin.Context) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	var settings models.Settings
	err := c.BindJSON(&settings)
	if err != nil {
		logger.Errorln("Cannot parse updated settings", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	settings.ApplyDefaults()
	if !validSummaryPageSize(settings.DashboardPageSize) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "dashboard_page_size must be one of 25, 50, 100, or 250"})
		return
	}
	if !validSummaryHostPageSize(settings.DashboardHostPageSize) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "dashboard_host_page_size must be one of 5, 10, 25, or 50"})
		return
	}

	err = deviceRepo.SaveSettings(c, settings)
	if err != nil {
		logger.Errorln("An error occurred while saving settings", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":                        true,
		"settings":                       settings,
		"server_version":                 version.VERSION,
		"collector_trigger_enabled":      collectorTriggerEnabled(),
		"zfs_pool_modifications_allowed": zfsPoolModificationsAllowed(c),
	})
}
