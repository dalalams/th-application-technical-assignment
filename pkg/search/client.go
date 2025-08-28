package search

import (
	"bytes"
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"strings"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/pkg/errors"
)


type OpenSearchClient struct {
	client *opensearch.Client
	config *Config
}

func NewClient(cfg *Config) (*OpenSearchClient, error) {
    if cfg == nil {
        return nil, errors.New("search config is nil")
    }

    slog.Info("config", "cfg", cfg)

	config := opensearch.Config{
		Addresses: []string{cfg.OpenSearchURL},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	if cfg.OpenSearchUsername != "" && cfg.OpenSearchPassword != "" {
		config.Username = cfg.OpenSearchUsername
		config.Password = cfg.OpenSearchPassword
	}

	client, err := opensearch.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create opensearch client")
	}

	return &OpenSearchClient{
		client: client,
		config: cfg,
	}, nil
}

func (c *OpenSearchClient) IndexExists(ctx context.Context, index string) (bool, error) {
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return false, errors.Wrap(err, "failed to check index existence")
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

func (c *OpenSearchClient) CreateIndex(ctx context.Context, index, mapping string) error {
	req := opensearchapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.Wrap(err, "failed to create index")
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("failed to create index: %s", res.String())
	}

	return nil
}

func (c *OpenSearchClient) IndexDocument(ctx context.Context, indexName string, documentID string, documentJSON []byte) error {
	req := opensearchapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       bytes.NewReader(documentJSON),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.Wrap(err, "failed to index document")
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("failed to index document: %s", res.String())
	}

	return nil
}

func (c *OpenSearchClient) DeleteDocument(ctx context.Context, indexName string, documentID string) error {
	req := opensearchapi.DeleteRequest{
		Index:      indexName,
		DocumentID: documentID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.Wrap(err, "failed to delete document")
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return errors.Errorf("failed to delete document: %s", res.String())
	}

	return nil
}

