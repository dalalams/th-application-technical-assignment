package cms

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/pkg/tasks"
	"th-application-technical-assignment/sqlc"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_postImportContent(t *testing.T) {
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
			name: "successful youtube import",
			requestBody: map[string]any{
				"source_type": "youtube",
				"source_url":  "https://youtube.com/watch?v=test123",
				"series_id":   uuid.New().String(),
			},
			mockSeries: sqlc.Series{
				ID:          uuid.New(),
				Title:       "Test Series",
				Description: stringPtr("A test series"),
				CategoryID:  uuid.New(),
				SeriesType:  "podcast",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name: "validation error - missing source_type",
			requestBody: map[string]any{
				"source_url": "https://youtube.com/watch?v=test123",
				"series_id":  uuid.New().String(),
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - missing source_url",
			requestBody: map[string]any{
				"source_type": "youtube",
				"series_id":   uuid.New().String(),
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - missing series_id",
			requestBody: map[string]any{
				"source_type": "youtube",
				"source_url":  "https://youtube.com/watch?v=test123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - invalid source_type",
			requestBody: map[string]any{
				"source_type": "invalid",
				"source_url":  "https://youtube.com/watch?v=test123",
				"series_id":   uuid.New().String(),
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - invalid series_id format",
			requestBody: map[string]any{
				"source_type": "youtube",
				"source_url":  "https://youtube.com/watch?v=test123",
				"series_id":   "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "database error",
			requestBody: map[string]any{
				"source_type": "youtube",
				"source_url":  "https://youtube.com/watch?v=test123",
				"series_id":   uuid.New().String(),
			},
			dbError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "queue error",
			requestBody: map[string]any{
				"source_type": "youtube",
				"source_url":  "https://youtube.com/watch?v=test123",
				"series_id":   uuid.New().String(),
			},
			mockSeries: sqlc.Series{
				ID:          uuid.New(),
				Title:       "Test Series",
				Description: stringPtr("A test series"),
				CategoryID:  uuid.New(),
				SeriesType:  "podcast",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			queueError:     assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			mockQueue := new(tasks.MockQueue)
			validator := validator.New()

			handler := &Handler{
				s: mockStore,
				v: validator,
				q: mockQueue,
			}

			if !tt.expectError || tt.dbError != nil || tt.queueError != nil {
				seriesUUID, _ := uuid.Parse(tt.requestBody["series_id"].(string))

				if tt.dbError != nil {
					mockQueries.On("GetSeries", mock.Anything, seriesUUID).
						Return(sqlc.Series{}, tt.dbError)
				} else {
					mockQueries.On("GetSeries", mock.Anything, seriesUUID).
						Return(tt.mockSeries, nil)

					if tt.queueError != nil {
						mockQueue.On("EnqueueImportContent", mock.Anything, mock.MatchedBy(func(p tasks.ImportContentPayload) bool {
							return p.SourceType == tt.requestBody["source_type"].(string) &&
								p.SourceURL == tt.requestBody["source_url"].(string) &&
								p.SeriesID == tt.requestBody["series_id"].(string)
						})).Return(tt.queueError)
					} else {
						mockQueue.On("EnqueueImportContent", mock.Anything, mock.MatchedBy(func(p tasks.ImportContentPayload) bool {
							return p.SourceType == tt.requestBody["source_type"].(string) &&
								p.SourceURL == tt.requestBody["source_url"].(string) &&
								p.SeriesID == tt.requestBody["series_id"].(string)
						})).Return(nil)
					}
				}
			}

			requestJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.postImportContent(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				assert.Empty(t, recorder.Body.String())
			}

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestHandler_postImportContent_SourceTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sourceType  string
		expectValid bool
	}{
		{"youtube source type", "youtube", true},
		{"invalid source type", "invalid", false},
		{"empty source type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requestBody := map[string]any{
				"source_type": tt.sourceType,
				"source_url":  "https://example.com/test",
				"series_id":   uuid.New().String(),
			}

			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			mockQueue := new(tasks.MockQueue)
			validator := validator.New()

			handler := &Handler{
				s: mockStore,
				v: validator,
				q: mockQueue,
			}

			if tt.expectValid {
				seriesUUID, _ := uuid.Parse(requestBody["series_id"].(string))
				mockSeries := sqlc.Series{
					ID:          uuid.New(),
					Title:       "Test Series",
					Description: stringPtr("A test series"),
					CategoryID:  uuid.New(),
					SeriesType:  "podcast",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockQueries.On("GetSeries", mock.Anything, seriesUUID).
					Return(mockSeries, nil)
				mockQueue.On("EnqueueImportContent", mock.Anything, mock.AnythingOfType("tasks.ImportContentPayload")).
					Return(nil)
			}

			requestJSON, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.postImportContent(recorder, req)

			if tt.expectValid {
				assert.Equal(t, http.StatusNoContent, recorder.Code)
			} else {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			}

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestHandler_postImportContent_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    map[string]any
		expectedStatus int
	}{
		{
			name: "very long source URL",
			requestBody: map[string]any{
				"source_type": "youtube",
				"source_url":  "https://youtube.com/watch?v=" + string(make([]byte, 1000)),
				"series_id":   uuid.New().String(),
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "URL with special characters",
			requestBody: map[string]any{
				"source_type": "youtube",
				"source_url":  "https://youtube.com/watch?v=test&param=value#fragment",
				"series_id":   uuid.New().String(),
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "empty source URL",
			requestBody: map[string]any{
				"source_type": "youtube",
				"source_url":  "",
				"series_id":   uuid.New().String(),
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			mockQueue := new(tasks.MockQueue)
			validator := validator.New()

			handler := &Handler{
				s: mockStore,
				v: validator,
				q: mockQueue,
			}

			if tt.expectedStatus == http.StatusNoContent {
				seriesUUID, _ := uuid.Parse(tt.requestBody["series_id"].(string))
				mockSeries := sqlc.Series{
					ID:          uuid.New(),
					Title:       "Test Series",
					Description: stringPtr("A test series"),
					CategoryID:  uuid.New(),
					SeriesType:  "podcast",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockQueries.On("GetSeries", mock.Anything, seriesUUID).
					Return(mockSeries, nil)
				mockQueue.On("EnqueueImportContent", mock.Anything, mock.AnythingOfType("tasks.ImportContentPayload")).
					Return(nil)
			}

			requestJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.postImportContent(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			mockQueries.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
		})
	}
}

func TestHandler_postImportContent_PayloadVerification(t *testing.T) {
	t.Parallel()

	requestBody := map[string]any{
		"source_type": "youtube",
		"source_url":  "https://youtube.com/watch?v=test123",
		"series_id":   uuid.New().String(),
	}

	mockQueries := new(database.MockQuerier)
	mockStore := &database.Store{Queries: mockQueries}
	mockQueue := new(tasks.MockQueue)
	validator := validator.New()

	handler := &Handler{
		s: mockStore,
		v: validator,
		q: mockQueue,
	}

	seriesUUID, _ := uuid.Parse(requestBody["series_id"].(string))
	mockSeries := sqlc.Series{
		ID:          uuid.New(),
		Title:       "Test Series",
		Description: stringPtr("A test series"),
		CategoryID:  uuid.New(),
		SeriesType:  "podcast",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockQueries.On("GetSeries", mock.Anything, seriesUUID).
		Return(mockSeries, nil)

	var capturedPayload tasks.ImportContentPayload
	mockQueue.On("EnqueueImportContent", mock.Anything, mock.MatchedBy(func(p tasks.ImportContentPayload) bool {
		capturedPayload = p
		return true
	})).Return(nil)

	requestJSON, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.postImportContent(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Empty(t, recorder.Body.String())

	assert.Equal(t, requestBody["source_type"], capturedPayload.SourceType)
	assert.Equal(t, requestBody["source_url"], capturedPayload.SourceURL)
	assert.Equal(t, requestBody["series_id"], capturedPayload.SeriesID)

	mockQueries.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}
