package tasks

import (
	"context"
	"encoding/json"
	"testing"
	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/sqlc"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImportEpisodeTaskProcessor_ProcessTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		payload          ImportContentPayload
		importerError    error
		createEpError    error
		createAssetError error
		enqueueError     error
		expectError      bool
	}{
		{
			name: "successful youtube import",
			payload: ImportContentPayload{
				SourceType: "youtube",
				SourceURL:  "https://youtube.com/watch?v=test123",
				SeriesID:   uuid.New().String(),
			},
			expectError: false,
		},
		{
			name: "database create episode fails",
			payload: ImportContentPayload{
				SourceType: "youtube",
				SourceURL:  "https://youtube.com/watch?v=test123",
				SeriesID:   uuid.New().String(),
			},
			createEpError: assert.AnError,
			expectError:   true,
		},
		{
			name: "database create asset fails",
			payload: ImportContentPayload{
				SourceType: "youtube",
				SourceURL:  "https://youtube.com/watch?v=test123",
				SeriesID:   uuid.New().String(),
			},
			createAssetError: assert.AnError,
			expectError:      true,
		},
		{
			name: "enqueue index task fails",
			payload: ImportContentPayload{
				SourceType: "youtube",
				SourceURL:  "https://youtube.com/watch?v=test123",
				SeriesID:   uuid.New().String(),
			},
			enqueueError: assert.AnError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			mockQueue := new(MockQueue)

			processor := NewImportEpisodeTaskProcessor(mockStore, mockQueue)

			payloadJSON, _ := json.Marshal(tt.payload)
			task := asynq.NewTask(TypeImportContent, payloadJSON)

			seriesUUID, _ := uuid.Parse(tt.payload.SeriesID)
			episodeID := uuid.New()
			assetID := uuid.New()

			expectedEpisode := &sqlc.Episode{
				ID:       episodeID,
				SeriesID: seriesUUID,
				Title:    "YouTube Import",
			}
			expectedAsset := &sqlc.EpisodeAsset{
				EpisodeID: episodeID,
				AssetType: "video",
				MimeType:  "video/mp4",
				Url:       &tt.payload.SourceURL,
			}

			if tt.importerError == nil {
				createdEpisode := sqlc.Episode{
					ID:       episodeID,
					SeriesID: seriesUUID,
					Title:    expectedEpisode.Title,
				}
				mockQueries.On("CreateEpisode", mock.Anything, mock.MatchedBy(func(params sqlc.CreateEpisodeParams) bool {
					return params.SeriesID == seriesUUID && params.Title == "YouTube Import"
				})).Return(createdEpisode, tt.createEpError)

				if tt.createEpError == nil {
					createdAsset := sqlc.EpisodeAsset{
						ID:        assetID,
						EpisodeID: episodeID,
						AssetType: expectedAsset.AssetType,
						MimeType:  expectedAsset.MimeType,
						Url:       expectedAsset.Url,
					}
					mockQueries.On("CreateAsset", mock.Anything, mock.MatchedBy(func(params sqlc.CreateAssetParams) bool {
						return params.EpisodeID == episodeID && params.AssetType == "video"
					})).Return(createdAsset, tt.createAssetError)

					if tt.createAssetError == nil {
						mockQueue.On("EnqueueIndexEpisode", mock.Anything, createdEpisode, []sqlc.EpisodeAsset{createdAsset}).Return(tt.enqueueError)
					}
				}
			}

			err := processor.ProcessTask(context.Background(), task)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}
