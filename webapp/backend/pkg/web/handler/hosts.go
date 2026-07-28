package handler

import (
	"net/http"
	"strings"

	"github.com/analogj/scrutiny/webapp/backend/pkg/database"
	"github.com/analogj/scrutiny/webapp/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type hostActionRequest struct {
	HostIDs      []string `json:"host_ids"`
	Confirmation string   `json:"confirmation,omitempty"`
}

func GetHosts(c *gin.Context) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	hosts, err := deviceRepo.GetHosts(c)
	if err != nil {
		logger.Errorln("An error occurred while listing SMART hosts", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": hosts})
}

func ArchiveHosts(c *gin.Context) {
	updateHostsArchived(c, true)
}

func UnarchiveHosts(c *gin.Context) {
	updateHostsArchived(c, false)
}

func updateHostsArchived(c *gin.Context, archived bool) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	hostIDs, ok := bindHostIDs(c)
	if !ok {
		return
	}
	devices, err := deviceRepo.GetDevices(c)
	if err != nil {
		logger.Errorln("An error occurred while loading host devices", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	results := make([]models.HostActionResult, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		count, updateErr := deviceRepo.UpdateHostArchived(c, hostID, archived)
		result := models.HostActionResult{HostID: hostID, DeviceCount: count}
		switch {
		case updateErr != nil:
			result.Error = updateErr.Error()
		case count == 0:
			result.Error = "host not found"
		default:
			result.Success = true
			for i := range devices {
				if devices[i].HostId != hostID {
					continue
				}
				if archived {
					removeMqttDevice(c, &devices[i])
				} else {
					devices[i].Archived = false
					publishMqttDeviceDiscovery(c, &devices[i])
				}
			}
		}
		results = append(results, result)
	}
	c.JSON(http.StatusOK, gin.H{"success": allHostActionsSucceeded(results), "data": results})
}

func PurgeHosts(c *gin.Context) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	var request hostActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	hostIDs, err := normalizeHostIDs(request.HostIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if request.Confirmation != "PURGE" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "confirmation must equal PURGE"})
		return
	}

	devices, err := deviceRepo.GetDevices(c)
	if err != nil {
		logger.Errorln("An error occurred while loading host devices for purge", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	results, err := deviceRepo.PurgeHosts(c, hostIDs)
	if err != nil {
		logger.Errorln("An error occurred while purging SMART hosts", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	for _, result := range results {
		if !result.Success {
			continue
		}
		for i := range devices {
			if devices[i].HostId == result.HostID {
				removeMqttDevice(c, &devices[i])
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": allHostActionsSucceeded(results), "data": results})
}

func bindHostIDs(c *gin.Context) ([]string, bool) {
	var request hostActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return nil, false
	}
	hostIDs, err := normalizeHostIDs(request.HostIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return nil, false
	}
	return hostIDs, true
}

func normalizeHostIDs(input []string) ([]string, error) {
	hostIDs := make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, rawHostID := range input {
		if strings.TrimSpace(rawHostID) == "" {
			return nil, &hostRequestError{message: "host_ids cannot contain empty values"}
		}
		hostID := rawHostID
		if _, exists := seen[hostID]; exists {
			continue
		}
		seen[hostID] = struct{}{}
		hostIDs = append(hostIDs, hostID)
	}
	if len(hostIDs) == 0 {
		return nil, &hostRequestError{message: "host_ids must contain at least one host"}
	}
	return hostIDs, nil
}

type hostRequestError struct {
	message string
}

func (e *hostRequestError) Error() string {
	return e.message
}

func allHostActionsSucceeded(results []models.HostActionResult) bool {
	if len(results) == 0 {
		return false
	}
	for _, result := range results {
		if !result.Success {
			return false
		}
	}
	return true
}
