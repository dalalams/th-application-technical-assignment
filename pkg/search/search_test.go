package search

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSearchQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		request        SearchRequest
		expectedFields []string
		shouldContain  []string
	}{
		{
			name: "simple text query",
			request: SearchRequest{
				Query:    "podcast series",
				Page:     1,
				PageSize: 20,
			},
			expectedFields: []string{"query", "sort"},
			shouldContain:  []string{"multi_match", "podcast series", "title^2", "description"},
		},
		{
			name: "query with filters",
			request: SearchRequest{
				Query:    "tech",
				Page:     1,
				PageSize: 10,
				Filters: map[string]any{
					"category_id": "123-456",
					"type":        "podcast",
				},
			},
			expectedFields: []string{"query", "sort"},
			shouldContain:  []string{"bool", "must", "multi_match", "tech", "term", "123-456", "podcast"},
		},
		{
			name: "empty query - should use match_all",
			request: SearchRequest{
				Query:    "",
				Page:     1,
				PageSize: 5,
			},
			expectedFields: []string{"query", "sort"},
			shouldContain:  []string{"match_all"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			queryJSON := buildSearchQuery(tt.request)

			var queryMap map[string]any
			err := json.Unmarshal([]byte(queryJSON), &queryMap)
			require.NoError(t, err, "Generated query should be valid JSON")

			for _, field := range tt.expectedFields {
				assert.Contains(t, queryMap, field, "Query should contain field: %s", field)
			}

			for _, expected := range tt.shouldContain {
				assert.Contains(t, queryJSON, expected, "Query JSON should contain: %s", expected)
			}

			assert.True(t, json.Valid([]byte(queryJSON)), "Generated query should be valid JSON")
		})
	}
}

func TestSearchIndexNaming(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		indexPrefix      string
		expectedSeries   string
		expectedEpisodes string
	}{
		{
			name:             "standard prefix",
			indexPrefix:      "prod",
			expectedSeries:   "prod-series",
			expectedEpisodes: "prod-episodes",
		},
		{
			name:             "test prefix",
			indexPrefix:      "test",
			expectedSeries:   "test-series",
			expectedEpisodes: "test-episodes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := &Config{IndexPrefix: tt.indexPrefix}

			expectedSeriesIndex := config.IndexPrefix + "-series"
			assert.Equal(t, tt.expectedSeries, expectedSeriesIndex)

			expectedEpisodesIndex := config.IndexPrefix + "-episodes"
			assert.Equal(t, tt.expectedEpisodes, expectedEpisodesIndex)
		})
	}
}
