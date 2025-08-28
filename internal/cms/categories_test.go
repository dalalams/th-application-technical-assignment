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
	"th-application-technical-assignment/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_postCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    map[string]any
		mockCategory   sqlc.Category
		dbError        error
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful category creation",
			requestBody: map[string]any{
				"name": "Technology",
			},
			mockCategory: sqlc.Category{
				ID:        uuid.New(),
				Slug:      "technology",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "category with special characters",
			requestBody: map[string]any{
				"name": "Health & Wellness!",
			},
			mockCategory: sqlc.Category{
				ID:        uuid.New(),
				Slug:      "health--wellness",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "validation error - missing name",
			requestBody:    map[string]any{},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - empty name",
			requestBody: map[string]any{
				"name": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "validation error - name too long",
			requestBody: map[string]any{
				"name": "This is a very long category name that exceeds the maximum allowed length of 100 characters and should fail validation",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "database error",
			requestBody: map[string]any{
				"name": "Technology",
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
			validator := validator.New()

			handler := &Handler{
				s: mockStore,
				v: validator,
			}

			if !tt.expectError || tt.dbError != nil {
				if tt.dbError != nil {
					mockQueries.On("CreateCategory", mock.Anything, mock.AnythingOfType("string")).
						Return(sqlc.Category{}, tt.dbError)
				} else {
					expectedSlug := "technology"
					if tt.requestBody["name"] == "Health & Wellness!" {
						expectedSlug = "health--wellness"
					}

					mockQueries.On("CreateCategory", mock.Anything, expectedSlug).
						Return(tt.mockCategory, nil)
				}
			}

			requestJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(requestJSON))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			handler.postCategory(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "id")
				assert.Contains(t, response, "slug")
				assert.Equal(t, tt.mockCategory.Slug, response["slug"])
			}

			mockQueries.AssertExpectations(t)
		})
	}
}

func TestHandler_getCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		categoryID     string
		mockCategory   sqlc.Category
		dbError        error
		expectedStatus int
		expectError    bool
	}{
		{
			name:       "successful category retrieval",
			categoryID: uuid.New().String(),
			mockCategory: sqlc.Category{
				ID:        uuid.New(),
				Slug:      "technology",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid category ID",
			categoryID:     "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "category not found",
			categoryID:     uuid.New().String(),
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

			if tt.categoryID != "invalid-uuid" {
				categoryUUID, _ := uuid.Parse(tt.categoryID)

				if tt.dbError != nil {
					mockQueries.On("GetCategory", mock.Anything, categoryUUID).
						Return(sqlc.Category{}, tt.dbError)
				} else {
					mockQueries.On("GetCategory", mock.Anything, categoryUUID).
						Return(tt.mockCategory, nil)
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/categories/"+tt.categoryID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.categoryID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.getCategory(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.mockCategory.Slug, response["slug"])
			}

			mockQueries.AssertExpectations(t)
		})
	}
}

func TestHandler_deleteCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		categoryID     string
		dbError        error
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "successful category deletion",
			categoryID:     uuid.New().String(),
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:           "invalid category ID",
			categoryID:     "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "database error",
			categoryID:     uuid.New().String(),
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

			if tt.categoryID != "invalid-uuid" {
				categoryUUID, _ := uuid.Parse(tt.categoryID)

				if tt.dbError != nil {
					mockQueries.On("DeleteCategory", mock.Anything, categoryUUID).
						Return(tt.dbError)
				} else {
					mockQueries.On("DeleteCategory", mock.Anything, categoryUUID).
						Return(nil)
				}
			}

			req := httptest.NewRequest(http.MethodDelete, "/categories/"+tt.categoryID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.categoryID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			recorder := httptest.NewRecorder()

			handler.deleteCategory(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				assert.Empty(t, recorder.Body.String())
			}

			mockQueries.AssertExpectations(t)
		})
	}
}
