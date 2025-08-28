package validation

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestDecodeAndValidate_GenericType(t *testing.T) {
	t.Parallel()

	type SimpleStruct struct {
		Value string `json:"value" validate:"required"`
	}

	tests := []struct {
		name     string
		jsonData string
		expected interface{}
	}{
		{
			name:     "simple struct",
			jsonData: `{"value": "test"}`,
			expected: SimpleStruct{Value: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := validator.New()

			if simple, ok := tt.expected.(SimpleStruct); ok {
				req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.jsonData))
				req.Header.Set("Content-Type", "application/json")

				result, err := DecodeAndValidate[SimpleStruct](req, v)

				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, simple.Value, result.Value)
			}

		})
	}
}
