package tasks

import (
	"context"
	"th-application-technical-assignment/sqlc"

	"github.com/stretchr/testify/mock"
)


type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) Enqueue(ctx context.Context, typename string, taskPayload any) error {
    args := m.Called(ctx, typename, taskPayload)
    return args.Error(0)
}

func (m *MockQueue) EnqueueIndexSeries(ctx context.Context, series sqlc.Series) error {
    args := m.Called(ctx, series)
    return args.Error(0)
}

func (m *MockQueue) EnqueueIndexEpisode(ctx context.Context, episode sqlc.Episode, assets []sqlc.EpisodeAsset) error {
	args := m.Called(ctx, episode, assets)
	return args.Error(0)
}

func (m *MockQueue) EnqueueDeleteSeries(ctx context.Context, seriesID string) error {
    args := m.Called(ctx, seriesID)
    return args.Error(0)
}

func (m *MockQueue) EnqueueDeleteEpisode(ctx context.Context, episodeID string) error {
    args := m.Called(ctx, episodeID)
    return args.Error(0)
}

func (m *MockQueue) EnqueueImportContent(ctx context.Context, payload ImportContentPayload) error {
    args := m.Called(ctx, payload)
    return args.Error(0)
}

func (m *MockQueue) Close() error {
    args := m.Called()
    return args.Error(0)
}
