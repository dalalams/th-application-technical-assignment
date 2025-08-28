package util

import (
	"context"
	"sync"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PaginationRequest struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

type PaginationMetadata struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	ItemCount int64 `json:"item_count"`
	PageCount int   `json:"page_count"`
}

type PaginatedResponse[T any] struct {
	Data       []T                `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

func CalculatePaginationResponse(page, pageSize int, itemCount int64) PaginationMetadata {
	pageCount := max(int((itemCount+int64(pageSize)-1)/int64(pageSize)), 1)

	return PaginationMetadata{
		Page:      page,
		PageSize:  pageSize,
		ItemCount: itemCount,
		PageCount: pageCount,
	}
}

func FetchPaginatedData[T any](
	ctx context.Context,
	fetchCount func(context.Context) (int64, error),
	fetchData func(context.Context) ([]T, error),
) (int64, []T, error) {
	var (
		itemCount  int64
		data       []T
		err        error
		wg         sync.WaitGroup
		errChan    = make(chan error, 2)
		isFinished bool
		mu         sync.Mutex
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		count, fetchErr := fetchCount(ctx)
		if fetchErr != nil {
			errChan <- fetchErr
			return
		}
		itemCount = count
	}()

	go func() {
		defer wg.Done()
		fetchedData, fetchErr := fetchData(ctx)
		if fetchErr != nil {
			errChan <- fetchErr
			return
		}
		data = fetchedData
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for e := range errChan {
		mu.Lock()
		if !isFinished {
			err = e
			isFinished = true
		}
		mu.Unlock()
	}

	return itemCount, data, err
}
