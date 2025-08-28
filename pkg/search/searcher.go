package search

import (
	"context"
)

type Searcher interface {
	SearchSeries(ctx context.Context, req SearchRequest) (*SearchResponse, error)
	SearchEpisodes(ctx context.Context, req SearchRequest) (*SearchResponse, error)
	IndexDocument(ctx context.Context, indexName string, documentID string, documentJSON []byte) error
	DeleteDocument(ctx context.Context, indexName string, documentID string) error
	IndexExists(ctx context.Context, indexName string) (bool, error)
	CreateIndex(ctx context.Context, indexName string, mapping string) error
}
