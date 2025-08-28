package cms

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	v1 "th-application-technical-assignment/pkg/api/cms/v1"
	"th-application-technical-assignment/pkg/database"
	"th-application-technical-assignment/sqlc"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorageClient struct {
	mock.Mock
}

func (m *MockStorageClient) EnsureBucket(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageClient) GeneratePresignedPutURL(ctx context.Context, key string, expiry time.Duration) (*url.URL, error) {
	args := m.Called(ctx, key, expiry)
	return args.Get(0).(*url.URL), args.Error(1)
}

func (m *MockStorageClient) GenerateKey(seriesID, episodeID uuid.UUID, filename string) string {
	args := m.Called(seriesID, episodeID, filename)
	return args.String(0)
}

func (m *MockStorageClient) GetBucketName() string {
	args := m.Called()
	return args.String(0)
}

func TestHandler_getEpisodeUploadURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		episodeID      string
		requestBody    map[string]any
		mockEpisode    sqlc.Episode
		mockKey        string
		mockURL        *url.URL
		mockBucketName string
		dbError        error
		storageError   error
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "successful upload url generation",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"filename": "episode.mp4",
			},
			mockEpisode: sqlc.Episode{
				ID:       uuid.New(),
				SeriesID: uuid.New(),
				Title:    "Test Episode",
			},
			mockKey:        "series/episode/filename.mp4",
			mockURL:        &url.URL{Scheme: "https", Host: "example.com", Path: "/upload"},
			mockBucketName: "test-bucket",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing episode id parameter",
			episodeID:      "",
			requestBody:    map[string]any{"filename": "test.mp4"},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid episode id format",
			episodeID:      "invalid-uuid",
			requestBody:    map[string]any{"filename": "test.mp4"},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:        "validation error - missing filename",
			episodeID:   uuid.New().String(),
			requestBody: map[string]any{
				// filename is required but missing
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "validation error - empty filename",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"filename": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "validation error - filename too long",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"filename": string(make([]byte, 256)), // over char limit
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "episode not found",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"filename": "test.mp4",
			},
			dbError:        sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:      "storage error",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"filename": "test.mp4",
			},
			mockEpisode: sqlc.Episode{
				ID:       uuid.New(),
				SeriesID: uuid.New(),
				Title:    "Test Episode",
			},
			mockKey:        "series/episode/filename.mp4",
			storageError:   assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			mockStorage := new(MockStorageClient)
			validator := validator.New()

			handler := &Handler{
				s:  mockStore,
				v:  validator,
				mc: mockStorage,
			}

			if tt.episodeID != "" && tt.episodeID != "invalid-uuid" &&
				!(tt.name == "validation error - missing filename" ||
					tt.name == "validation error - empty filename" ||
					tt.name == "validation error - filename too long") {
				episodeUUID, _ := uuid.Parse(tt.episodeID)

				if tt.dbError != nil {
					mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
						Return(sqlc.Episode{}, tt.dbError)
				} else {
					mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
						Return(tt.mockEpisode, nil)

					if tt.mockKey != "" {
						mockStorage.On("GenerateKey", tt.mockEpisode.SeriesID, tt.mockEpisode.ID, tt.requestBody["filename"].(string)).
							Return(tt.mockKey)

						if tt.storageError != nil {
							mockStorage.On("GeneratePresignedPutURL", mock.Anything, tt.mockKey, mock.AnythingOfType("time.Duration")).
								Return((*url.URL)(nil), tt.storageError)
						} else {
							mockStorage.On("GeneratePresignedPutURL", mock.Anything, tt.mockKey, mock.AnythingOfType("time.Duration")).
								Return(tt.mockURL, nil)

							mockStorage.On("GetBucketName").Return(tt.mockBucketName)
						}
					}
				}
			}

			requestJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/series/episodes/"+tt.episodeID+"/upload-url", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.episodeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.getEpisodeUploadURL(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response v1.UploadURLResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.mockURL.String(), response.UploadURL)
				assert.Equal(t, tt.mockKey, response.S3Key)
				assert.Equal(t, tt.mockBucketName, response.S3Bucket)
				assert.True(t, response.ExpiresAt.After(time.Now()))
			}

			mockQueries.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestHandler_confirmEpisodeUpload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		episodeID        string
		requestBody      map[string]any
		mockEpisode      sqlc.Episode
		mockAssets       []sqlc.EpisodeAsset
		mockAsset        sqlc.EpisodeAsset
		dbError          error
		createAssetError error
		listAssetsError  error
		expectedStatus   int
		expectError      bool
	}{
		{
			name:      "successful upload confirmation",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"s3_key":     "series/episode/file.mp4",
				"mime_type":  "video/mp4",
				"size":       1024000,
				"asset_type": "video",
			},
			mockEpisode: sqlc.Episode{
				ID:       uuid.New(),
				SeriesID: uuid.New(),
				Title:    "Test Episode",
			},
			mockAsset: sqlc.EpisodeAsset{
				ID:        uuid.New(),
				EpisodeID: uuid.New(),
				AssetType: "video",
				MimeType:  "video/mp4",
				SizeBytes: int64Ptr(1024000),
				Url:       stringPtr("series/episode/file.mp4"),
			},
			mockAssets: []sqlc.EpisodeAsset{
				{
					ID:        uuid.New(),
					EpisodeID: uuid.New(),
					AssetType: "video",
					MimeType:  "video/mp4",
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing episode ID parameter",
			episodeID:      "",
			requestBody:    map[string]any{"s3_key": "test", "mime_type": "video/mp4", "size": 1000, "asset_type": "video"},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid episode ID format",
			episodeID:      "invalid-uuid",
			requestBody:    map[string]any{"s3_key": "test", "mime_type": "video/mp4", "size": 1000, "asset_type": "video"},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "validation error - missing s3_key",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"mime_type":  "video/mp4",
				"size":       1000,
				"asset_type": "video",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "validation error - missing mime_type",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"s3_key":     "test",
				"size":       1000,
				"asset_type": "video",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "validation error - invalid size",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"s3_key":     "test",
				"mime_type":  "video/mp4",
				"size":       0, // must be > 0
				"asset_type": "video",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "validation error - invalid asset_type",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"s3_key":     "test",
				"mime_type":  "video/mp4",
				"size":       1000,
				"asset_type": "invalid", // must be audio, video, or thumbnail
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "episode not found",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"s3_key":     "test",
				"mime_type":  "video/mp4",
				"size":       1000,
				"asset_type": "video",
			},
			dbError:        sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:      "create asset fails",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"s3_key":     "test",
				"mime_type":  "video/mp4",
				"size":       1000,
				"asset_type": "video",
			},
			mockEpisode: sqlc.Episode{
				ID:       uuid.New(),
				SeriesID: uuid.New(),
				Title:    "Test Episode",
			},
			createAssetError: assert.AnError,
			expectedStatus:   http.StatusInternalServerError,
			expectError:      true,
		},
		{
			name:      "get episode after asset creation fails",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"s3_key":     "test",
				"mime_type":  "video/mp4",
				"size":       1000,
				"asset_type": "video",
			},
			mockEpisode: sqlc.Episode{
				ID:       uuid.New(),
				SeriesID: uuid.New(),
				Title:    "Test Episode",
			},
			mockAsset: sqlc.EpisodeAsset{
				ID:        uuid.New(),
				EpisodeID: uuid.New(),
				AssetType: "video",
				MimeType:  "video/mp4",
			},
			dbError:        sql.ErrNoRows,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:      "list assets fails",
			episodeID: uuid.New().String(),
			requestBody: map[string]any{
				"s3_key":     "test",
				"mime_type":  "video/mp4",
				"size":       1000,
				"asset_type": "video",
			},
			mockEpisode: sqlc.Episode{
				ID:       uuid.New(),
				SeriesID: uuid.New(),
				Title:    "Test Episode",
			},
			mockAsset: sqlc.EpisodeAsset{
				ID:        uuid.New(),
				EpisodeID: uuid.New(),
				AssetType: "video",
				MimeType:  "video/mp4",
			},
			listAssetsError: assert.AnError,
			expectedStatus:  http.StatusInternalServerError,
			expectError:     true,
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

			if tt.episodeID != "" && tt.episodeID != "invalid-uuid" &&
				!(tt.name == "validation error - missing s3_key" ||
					tt.name == "validation error - missing mime_type" ||
					tt.name == "validation error - invalid size" ||
					tt.name == "validation error - invalid asset_type") {

				episodeUUID, _ := uuid.Parse(tt.episodeID)

				if tt.createAssetError != nil {
					mockQueries.On("CreateAsset", mock.Anything, mock.AnythingOfType("sqlc.CreateAssetParams")).
						Return(sqlc.EpisodeAsset{}, tt.createAssetError)
				} else if tt.dbError != nil && tt.name == "episode not found" {
					mockQueries.On("CreateAsset", mock.Anything, mock.AnythingOfType("sqlc.CreateAssetParams")).
						Return(tt.mockAsset, nil)
					mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
						Return(sqlc.Episode{}, tt.dbError)
				} else if tt.dbError != nil && tt.name == "get episode after asset creation fails" {
					mockQueries.On("CreateAsset", mock.Anything, mock.AnythingOfType("sqlc.CreateAssetParams")).
						Return(tt.mockAsset, nil)
					mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
						Return(sqlc.Episode{}, tt.dbError)
				} else if tt.listAssetsError != nil {
					mockQueries.On("CreateAsset", mock.Anything, mock.AnythingOfType("sqlc.CreateAssetParams")).
						Return(tt.mockAsset, nil)
					mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
						Return(tt.mockEpisode, nil)
					mockQueries.On("ListAssetsByEpisode", mock.Anything, episodeUUID).
						Return([]sqlc.EpisodeAsset{}, tt.listAssetsError)
				} else {
					mockQueries.On("CreateAsset", mock.Anything, mock.MatchedBy(func(params sqlc.CreateAssetParams) bool {
						return params.EpisodeID == episodeUUID &&
							params.AssetType == tt.requestBody["asset_type"].(string) &&
							params.MimeType == tt.requestBody["mime_type"].(string)
					})).Return(tt.mockAsset, nil)

					mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
						Return(tt.mockEpisode, nil)
					mockQueries.On("ListAssetsByEpisode", mock.Anything, episodeUUID).
						Return(tt.mockAssets, nil)
				}
			}

			requestJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/series/episodes/"+tt.episodeID+"/upload-confirm", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.episodeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.confirmEpisodeUpload(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "id")
				assert.Contains(t, response, "title")
				assert.Contains(t, response, "assets")
			}

			mockQueries.AssertExpectations(t)
		})
	}
}

