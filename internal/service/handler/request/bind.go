package request

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Validator interface {
	Validate() error
}

func Bind(r *http.Request, obj any) error {
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		return fmt.Errorf("bind: %w", err)
	}

	if validator, ok := obj.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}
