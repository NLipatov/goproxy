package http_objects

import (
	"encoding/json"
	"net/http"
)

type JSONResponse[T any] struct {
	w http.ResponseWriter
}

func NewJSONResponse[T any](w http.ResponseWriter) JSONResponse[T] {
	return JSONResponse[T]{
		w: w,
	}
}

func (j JSONResponse[T]) Respond(response T) error {
	responseBytes, responseBytesErr := json.Marshal(response)
	if responseBytesErr != nil {
		return responseBytesErr
	}

	_, err := j.w.Write(responseBytes)
	if err != nil {
		return err
	}

	return nil
}
