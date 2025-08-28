// File: internal/discovery/search_test.go
package discovery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	v1 "th-application-technical-assignment/pkg/api/discovery/v1"
	"th-application-technical-assignment/pkg/search"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSearchClient struct {
	mock.Mock
}

func (m *MockSearchClient) SearchSeries(ctx context.Context, req search.SearchRequest) (*search.SearchResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*search.SearchResponse), args.Error(1)
}

func (m *MockSearchClient) SearchEpisodes(ctx context.Context, req search.SearchRequest) (*search.SearchResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*search.SearchResponse), args.Error(1)
}

func (m *MockSearchClient) IndexDocument(ctx context.Context, indexName, documentID string, documentJSON []byte) error {
	args := m.Called(ctx, indexName, documentID, documentJSON)
	return args.Error(0)
}

func (m *MockSearchClient) DeleteDocument(ctx context.Context, indexName, documentID string) error {
	args := m.Called(ctx, indexName, documentID)
	return args.Error(0)
}

func (m *MockSearchClient) IndexExists(ctx context.Context, indexName string) (bool, error) {
	args := m.Called(ctx, indexName)
	return args.Bool(0), args.Error(1)
}

func (m *MockSearchClient) CreateIndex(ctx context.Context, indexName, mapping string) error {
	args := m.Called(ctx, indexName, mapping)
	return args.Error(0)
}

func TestHandler_searchSeries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockResponse   *search.SearchResponse
		mockError      error
		expectedStatus int
		expectedTotal  int64
		expectedPage   int
		expectError    bool
	}{
		{
			name: "successful search with results",
			queryParams: map[string]string{
				"q":         "tech podcast",
				"page":      "1",
				"page_size": "20",
			},
			mockResponse: &search.SearchResponse{
				Total: 2,
				Hits: []map[string]any{
					{"id": "1", "title": "Tech Talk", "type": "podcast"},
					{"id": "2", "title": "Dev Stories", "type": "podcast"},
				},
				Page: 1,
				Size: 20,
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  2,
			expectedPage:   1,
			expectError:    false,
		},
		{
			name: "empty search results",
			queryParams: map[string]string{
				"q":         "nonexistent",
				"page":      "1",
				"page_size": "20",
			},
			mockResponse: &search.SearchResponse{
				Total: 0,
				Hits:  []map[string]any{},
				Page:  1,
				Size:  20,
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  0,
			expectedPage:   1,
			expectError:    false,
		},
		{
			name: "search with filters",
			queryParams: map[string]string{
				"q":           "podcast",
				"category_id": "123e4567-e89b-12d3-a456-426614174000",
				"type":        "podcast",
				"language":    "en",
				"page":        "2",
				"page_size":   "10",
			},
			mockResponse: &search.SearchResponse{
				Total: 11,
				Hits: []map[string]any{
					{"id": "3", "title": "English Podcast", "type": "podcast"},
				},
				Page: 2,
				Size: 10,
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  11,
			expectedPage:   2,
			expectError:    false,
		},
		{
			name:           "search client error",
			queryParams:    map[string]string{"q": "test"},
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "invalid page parameter",
			queryParams: map[string]string{
				"q":    "test",
				"page": "invalid",
			},
			mockResponse: &search.SearchResponse{
				Total: 0,
				Hits:  []map[string]any{},
				Page:  1,
				Size:  20,
			},
			expectedStatus: http.StatusOK,
			expectedPage:   1,
		},
		{
			name: "large page size gets capped",
			queryParams: map[string]string{
				"q":         "test",
				"page_size": "500",
			},
			mockResponse: &search.SearchResponse{
				Total: 1,
				Hits:  []map[string]any{{"id": "1"}},
				Page:  1,
				Size:  20,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSearcher := new(MockSearchClient)
			validator := validator.New()
			handler := &Handler{
				v:            validator,
				searchClient: mockSearcher,
			}

			if tt.mockError != nil {
				mockSearcher.On("SearchSeries", mock.Anything, mock.AnythingOfType("search.SearchRequest")).
					Return((*search.SearchResponse)(nil), tt.mockError)
			} else if tt.mockResponse != nil {
				mockSearcher.On("SearchSeries", mock.Anything, mock.MatchedBy(func(req search.SearchRequest) bool {
					if q, exists := tt.queryParams["q"]; exists && req.Query != q {
						return false
					}
					return true
				})).Return(tt.mockResponse, nil)
			}

			req := httptest.NewRequest(http.MethodGet, "/search/series", nil)

			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for key, value := range tt.queryParams {
					values.Add(key, value)
				}
				req.URL.RawQuery = values.Encode()
			}

			recorder := httptest.NewRecorder()

			handler.searchSeries(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if !tt.expectError {
				var response v1.SearchResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err, "Response should be valid JSON")

				assert.NotNil(t, response.Total)
				assert.NotNil(t, response.Page)
				assert.NotNil(t, response.PageSize)
				assert.NotNil(t, response.Results)

				if tt.expectedTotal > 0 {
					assert.Equal(t, tt.expectedTotal, response.Total)
				}
				if tt.expectedPage > 0 {
					assert.Equal(t, tt.expectedPage, response.Page)
				}

				assert.IsType(t, []map[string]any{}, response.Results)

				if tt.mockResponse != nil {
					assert.Len(t, response.Results, len(tt.mockResponse.Hits))
				}
			}

			mockSearcher.AssertExpectations(t)
		})
	}
}

func TestHandler_searchEpisodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockResponse   *search.SearchResponse
		mockError      error
		expectedStatus int
		expectedTotal  int64
	}{
		{
			name: "successful episode search",
			queryParams: map[string]string{
				"q":         "javascript tutorial",
				"series_id": "456e7890-e12b-34d5-a678-901234567890",
			},
			mockResponse: &search.SearchResponse{
				Total: 12,
				Hits: []map[string]any{
					{"id": "ep1", "title": "JS Basics", "series_id": "456e7890-e12b-34d5-a678-901234567890"},
					{"id": "ep2", "title": "Advanced JS", "series_id": "456e7890-e12b-34d5-a678-901234567890"},
				},
				Page: 1,
				Size: 20,
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  12,
		},
		{
			name: "search client returns error",
			queryParams: map[string]string{
				"q": "test",
			},
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSearcher := new(MockSearchClient)
			validator := validator.New()
			handler := &Handler{
				v:            validator,
				searchClient: mockSearcher,
			}

			if tt.mockError != nil {
				mockSearcher.On("SearchEpisodes", mock.Anything, mock.AnythingOfType("search.SearchRequest")).
					Return((*search.SearchResponse)(nil), tt.mockError)
			} else {
				mockSearcher.On("SearchEpisodes", mock.Anything, mock.MatchedBy(func(req search.SearchRequest) bool {
					// Verify filters were applied correctly
					if seriesID, exists := tt.queryParams["series_id"]; exists {
						return req.Filters != nil && req.Filters["series_id"] == seriesID
					}
					return true
				})).Return(tt.mockResponse, nil)
			}

			req := httptest.NewRequest(http.MethodGet, "/search/episodes", nil)

			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for key, value := range tt.queryParams {
					values.Add(key, value)
				}
				req.URL.RawQuery = values.Encode()
			}

			recorder := httptest.NewRecorder()

			handler.searchEpisodes(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.mockError == nil {
				var response map[string]any
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, float64(tt.expectedTotal), response["total"])
				assert.Contains(t, response, "results")
			}

			mockSearcher.AssertExpectations(t)
		})
	}
}

func TestSearchParameterParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		queryString string
		expectedReq func() map[string]any
	}{
		{
			name:        "default parameters",
			queryString: "",
			expectedReq: func() map[string]any {
				return map[string]any{
					"page":      1,
					"page_size": 20,
					"query":     "",
				}
			},
		},
		{
			name:        "custom pagination",
			queryString: "page=3&page_size=50",
			expectedReq: func() map[string]any {
				return map[string]any{
					"page":      3,
					"page_size": 50,
				}
			},
		},
		{
			name:        "invalid pagination falls back to defaults",
			queryString: "page=invalid&page_size=999",
			expectedReq: func() map[string]any {
				return map[string]any{
					"page":      1,
					"page_size": 20,
				}
			},
		},
		{
			name:        "valid page size at limit",
			queryString: "page_size=100",
			expectedReq: func() map[string]any {
				return map[string]any{
					"page":      1,
					"page_size": 100,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSearcher := new(MockSearchClient)
			validator := validator.New()
			handler := &Handler{
				v:            validator,
				searchClient: mockSearcher,
			}

			expected := tt.expectedReq()

			mockSearcher.On("SearchSeries",
				mock.Anything,
				mock.MatchedBy(func(req search.SearchRequest) bool {
					if expectedPage, exists := expected["page"]; exists {
						if req.Page != expectedPage.(int) {
							t.Logf("Expected page %v, got %v", expectedPage, req.Page)
							return false
						}
					}
					if expectedPageSize, exists := expected["page_size"]; exists {
						if req.PageSize != expectedPageSize.(int) {
							t.Logf("Expected page_size %v, got %v", expectedPageSize, req.PageSize)
							return false
						}
					}
					if expectedQuery, exists := expected["query"]; exists {
						if req.Query != expectedQuery.(string) {
							t.Logf("Expected query %v, got %v", expectedQuery, req.Query)
							return false
						}
					}
					return true
				})).Return(&search.SearchResponse{Total: 0, Hits: []map[string]any{}, Page: 1, Size: 20}, nil)

			req := httptest.NewRequest(http.MethodGet, "/search/series?"+tt.queryString, nil)
			recorder := httptest.NewRecorder()

			handler.searchSeries(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code)
			mockSearcher.AssertExpectations(t)
		})
	}
}
