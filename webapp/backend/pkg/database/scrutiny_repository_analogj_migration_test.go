package database

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/analogj/scrutiny/webapp/backend/pkg"
	mock_config "github.com/analogj/scrutiny/webapp/backend/pkg/config/mock"
	"github.com/analogj/scrutiny/webapp/backend/pkg/deviceid"
	"github.com/analogj/scrutiny/webapp/backend/pkg/models"
	"github.com/glebarez/sqlite"
	"github.com/golang/mock/gomock"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAnalogJInfluxMigrationQueryRewritesOnlyLegacySmartData(t *testing.T) {
	query := analogJInfluxMigrationQuery(
		"metrics_weekly",
		"scrutiny",
		"9027cd65-fe19-562d-ac5d-0bb05818087f",
		"nvme123",
	)

	require.Contains(t, query, `from(bucket: "metrics_weekly")`)
	require.Contains(t, query, `r["_measurement"] == "smart" or r["_measurement"] == "temp"`)
	require.Contains(t, query, `exists r["scrutiny_uuid"] and r["scrutiny_uuid"] == "9027cd65-fe19-562d-ac5d-0bb05818087f"`)
	require.Contains(t, query, `drop(columns: ["scrutiny_uuid"])`)
	require.Contains(t, query, `set(key: "device_wwn", value: "nvme123")`)
	require.Contains(t, query, `to(bucket: "metrics_weekly", org: "scrutiny")`)
	require.Equal(t, 1, strings.Count(query, `to(bucket:`))
}

func TestLegacyScrutinyUUIDMatchesAnalogJIdentity(t *testing.T) {
	require.Equal(t,
		"9027cd65-fe19-562d-ac5d-0bb05818087f",
		legacyScrutinyUUID("Samsung SSD 980 PRO 2TB", "NVME123", ""),
	)
	require.Equal(t,
		"7a14ef35-ed3a-552a-9980-46291bd986e3",
		legacyScrutinyUUID("WDC WD80EFZZ-68BTXN0", "WD-CA2XZ08L", "0x50014ee2c06ce3c3"),
	)
}

func TestPrepareAnalogJDeviceRepairRestoresNVMeIdentityAndMergesDuplicate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Device{}))

	legacyCreatedAt := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	model := "Samsung SSD 980 PRO 2TB"
	serial := "NVME123"
	legacyID := deviceid.Generate(model, serial, "")
	canonicalWWN := "nvme123"
	canonicalID := deviceid.Generate(model, serial, canonicalWWN)

	require.NoError(t, db.Create(&models.Device{
		DeviceID:                  legacyID,
		WWN:                       "",
		DeviceName:                "nvme0",
		ModelName:                 model,
		SerialNumber:              serial,
		DeviceProtocol:            pkg.DeviceProtocolNvme,
		Label:                     "Archive Bay",
		Archived:                  true,
		Muted:                     true,
		MissedPingTimeoutOverride: 90,
		SmartDisplayMode:          "raw",
		CreatedAt:                 legacyCreatedAt,
	}).Error)
	require.NoError(t, db.Create(&models.Device{
		DeviceID:       canonicalID,
		WWN:            canonicalWWN,
		DeviceName:     "nvme0",
		ModelName:      model,
		SerialNumber:   serial,
		DeviceProtocol: pkg.DeviceProtocolNvme,
		DeviceStatus:   pkg.DeviceStatusFailedScrutiny,
	}).Error)

	sataModel := "WDC WD80EFZZ-68BTXN0"
	sataSerial := "WD-CA2XZ08L"
	sataWWN := "0x50014ee2c06ce3c3"
	require.NoError(t, db.Create(&models.Device{
		DeviceID:     deviceid.Generate(sataModel, sataSerial, sataWWN),
		WWN:          sataWWN,
		DeviceName:   "sda",
		ModelName:    sataModel,
		SerialNumber: sataSerial,
	}).Error)

	mapping, err := prepareAnalogJDeviceRepair(context.Background(), db)
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"9027cd65-fe19-562d-ac5d-0bb05818087f": canonicalWWN,
		"7a14ef35-ed3a-552a-9980-46291bd986e3": sataWWN,
	}, mapping)

	var devices []models.Device
	require.NoError(t, db.Order("device_name ASC").Find(&devices).Error)
	require.Len(t, devices, 2)

	var repaired models.Device
	require.NoError(t, db.Where(queryDeviceID, canonicalID).First(&repaired).Error)
	require.Equal(t, canonicalWWN, repaired.WWN)
	require.Equal(t, "Archive Bay", repaired.Label)
	require.True(t, repaired.Archived)
	require.True(t, repaired.Muted)
	require.Equal(t, 90, repaired.MissedPingTimeoutOverride)
	require.Equal(t, "raw", repaired.SmartDisplayMode)
	require.Equal(t, pkg.DeviceStatusFailedScrutiny, repaired.DeviceStatus)
	require.Equal(t, legacyCreatedAt, repaired.CreatedAt.UTC())

	secondMapping, err := prepareAnalogJDeviceRepair(context.Background(), db)
	require.NoError(t, err)
	require.Equal(t, mapping, secondMapping)
	require.NoError(t, db.Find(&devices).Error)
	require.Len(t, devices, 2)
}

