package requests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Bind(r *http.Request, obj any) error {
	op := "Bind: %w"
	bts, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = json.Unmarshal(bts, obj)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	return nil
}
