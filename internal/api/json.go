package api

import (
	"encoding/json"
	"io"
	"net/http"
)


func Body[T any](r io.Reader) (*T, error) {
	t := new(T)
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()

	if err := d.Decode(t); err != nil {
		return nil, err
	}
	return t, nil
}

func JSON(w http.ResponseWriter, data any) {
	_ = json.NewEncoder(w).Encode(data)
}