func TestPrepareAnalogJDeviceRepairUsesCollectorDiscoveredWWN(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Device{}))

	model := "Micron NVMe"
	canonicalModel := model + " Gen4"
	serial := "MTFD123"
	legacyID := deviceid.Generate(model, serial, "")
	canonicalWWN := "eui.0025385b11aa2233"
	canonicalID := deviceid.Generate(canonicalModel, serial, canonicalWWN)
	deletedAt := time.Now().UTC()
	require.NoError(t, db.Create(&models.Device{
		DeviceID:     legacyID,
		ModelName:    model,
		SerialNumber: serial,
		HostId:       "nas-1",
		Label:        "Cache",
	}).Error)
	require.NoError(t, db.Create(&models.Device{
		DeviceID:     canonicalID,
		WWN:          canonicalWWN,
		ModelName:    canonicalModel,
		SerialNumber: serial,
		HostId:       "nas-1",
		DeletedAt:    &deletedAt,
	}).Error)

	mapping, err := prepareAnalogJDeviceRepair(context.Background(), db)
	require.NoError(t, err)
	require.Equal(t, canonicalWWN, mapping[legacyScrutinyUUID(model, serial, "")])

	var devices []models.Device
	require.NoError(t, db.Find(&devices).Error)
	require.Len(t, devices, 1)
	require.Equal(t, canonicalID, devices[0].DeviceID)
	require.Equal(t, "Cache", devices[0].Label)
}

