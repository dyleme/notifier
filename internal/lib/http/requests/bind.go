package requests

import (
	"encoding/json"
	"io"
	"net/http"
)

func Bind(r *http.Request, obj any) error {
	bts, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bts, obj)
	if err != nil {
		return err
	}

	return nil
}
