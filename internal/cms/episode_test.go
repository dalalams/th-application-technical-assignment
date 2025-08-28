package cms

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/pkg/tasks"
	"th-application-technical-assignment/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_postSeriesEpisode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    map[string]any
		mockEpisode    sqlc.Episode
		mockAssets     []sqlc.EpisodeAsset
		dbError        error
		queueError     error
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful episode creation",
			requestBody: map[string]any{
				"series_id":        uuid.New().String(),
				"title":            "Test Episode",
				"description":      "A test episode description",
				"duration_seconds": 3600,
				"publish_date":     "2023-06-01T12:00:00Z",
			},
			mockEpisode: sqlc.Episode{
				ID:              uuid.New(),
				SeriesID:        uuid.New(),
				Title:           "Test Episode",
				Description:     stringPtr("A test episode description"),
				DurationSeconds: int32Ptr(3600),
				PublishDate:     timePtr(time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC)),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			mockAssets:     []sqlc.EpisodeAsset{},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "minimal episode creation",
			requestBody: map[string]any{
				"series_id": uuid.New().String(),
				"title":     "Minimal Episode",
			},
			mockEpisode: sqlc.Episode{
				ID:        uuid.New(),
				SeriesID:  uuid.New(),
				Title:     "Minimal Episode",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockAssets:     []sqlc.EpisodeAsset{},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "validation error - missing title",
			requestBody: map[string]any{
				"series_id": uuid.New().String(),
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - invalid series_id",
			requestBody: map[string]any{
				"series_id": "invalid-uuid",
				"title":     "Test Episode",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "database error",
			requestBody: map[string]any{
				"series_id": uuid.New().String(),
				"title":     "Test Episode",
			},
			dbError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "queue error - should still succeed",
			requestBody: map[string]any{
				"series_id": uuid.New().String(),
				"title":     "Test Episode",
			},
			mockEpisode: sqlc.Episode{
				ID:        uuid.New(),
				SeriesID:  uuid.New(),
				Title:     "Test Episode",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockAssets:     []sqlc.EpisodeAsset{},
			queueError:     assert.AnError,
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			mockQueue := new(tasks.MockQueue)
			validator := validator.New()

			handler := &Handler{
				s: mockStore,
				v: validator,
				q: mockQueue,
			}

			if !tt.expectError || tt.dbError != nil {
				if tt.dbError != nil {
					mockQueries.On("CreateEpisode", mock.Anything, mock.AnythingOfType("sqlc.CreateEpisodeParams")).
						Return(sqlc.Episode{}, tt.dbError)
				} else {
					mockQueries.On("CreateEpisode", mock.Anything, mock.MatchedBy(func(params sqlc.CreateEpisodeParams) bool {
						return params.Title == tt.requestBody["title"].(string)
					})).Return(tt.mockEpisode, nil)

					mockQueries.On("ListAssetsByEpisode", mock.Anything, tt.mockEpisode.ID).
						Return(tt.mockAssets, nil)

					if tt.queueError != nil {
						mockQueue.On("EnqueueIndexEpisode", mock.Anything, tt.mockEpisode, tt.mockAssets).
							Return(tt.queueError)
					} else {
						mockQueue.On("EnqueueIndexEpisode", mock.Anything, tt.mockEpisode, tt.mockAssets).
							Return(nil)
					}
				}
			}

			requestJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/series/episodes", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.postSeriesEpisode(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err, "Response should be valid JSON")

				assert.Contains(t, response, "id")
				assert.Contains(t, response, "title")
				assert.Contains(t, response, "series_id")
				assert.Contains(t, response, "created_at")
				assert.Contains(t, response, "updated_at")
				assert.Contains(t, response, "assets")

				assert.Equal(t, tt.mockEpisode.Title, response["title"])
			}

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestHandler_getSeriesEpisode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		episodeID      string
		mockEpisode    sqlc.Episode
		mockAssets     []sqlc.EpisodeAsset
		dbError        error
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "successful episode retrieval",
			episodeID: uuid.New().String(),
			mockEpisode: sqlc.Episode{
				ID:          uuid.New(),
				SeriesID:    uuid.New(),
				Title:       "Test Episode",
				Description: stringPtr("Test description"),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockAssets: []sqlc.EpisodeAsset{
				{
					ID:        uuid.New(),
					EpisodeID: uuid.New(),
					AssetType: "audio",
					MimeType:  "audio/mpeg",
					CreatedAt: time.Now(),
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid episode ID",
			episodeID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "database error - generic",
			episodeID:      uuid.New().String(),
			dbError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			validator := validator.New()

			handler := &Handler{
				s: mockStore,
				v: validator,
			}

			if tt.episodeID != "invalid-uuid" {
				episodeUUID, _ := uuid.Parse(tt.episodeID)

				if tt.dbError != nil {
					mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
						Return(sqlc.Episode{}, tt.dbError)
				} else {
					mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
						Return(tt.mockEpisode, nil)
					mockQueries.On("ListAssetsByEpisode", mock.Anything, episodeUUID).
						Return(tt.mockAssets, nil)
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/series/episodes/"+tt.episodeID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.episodeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.getSeriesEpisode(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.mockEpisode.Title, response["title"])
				assert.Contains(t, response, "assets")

				assets := response["assets"].([]interface{})
				assert.Len(t, assets, len(tt.mockAssets))
			}

			mockQueries.AssertExpectations(t)
		})
	}
}

func TestHandler_deleteSeriesEpisode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		episodeID      string
		dbError        error
		queueError     error
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful episode deletion",
			episodeID:      uuid.New().String(),
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:           "invalid episode ID",
			episodeID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "database error",
			episodeID:      uuid.New().String(),
			dbError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "queue error - should still succeed",
			episodeID:      uuid.New().String(),
			queueError:     assert.AnError,
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			mockQueue := new(tasks.MockQueue)
			validator := validator.New()

			handler := &Handler{
				s: mockStore,
				v: validator,
				q: mockQueue,
			}

			if tt.episodeID != "invalid-uuid" {
				episodeUUID, _ := uuid.Parse(tt.episodeID)

				if tt.dbError != nil {
					mockQueries.On("DeleteEpisode", mock.Anything, episodeUUID).
						Return(tt.dbError)
				} else {
					mockQueries.On("DeleteEpisode", mock.Anything, episodeUUID).
						Return(nil)

					if tt.queueError != nil {
						mockQueue.On("EnqueueDeleteEpisode", mock.Anything, tt.episodeID).
							Return(tt.queueError)
					} else {
						mockQueue.On("EnqueueDeleteEpisode", mock.Anything, tt.episodeID).
							Return(nil)
					}
				}
			}

			req := httptest.NewRequest(http.MethodDelete, "/series/episodes/"+tt.episodeID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.episodeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.deleteSeriesEpisode(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				assert.Empty(t, recorder.Body.String())
			}

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string     { return &s }
func int32Ptr(i int32) *int32        { return &i }
func timePtr(t time.Time) *time.Time { return &t }
