package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Bind(r *http.Request, obj any) error {
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		return fmt.Errorf("bind: %w", err)
	}

	return nil
}
