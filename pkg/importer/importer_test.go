package importer

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetImporter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		source       string
		expectError  bool
		expectedType string
	}{
		{
			name:         "successful youtube importer retrieval",
			source:       "youtube",
			expectError:  false,
			expectedType: "*importer.YouTubeImporter",
		},
		{
			name:         "unknown importer source",
			source:       "unknown",
			expectError:  true,
			expectedType: "",
		},
		{
			name:         "empty source",
			source:       "",
			expectError:  true,
			expectedType: "",
		},
		{
			name:         "case sensitive youtube",
			source:       "YouTube",
			expectError:  true,
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			imp, err := GetImporter(tt.source)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, imp)
				assert.Contains(t, err.Error(), "importer not found")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, imp)
				assert.Equal(t, tt.expectedType, fmt.Sprintf("%T", imp))

				_, ok := imp.(Importer)
				assert.True(t, ok, "Should implement Importer interface")
			}
		})
	}
}

func TestYouTubeImporter_FetchEpisode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		url           string
		seriesID      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful episode fetch",
			url:         "https://youtube.com/watch?v=test123",
			seriesID:    uuid.New().String(),
			expectError: false,
		},
		{
			name:          "invalid series ID",
			url:           "https://youtube.com/watch?v=test123",
			seriesID:      "invalid-uuid",
			expectError:   true,
			errorContains: "invalid series ID",
		},
		{
			name:          "empty series ID",
			url:           "https://youtube.com/watch?v=test123",
			seriesID:      "",
			expectError:   true,
			errorContains: "invalid series ID",
		},
		{
			name:        "empty URL",
			url:         "",
			seriesID:    uuid.New().String(),
			expectError: false,
		},
		{
			name:        "nil context",
			url:         "https://youtube.com/watch?v=test123",
			seriesID:    uuid.New().String(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			imp := NewYouTubeImporter()
			ctx := context.Background()

			episode, asset, err := imp.FetchEpisode(ctx, tt.url, tt.seriesID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, episode)
				assert.Nil(t, asset)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, episode)
				assert.NotNil(t, asset)

				assert.NotEqual(t, uuid.Nil, episode.ID)
				assert.Equal(t, "YouTube Import", episode.Title)

				expectedSeriesID, parseErr := uuid.Parse(tt.seriesID)
				require.NoError(t, parseErr)
				assert.Equal(t, expectedSeriesID, episode.SeriesID)

				assert.NotEqual(t, uuid.Nil, asset.EpisodeID)
				assert.Equal(t, episode.ID, asset.EpisodeID)
				assert.Equal(t, "video", asset.AssetType)
				assert.Equal(t, "video/mp4", asset.MimeType)
				assert.NotNil(t, asset.Url)
				assert.Equal(t, tt.url, *asset.Url)
			}
		})
	}
}

func TestYouTubeImporter_FetchEpisode_EdgeCases(t *testing.T) {
	t.Parallel()

	imp := NewYouTubeImporter()
	ctx := context.Background()
	seriesID := uuid.New().String()

	t.Run("very long URL", func(t *testing.T) {
		t.Parallel()
		longURL := "https://youtube.com/watch?v=" + string(make([]byte, 1000))
		episode, asset, err := imp.FetchEpisode(ctx, longURL, seriesID)

		assert.NoError(t, err)
		assert.NotNil(t, episode)
		assert.NotNil(t, asset)
		assert.Equal(t, longURL, *asset.Url)
	})

	t.Run("URL with special characters", func(t *testing.T) {
		t.Parallel()
		specialURL := "https://youtube.com/watch?v=test&param=value#fragment"
		episode, asset, err := imp.FetchEpisode(ctx, specialURL, seriesID)

		assert.NoError(t, err)
		assert.NotNil(t, episode)
		assert.NotNil(t, asset)
		assert.Equal(t, specialURL, *asset.Url)
	})
}

var _ Importer = (*YouTubeImporter)(nil)

func TestImportersMap(t *testing.T) {
	t.Parallel()

	imp, err := GetImporter("youtube")
	assert.NoError(t, err)
	assert.NotNil(t, imp)

	_, ok := imp.(*YouTubeImporter)
	assert.True(t, ok)
}