func TestAnalogJMigrationIntegration(t *testing.T) {
	ResetMigrationGuardForTests()

	influxHost := "localhost"
	if _, isGitHubActions := os.LookupEnv("GITHUB_ACTIONS"); isGitHubActions {
		influxHost = "influxdb"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(fmt.Sprintf("http://%s:8086/api/v2/setup", influxHost))
	if err != nil {
		t.Skip("Skipping integration test: InfluxDB not available at " + influxHost + ":8086")
	}
	require.NoError(t, response.Body.Close())

	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	fakeConfig := mock_config.NewMockInterface(mockCtrl)
	fakeConfig.EXPECT().GetString("web.database.location").Return(filepath.Join(t.TempDir(), "scrutiny.db")).AnyTimes()
	fakeConfig.EXPECT().GetString("web.database.journal_mode").Return("WAL").AnyTimes()
	fakeConfig.EXPECT().GetString("log.level").Return("INFO").AnyTimes()
	fakeConfig.EXPECT().GetString("web.influxdb.scheme").Return("http").AnyTimes()
	fakeConfig.EXPECT().GetString("web.influxdb.host").Return(influxHost).AnyTimes()
	fakeConfig.EXPECT().GetString("web.influxdb.port").Return("8086").AnyTimes()
	fakeConfig.EXPECT().GetString("web.influxdb.token").Return("my-super-secret-auth-token").AnyTimes()
	fakeConfig.EXPECT().GetString("web.influxdb.org").Return("scrutiny").AnyTimes()
	fakeConfig.EXPECT().GetString("web.influxdb.bucket").Return("metrics").AnyTimes()
	fakeConfig.EXPECT().GetString("web.influxdb.init_username").Return("admin").AnyTimes()
	fakeConfig.EXPECT().GetString("web.influxdb.init_password").Return("password12345").AnyTimes()
	fakeConfig.EXPECT().GetBool("web.influxdb.tls.insecure_skip_verify").Return(false).AnyTimes()
	fakeConfig.EXPECT().GetBool("web.influxdb.retention_policy").Return(false).AnyTimes()
	fakeConfig.EXPECT().GetInt("web.influxdb.retention_policy.default_retention_period_days").Return(0).AnyTimes()
	fakeConfig.EXPECT().GetInt("web.influxdb.retention_policy.monthly_retention_period_months").Return(0).AnyTimes()
	fakeConfig.EXPECT().GetInt("web.influxdb.retention_policy.yearly_retention_period_months").Return(0).AnyTimes()
	fakeConfig.EXPECT().GetIntSlice("failures.transient.ata").Return([]int{195}).AnyTimes()
	fakeConfig.EXPECT().GetStringSlice("failures.ignored.devstat").Return([]string{}).AnyTimes()
	fakeConfig.EXPECT().Get("smart.attribute_overrides").Return(nil).AnyTimes()

	repoInterface, err := NewScrutinyRepository(fakeConfig, logrus.WithField("test", t.Name()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, repoInterface.Close()) })
	repo := repoInterface.(*scrutinyRepository)
	ctx := context.Background()

	model := "Issue 694 NVMe"
	serial := "ISSUE694NVME"
	legacyUUID := legacyScrutinyUUID(model, serial, "")
	targetWWN := strings.ToLower(serial)
	legacyDeviceID := deviceid.Generate(model, serial, "")
	currentDeviceID := deviceid.Generate(model, serial, targetWWN)
	require.NoError(t, repo.gormClient.Create(&models.Device{
		DeviceID:       legacyDeviceID,
		DeviceName:     "nvme694",
		ModelName:      model,
		SerialNumber:   serial,
		DeviceProtocol: pkg.DeviceProtocolNvme,
	}).Error)
	require.NoError(t, repo.gormClient.Exec(
		"INSERT OR IGNORE INTO migrations (id) VALUES (?)",
		analogJScrutinyUUIDMigrationID,
	).Error)

	buckets := []string{"metrics", "metrics_weekly", "metrics_monthly", "metrics_yearly"}
	pointTime := time.Now().UTC().Add(-2 * time.Hour)
	for _, bucket := range buckets {
		writeAPI := repo.influxClient.WriteAPIBlocking("scrutiny", bucket)
		tags := map[string]string{"scrutiny_uuid": legacyUUID, "device_protocol": pkg.DeviceProtocolNvme}
		require.NoError(t, writeAPI.WritePoint(ctx, influxdb2.NewPoint(
			"smart",
			tags,
			map[string]interface{}{"temp": int64(41), "power_on_hours": int64(694)},
			pointTime,
		)))
		require.NoError(t, writeAPI.WritePoint(ctx, influxdb2.NewPoint(
			"temp",
			tags,
			map[string]interface{}{"temp": int64(41)},
			pointTime,
		)))
	}

	deleteStart := time.Unix(0, 0).UTC()
	deleteStop := time.Now().UTC().Add(24 * time.Hour)
	t.Cleanup(func() {
		for _, bucket := range buckets {
			_ = repo.influxClient.DeleteAPI().DeleteWithName(ctx, "scrutiny", bucket, deleteStart, deleteStop, fmt.Sprintf("device_wwn=%q", targetWWN))
			_ = repo.influxClient.DeleteAPI().DeleteWithName(ctx, "scrutiny", bucket, deleteStart, deleteStop, fmt.Sprintf("scrutiny_uuid=%q", legacyUUID))
		}
	})

	require.NoError(t, repo.gormClient.Transaction(func(tx *gorm.DB) error {
		return repo.migrateM20260803000000(ctx, tx)
	}))

	var repaired models.Device
	require.NoError(t, repo.gormClient.Where(queryDeviceID, currentDeviceID).First(&repaired).Error)
	require.Equal(t, targetWWN, repaired.WWN)
	for _, bucket := range buckets {
		require.Positive(t, influxRecordCount(t, repo, bucket, "device_wwn", targetWWN))
		require.Zero(t, influxRecordCount(t, repo, bucket, "scrutiny_uuid", legacyUUID))
	}

	summary, err := repo.GetSummary(ctx)
	require.NoError(t, err)
	require.NotNil(t, summary[currentDeviceID].SmartResults)
	require.Equal(t, int64(694), summary[currentDeviceID].SmartResults.PowerOnHours)

	require.NoError(t, repo.gormClient.Transaction(func(tx *gorm.DB) error {
		return repo.migrateM20260803000000(ctx, tx)
	}))
	require.Positive(t, influxRecordCount(t, repo, "metrics", "device_wwn", targetWWN))
}

func influxRecordCount(t *testing.T, repo *scrutinyRepository, bucket, tag, value string) int {
	t.Helper()
	query := fmt.Sprintf(`from(bucket: %q)
|> range(start: 0)
|> filter(fn: (r) => r[%q] == %q)
|> count()`, bucket, tag, value)
	result, err := repo.influxQueryApi.Query(context.Background(), query)
	require.NoError(t, err)
	defer result.Close()
	count := 0
	for result.Next() {
		count++
	}
	require.NoError(t, result.Err())
	return count
}
