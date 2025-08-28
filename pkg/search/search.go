package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/pkg/errors"
)

type SearchRequest struct {
	Query    string         `json:"query"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Filters  map[string]any `json:"filters,omitempty"`
}

type SearchResponse struct {
	Total int64            `json:"total"`
	Hits  []map[string]any `json:"hits"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

func (c *OpenSearchClient) SearchSeries(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	index := fmt.Sprintf("%s-series", c.config.IndexPrefix)
	return c.search(ctx, index, req)
}

func (c *OpenSearchClient) SearchEpisodes(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	index := fmt.Sprintf("%s-episodes", c.config.IndexPrefix)
	return c.search(ctx, index, req)
}

func (c *OpenSearchClient) search(ctx context.Context, index string, req SearchRequest) (*SearchResponse, error) {
	query := buildSearchQuery(req)

    from := req.PageSize * (req.Page - 1)

	searchReq := opensearchapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(query),
		From:  &from,
		Size:  &req.PageSize,
	}

	res, err := searchReq.Do(ctx, c.client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute search")
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.Errorf("search failed: %s", res.String())
	}

	var searchResult struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, errors.Wrap(err, "failed to decode search response")
	}

	hits := make([]map[string]any, len(searchResult.Hits.Hits))
	for i, hit := range searchResult.Hits.Hits {
		hits[i] = hit.Source
	}

	return &SearchResponse{
		Total: searchResult.Hits.Total.Value,
		Hits:  hits,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

func buildSearchQuery(req SearchRequest) string {
	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []any{},
			},
		},
		"sort": []any{
			map[string]any{
				"created_at": map[string]any{
					"order": "desc",
				},
			},
		},
	}

	boolQuery := query["query"].(map[string]any)["bool"].(map[string]any)
	must := boolQuery["must"].([]any)

	if req.Query != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  req.Query,
				"fields": []string{"title^2", "description"},
				"type":   "best_fields",
			},
		})
	}

	if req.Filters != nil {
		for field, value := range req.Filters {
			must = append(must, map[string]any{
				"term": map[string]any{
					field: value,
				},
			})
		}
	}

	if len(must) == 0 {
		query["query"] = map[string]any{
			"match_all": map[string]any{},
		}
	} else {
		boolQuery["must"] = must
	}

	queryJSON, _ := json.Marshal(query)
	return string(queryJSON)
}
