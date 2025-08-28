package v1

import "time"

type UploadURLRequest struct {
	Filename string `json:"filename" validate:"required,min=1,max=255"`
}

type UploadURLResponse struct {
	UploadURL string    `json:"upload_url"`
	S3Key     string    `json:"s3_key"`
	S3Bucket  string    `json:"s3_bucket"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ConfirmUploadRequest struct {
	S3Key     string `json:"s3_key" validate:"required"`
	MimeType  string `json:"mime_type" validate:"required"`
	Size      int64  `json:"size" validate:"required,min=1"`
	AssetType string `json:"asset_type" validate:"required,oneof=audio video thumbnail"`
}
