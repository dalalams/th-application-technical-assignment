package util

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPaginationMath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		itemCount int64
		pageSize  int
		expected  int
	}{
		{
			name:      "perfect division",
			itemCount: 100,
			pageSize:  20,
			expected:  5,
		},
		{
			name:      "needs rounding up",
			itemCount: 101,
			pageSize:  20,
			expected:  6,
		},
		{
			name:      "single remainder",
			itemCount: 21,
			pageSize:  20,
			expected:  2,
		},
		{
			name:      "zero items",
			itemCount: 0,
			pageSize:  10,
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CalculatePaginationResponse(1, tt.pageSize, tt.itemCount)
			assert.Equal(t, tt.expected, result.PageCount, "Page count calculation should be correct")
		})
	}
}

func TestPaginatedResponse(t *testing.T) {
	t.Parallel()

	data := []string{"item1", "item2", "item3"}
	metadata := PaginationMetadata{
		Page:      2,
		PageSize:  10,
		ItemCount: 25,
		PageCount: 3,
	}

	response := PaginatedResponse[string]{
		Data:       data,
		Pagination: metadata,
	}

	assert.Len(t, response.Data, 3)
	assert.Equal(t, "item1", response.Data[0])
	assert.Equal(t, 2, response.Pagination.Page)
	assert.Equal(t, int64(25), response.Pagination.ItemCount)
}

func TestFetchPaginatedData_TimeoutScenarios(t *testing.T) {
	tests := []struct {
		name           string
		fetchCountFunc func(context.Context) (int64, error)
		fetchDataFunc  func(context.Context) ([]string, error)
		expectError    bool
		timeout        time.Duration
	}{
		{
			name: "count function timeout",
			fetchCountFunc: func(ctx context.Context) (int64, error) {
				select {
				case <-ctx.Done():
					return 0, ctx.Err()
				case <-time.After(2 * time.Second):
					return 100, nil
				}
			},
			fetchDataFunc: func(ctx context.Context) ([]string, error) {
				return []string{"data1", "data2"}, nil
			},
			expectError: true,
			timeout:     100 * time.Millisecond,
		},
		{
			name: "data function timeout",
			fetchCountFunc: func(ctx context.Context) (int64, error) {
				return 100, nil
			},
			fetchDataFunc: func(ctx context.Context) ([]string, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(2 * time.Second):
					return []string{"data1", "data2"}, nil
				}
			},
			expectError: true,
			timeout:     100 * time.Millisecond,
		},
		{
			name: "both functions timeout",
			fetchCountFunc: func(ctx context.Context) (int64, error) {
				select {
				case <-ctx.Done():
					return 0, ctx.Err()
				case <-time.After(2 * time.Second):
					return 100, nil
				}
			},
			fetchDataFunc: func(ctx context.Context) ([]string, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(2 * time.Second):
					return []string{"data1", "data2"}, nil
				}
			},
			expectError: true,
			timeout:     100 * time.Millisecond,
		},
		{
			name: "successful with short timeout",
			fetchCountFunc: func(ctx context.Context) (int64, error) {
				return 100, nil
			},
			fetchDataFunc: func(ctx context.Context) ([]string, error) {
				return []string{"data1", "data2"}, nil
			},
			expectError: false,
			timeout:     1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			count, data, err := FetchPaginatedData(ctx, tt.fetchCountFunc, tt.fetchDataFunc)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(100), count)
				assert.Len(t, data, 2)
			}
		})
	}
}

func TestFetchPaginatedData_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		fetchCountFunc func(context.Context) (int64, error)
		fetchDataFunc  func(context.Context) ([]string, error)
		expectError    bool
		expectedError  string
	}{
		{
			name: "count function error",
			fetchCountFunc: func(ctx context.Context) (int64, error) {
				return 0, errors.New("database connection failed")
			},
			fetchDataFunc: func(ctx context.Context) ([]string, error) {
				return []string{"data1", "data2"}, nil
			},
			expectError:   true,
			expectedError: "database connection failed",
		},
		{
			name: "data function error",
			fetchCountFunc: func(ctx context.Context) (int64, error) {
				return 100, nil
			},
			fetchDataFunc: func(ctx context.Context) ([]string, error) {
				return nil, errors.New("query execution failed")
			},
			expectError:   true,
			expectedError: "query execution failed",
		},
		{
			name: "both functions error",
			fetchCountFunc: func(ctx context.Context) (int64, error) {
				return 0, errors.New("count error")
			},
			fetchDataFunc: func(ctx context.Context) ([]string, error) {
				return nil, errors.New("data error")
			},
			expectError:   true,
			expectedError: "error",
		},
		{
			name: "successful execution",
			fetchCountFunc: func(ctx context.Context) (int64, error) {
				return 100, nil
			},
			fetchDataFunc: func(ctx context.Context) ([]string, error) {
				return []string{"data1", "data2", "data3"}, nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			count, data, err := FetchPaginatedData(ctx, tt.fetchCountFunc, tt.fetchDataFunc)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(100), count)
				assert.Len(t, data, 3)
			}
		})
	}
}
