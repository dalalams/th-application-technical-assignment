package validation

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)


func DecodeAndValidate[T any](r *http.Request, v *validator.Validate) (*T, error) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	if err := v.Struct(&req); err != nil {
		return nil, err
	}

	return &req, nil
}
