package database

import (
	"context"
	"th-application-technical-assignment/sqlc"

	"github.com/stretchr/testify/mock"
    "github.com/google/uuid"
)

type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreateEpisode(ctx context.Context, params sqlc.CreateEpisodeParams) (sqlc.Episode, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Episode), args.Error(1)
}

func (m *MockQuerier) GetEpisode(ctx context.Context, id uuid.UUID) (sqlc.Episode, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Episode), args.Error(1)
}

func (m *MockQuerier) UpdateEpisode(ctx context.Context, params sqlc.UpdateEpisodeParams) (sqlc.Episode, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Episode), args.Error(1)
}

func (m *MockQuerier) DeleteEpisode(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) CountEpisodesBySeries(ctx context.Context, seriesID uuid.UUID) (int64, error) {
	args := m.Called(ctx, seriesID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) ListEpisodesBySeries(ctx context.Context, seriesID uuid.UUID) ([]sqlc.Episode, error) {
	args := m.Called(ctx, seriesID)
	return args.Get(0).([]sqlc.Episode), args.Error(1)
}

func (m *MockQuerier) ListEpisodesBySeriesPaginated(ctx context.Context, params sqlc.ListEpisodesBySeriesPaginatedParams) ([]sqlc.Episode, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]sqlc.Episode), args.Error(1)
}

func (m *MockQuerier) GetEpisodeWithAssets(ctx context.Context, id uuid.UUID) ([]sqlc.GetEpisodeWithAssetsRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]sqlc.GetEpisodeWithAssetsRow), args.Error(1)
}

func (m *MockQuerier) ListEpisodesWithAssetsBySeriesPaginated(ctx context.Context, params sqlc.ListEpisodesWithAssetsBySeriesPaginatedParams) ([]sqlc.ListEpisodesWithAssetsBySeriesPaginatedRow, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]sqlc.ListEpisodesWithAssetsBySeriesPaginatedRow), args.Error(1)
}

// Asset operations
func (m *MockQuerier) CreateAsset(ctx context.Context, params sqlc.CreateAssetParams) (sqlc.EpisodeAsset, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.EpisodeAsset), args.Error(1)
}

func (m *MockQuerier) GetAsset(ctx context.Context, id uuid.UUID) (sqlc.EpisodeAsset, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.EpisodeAsset), args.Error(1)
}

func (m *MockQuerier) UpdateAsset(ctx context.Context, params sqlc.UpdateAssetParams) (sqlc.EpisodeAsset, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.EpisodeAsset), args.Error(1)
}

func (m *MockQuerier) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) ListAssetsByEpisode(ctx context.Context, episodeID uuid.UUID) ([]sqlc.EpisodeAsset, error) {
	args := m.Called(ctx, episodeID)
	return args.Get(0).([]sqlc.EpisodeAsset), args.Error(1)
}

// Series operations  
func (m *MockQuerier) CreateSeries(ctx context.Context, params sqlc.CreateSeriesParams) (sqlc.Series, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Series), args.Error(1)
}

func (m *MockQuerier) GetSeries(ctx context.Context, id uuid.UUID) (sqlc.Series, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Series), args.Error(1)
}

func (m *MockQuerier) UpdateSeries(ctx context.Context, params sqlc.UpdateSeriesParams) (sqlc.Series, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Series), args.Error(1)
}

func (m *MockQuerier) DeleteSeries(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) CountSeries(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) ListSeries(ctx context.Context) ([]sqlc.Series, error) {
	args := m.Called(ctx)
	return args.Get(0).([]sqlc.Series), args.Error(1)
}

func (m *MockQuerier) ListSeriesPaginated(ctx context.Context, params sqlc.ListSeriesPaginatedParams) ([]sqlc.Series, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]sqlc.Series), args.Error(1)
}

// Category operations
func (m *MockQuerier) CreateCategory(ctx context.Context, slug string) (sqlc.Category, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(sqlc.Category), args.Error(1)
}

func (m *MockQuerier) GetCategory(ctx context.Context, id uuid.UUID) (sqlc.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Category), args.Error(1)
}

func (m *MockQuerier) UpdateCategory(ctx context.Context, params sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Category), args.Error(1)
}

func (m *MockQuerier) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) CountCategories(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) ListCategories(ctx context.Context) ([]sqlc.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]sqlc.Category), args.Error(1)
}

func (m *MockQuerier) ListCategoriesPaginated(ctx context.Context, params sqlc.ListCategoriesPaginatedParams) ([]sqlc.Category, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]sqlc.Category), args.Error(1)
}
