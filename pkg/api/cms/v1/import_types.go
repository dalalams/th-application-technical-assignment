package v1

type ImportRequest struct {
	SourceType string `json:"source_type" validate:"required,oneof=youtube spotify rss vimeo"`
	SourceURL  string `json:"source_url" validate:"required"`
	SeriesID   string `json:"series_id" validate:"required,uuid"`
}

