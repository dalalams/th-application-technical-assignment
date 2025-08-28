package storage

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMinioClient struct {
	mock.Mock
}

func (m *MockMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	args := m.Called(ctx, bucketName)
	return args.Bool(0), args.Error(1)
}

func (m *MockMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	args := m.Called(ctx, bucketName, opts)
	return args.Error(0)
}

func (m *MockMinioClient) PresignedPutObject(ctx context.Context, bucketName, objectName string, expiration time.Duration) (*url.URL, error) {
	args := m.Called(ctx, bucketName, objectName, expiration)
	return args.Get(0).(*url.URL), args.Error(1)
}

type TestClient struct {
	client     MinioClient
	bucketName string
}

var _ MinioClient = (*minio.Client)(nil)

func TestMinIOClient_EnsureBucket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		bucketName       string
		bucketExists     bool
		bucketExistsErr  error
		makeBucketErr    error
		expectMakeBucket bool
		expectError      bool
	}{
		{
			name:             "bucket already exists",
			bucketName:       "existing-bucket",
			bucketExists:     true,
			bucketExistsErr:  nil,
			expectMakeBucket: false,
			expectError:      false,
		},
		{
			name:             "bucket doesn't exist - create successfully",
			bucketName:       "new-bucket",
			bucketExists:     false,
			bucketExistsErr:  nil,
			makeBucketErr:    nil,
			expectMakeBucket: true,
			expectError:      false,
		},
		{
			name:             "error checking bucket existence",
			bucketName:       "error-bucket",
			bucketExists:     false,
			bucketExistsErr:  assert.AnError,
			expectMakeBucket: false,
			expectError:      true,
		},
		{
			name:             "bucket doesn't exist but creation fails",
			bucketName:       "fail-create-bucket",
			bucketExists:     false,
			bucketExistsErr:  nil,
			makeBucketErr:    assert.AnError,
			expectMakeBucket: true,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockClient := new(MockMinioClient)

			mockClient.On("BucketExists", mock.Anything, tt.bucketName).
				Return(tt.bucketExists, tt.bucketExistsErr)

			if tt.expectMakeBucket {
				mockClient.On("MakeBucket", mock.Anything, tt.bucketName, mock.AnythingOfType("minio.MakeBucketOptions")).
					Return(tt.makeBucketErr)
			}

			client := &MinIOStorage{
				client:     mockClient,
				bucketName: tt.bucketName,
			}

			ctx := context.Background()
			err := client.EnsureBucket(ctx)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
