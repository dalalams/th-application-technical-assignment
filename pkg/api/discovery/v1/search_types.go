package v1

type SearchSeriesRequest struct {
	Query      string  `json:"query,omitempty"`
	Page       int     `json:"page" validate:"min=1"`
	PageSize   int     `json:"page_size" validate:"min=1,max=100"`
	CategoryID *string `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Type       *string `json:"type,omitempty" validate:"omitempty,oneof=documentary podcast blog"`
	Language   *string `json:"language,omitempty"`
}

type SearchEpisodesRequest struct {
	Query    string  `json:"query,omitempty"`
	Page     int     `json:"page" validate:"min=1"`
	PageSize int     `json:"page_size" validate:"min=1,max=100"`
	SeriesID *string `json:"series_id,omitempty" validate:"omitempty,uuid"`
}

type SearchResponse struct {
	Query     string           `json:"query"`
	Total     int64            `json:"total"`
	Page      int              `json:"page"`
	PageSize  int              `json:"page_size"`
	PageCount int              `json:"page_count"`
	Results   []map[string]any `json:"results"`
}
