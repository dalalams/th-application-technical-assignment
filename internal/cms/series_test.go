package cms

import (
	"bytes"
	"context"
	"database/sql"
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

func TestHandler_postSeries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    map[string]any
		mockSeries     sqlc.Series
		dbError        error
		queueError     error
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful series creation",
			requestBody: map[string]any{
				"title":       "Tech Talk Podcast",
				"description": "A podcast about technology",
				"category_id": uuid.New().String(),
				"language":    "en",
				"type":        "podcast",
			},
			mockSeries: sqlc.Series{
				ID:          uuid.New(),
				Title:       "Tech Talk Podcast",
				Description: stringPtr("A podcast about technology"),
				CategoryID:  uuid.New(),
				Language:    stringPtr("en"),
				SeriesType:  "podcast",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "minimal series creation",
			requestBody: map[string]any{
				"title":       "Minimal Series",
				"category_id": uuid.New().String(),
				"type":        "documentary",
			},
			mockSeries: sqlc.Series{
				ID:         uuid.New(),
				Title:      "Minimal Series",
				CategoryID: uuid.New(),
				SeriesType: "documentary",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "validation error - missing title",
			requestBody: map[string]any{
				"category_id": uuid.New().String(),
				"type":        "podcast",
				// title required
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - invalid category_id",
			requestBody: map[string]any{
				"title":       "Test Series",
				"category_id": "invalid-uuid",
				"type":        "podcast",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - invalid type",
			requestBody: map[string]any{
				"title":       "Test Series",
				"category_id": uuid.New().String(),
				"type":        "invalid-type",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - description too long",
			requestBody: map[string]any{
				"title":       "Test Series",
				"description": string(make([]byte, 1001)), // over 1000 char limit
				"category_id": uuid.New().String(),
				"type":        "podcast",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "database error",
			requestBody: map[string]any{
				"title":       "Test Series",
				"category_id": uuid.New().String(),
				"type":        "podcast",
			},
			dbError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "queue error - should still succeed",
			requestBody: map[string]any{
				"title":       "Test Series",
				"category_id": uuid.New().String(),
				"type":        "podcast",
			},
			mockSeries: sqlc.Series{
				ID:         uuid.New(),
				Title:      "Test Series",
				CategoryID: uuid.New(),
				SeriesType: "podcast",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			queueError:     assert.AnError,
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
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
					mockQueries.On("CreateSeries", mock.Anything, mock.AnythingOfType("sqlc.CreateSeriesParams")).
						Return(sqlc.Series{}, tt.dbError)
				} else {
					mockQueries.On("CreateSeries", mock.Anything, mock.MatchedBy(func(params sqlc.CreateSeriesParams) bool {
						return params.Title == tt.requestBody["title"].(string) &&
							params.SeriesType == tt.requestBody["type"].(string)
					})).Return(tt.mockSeries, nil)

					if tt.queueError != nil {
						mockQueue.On("EnqueueIndexSeries", mock.Anything, tt.mockSeries).
							Return(tt.queueError)
					} else {
						mockQueue.On("EnqueueIndexSeries", mock.Anything, tt.mockSeries).
							Return(nil)
					}
				}
			}

			// Create request
			requestJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/series", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// ACT
			handler.postSeries(recorder, req)

			// ASSERT
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "id")
				assert.Contains(t, response, "title")
				assert.Contains(t, response, "category_id")
				assert.Contains(t, response, "type")
				assert.Equal(t, tt.mockSeries.Title, response["title"])
				assert.Equal(t, tt.mockSeries.SeriesType, response["type"])
			}

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestHandler_getSeries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		seriesID       string
		mockSeries     sqlc.Series
		dbError        error
		expectedStatus int
		expectError    bool
	}{
		{
			name:     "successful series retrieval",
			seriesID: uuid.New().String(),
			mockSeries: sqlc.Series{
				ID:          uuid.New(),
				Title:       "Tech Talk Podcast",
				Description: stringPtr("A podcast about technology"),
				CategoryID:  uuid.New(),
				Language:    stringPtr("en"),
				SeriesType:  "podcast",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid series ID",
			seriesID:       "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "series not found",
			seriesID:       uuid.New().String(),
			dbError:        sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
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

			if tt.seriesID != "invalid-uuid" {
				seriesUUID, _ := uuid.Parse(tt.seriesID)

				if tt.dbError != nil {
					mockQueries.On("GetSeries", mock.Anything, seriesUUID).
						Return(sqlc.Series{}, tt.dbError)
				} else {
					mockQueries.On("GetSeries", mock.Anything, seriesUUID).
						Return(tt.mockSeries, nil)
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/series/"+tt.seriesID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.seriesID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.getSeries(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.mockSeries.Title, response["title"])
				assert.Equal(t, tt.mockSeries.SeriesType, response["type"])

				if tt.mockSeries.Description != nil {
					assert.Equal(t, *tt.mockSeries.Description, response["description"])
				}
			}

			mockQueries.AssertExpectations(t)
		})
	}
}

func TestHandler_putSeries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		seriesID       string
		requestBody    map[string]any
		mockSeries     sqlc.Series
		dbError        error
		queueError     error
		expectedStatus int
		expectError    bool
	}{
		{
			name:     "successful series update",
			seriesID: uuid.New().String(),
			requestBody: map[string]any{
				"title":       "Updated Tech Talk",
				"description": "An updated description",
				"category_id": uuid.New().String(),
				"language":    "en",
				"type":        "podcast",
			},
			mockSeries: sqlc.Series{
				ID:          uuid.New(),
				Title:       "Updated Tech Talk",
				Description: stringPtr("An updated description"),
				CategoryID:  uuid.New(),
				Language:    stringPtr("en"),
				SeriesType:  "podcast",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid series ID",
			seriesID:       "invalid-uuid",
			requestBody:    map[string]any{"title": "Test", "type": "podcast"},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:     "validation error",
			seriesID: uuid.New().String(),
			requestBody: map[string]any{
				"title": "", // should fail validation
				"type":  "podcast",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:     "database error",
			seriesID: uuid.New().String(),
			requestBody: map[string]any{
				"title": "Test Series",
				"type":  "podcast",
			},
			dbError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
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

			if tt.seriesID != "invalid-uuid" && (!tt.expectError || tt.dbError != nil) {
				seriesUUID, _ := uuid.Parse(tt.seriesID)

				if tt.dbError != nil {
					mockQueries.On("UpdateSeries", mock.Anything, mock.AnythingOfType("sqlc.UpdateSeriesParams")).
						Return(sqlc.Series{}, tt.dbError)
				} else {
					mockQueries.On("UpdateSeries", mock.Anything, mock.MatchedBy(func(params sqlc.UpdateSeriesParams) bool {
						return params.ID == seriesUUID && params.Title == tt.requestBody["title"].(string)
					})).Return(tt.mockSeries, nil)

					mockQueue.On("EnqueueIndexSeries", mock.Anything, tt.mockSeries).Return(tt.queueError)
				}
			}

			requestJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/series/"+tt.seriesID, bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.seriesID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.putSeries(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.mockSeries.Title, response["title"])
			}

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestHandler_deleteSeries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		seriesID       string
		dbError        error
		queueError     error
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful series deletion",
			seriesID:       uuid.New().String(),
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:           "invalid series ID",
			seriesID:       "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "database error",
			seriesID:       uuid.New().String(),
			dbError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "queue error - should still succeed",
			seriesID:       uuid.New().String(),
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

			if tt.seriesID != "invalid-uuid" {
				seriesUUID, _ := uuid.Parse(tt.seriesID)

				if tt.dbError != nil {
					mockQueries.On("DeleteSeries", mock.Anything, seriesUUID).
						Return(tt.dbError)
				} else {
					mockQueries.On("DeleteSeries", mock.Anything, seriesUUID).
						Return(nil)

					if tt.queueError != nil {
						mockQueue.On("EnqueueDeleteSeries", mock.Anything, tt.seriesID).
							Return(tt.queueError)
					} else {
						mockQueue.On("EnqueueDeleteSeries", mock.Anything, tt.seriesID).
							Return(nil)
					}
				}
			}

			req := httptest.NewRequest(http.MethodDelete, "/series/"+tt.seriesID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.seriesID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.deleteSeries(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				assert.Empty(t, recorder.Body.String())
			}

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}
