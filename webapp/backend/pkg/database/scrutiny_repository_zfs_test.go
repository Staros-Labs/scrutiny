package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/analogj/scrutiny/webapp/backend/pkg/models"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createZFSTestRepository(t *testing.T) *scrutinyRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.ZFSPool{}, &models.ZFSVdev{}))

	return &scrutinyRepository{gormClient: db}
}

func TestZFSPoolSummariesKeepArchivedPoolsOutOfMonitoringButAvailableToUI(t *testing.T) {
	repo := createZFSTestRepository(t)
	ctx := context.Background()

	activePool := models.ZFSPool{GUID: "1", Name: "active"}
	archivedPool := models.ZFSPool{GUID: "2", Name: "archived"}
	require.NoError(t, repo.RegisterZFSPool(ctx, activePool))
	require.NoError(t, repo.RegisterZFSPool(ctx, archivedPool))
	require.NoError(t, repo.UpdateZFSPoolArchived(ctx, archivedPool.GUID, true))

	monitoringSummary, err := repo.GetZFSPoolsSummary(ctx)
	require.NoError(t, err)
	require.Contains(t, monitoringSummary, activePool.GUID)
	require.NotContains(t, monitoringSummary, archivedPool.GUID)

	uiSummary, err := repo.GetAllZFSPoolsSummary(ctx)
	require.NoError(t, err)
	require.Contains(t, uiSummary, activePool.GUID)
	require.Contains(t, uiSummary, archivedPool.GUID)
	require.True(t, uiSummary[archivedPool.GUID].Archived)
}

func TestRegisterZFSPoolPreservesArchivedState(t *testing.T) {
	repo := createZFSTestRepository(t)
	ctx := context.Background()
	pool := models.ZFSPool{GUID: "1", Name: "tank"}

	require.NoError(t, repo.RegisterZFSPool(ctx, pool))
	require.NoError(t, repo.UpdateZFSPoolArchived(ctx, pool.GUID, true))
	pool.Name = "renamed"
	require.NoError(t, repo.RegisterZFSPool(ctx, pool))

	summary, err := repo.GetAllZFSPoolsSummary(ctx)
	require.NoError(t, err)
	require.True(t, summary[pool.GUID].Archived)
	require.Equal(t, "renamed", summary[pool.GUID].Name)
}