func TestHandler_confirmEpisodeUpload_AssetTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		assetType   string
		expectValid bool
	}{
		{"audio asset type", "audio", true},
		{"video asset type", "video", true},
		{"thumbnail asset type", "thumbnail", true},
		{"invalid asset type", "document", false},
		{"empty asset type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			episodeID := uuid.New().String()
			requestBody := map[string]any{
				"s3_key":     "test",
				"mime_type":  "video/mp4",
				"size":       1000,
				"asset_type": tt.assetType,
			}

			mockQueries := new(database.MockQuerier)
			mockStore := &database.Store{Queries: mockQueries}
			validator := validator.New()

			handler := &Handler{
				s: mockStore,
				v: validator,
			}

			if tt.expectValid {
				episodeUUID, _ := uuid.Parse(episodeID)
				mockEpisode := sqlc.Episode{
					ID:       uuid.New(),
					SeriesID: uuid.New(),
					Title:    "Test Episode",
				}
				mockAsset := sqlc.EpisodeAsset{
					ID:        uuid.New(),
					EpisodeID: uuid.New(),
					AssetType: tt.assetType,
					MimeType:  "video/mp4",
				}
				mockAssets := []sqlc.EpisodeAsset{mockAsset}

				mockQueries.On("CreateAsset", mock.Anything, mock.AnythingOfType("sqlc.CreateAssetParams")).
					Return(mockAsset, nil)
				mockQueries.On("GetEpisode", mock.Anything, episodeUUID).
					Return(mockEpisode, nil)
				mockQueries.On("ListAssetsByEpisode", mock.Anything, episodeUUID).
					Return(mockAssets, nil)
			}

			requestJSON, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, "/series/episodes/"+episodeID+"/upload-confirm", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", episodeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.confirmEpisodeUpload(recorder, req)

			if tt.expectValid {
				assert.Equal(t, http.StatusOK, recorder.Code)
			} else {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			}

			mockQueries.AssertExpectations(t)
		})
	}
}

func int64Ptr(i int64) *int64 { return &i }
