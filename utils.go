package gogi

import (
	"bytes"
	"encoding/json"
	"io"
)

func ReaderToStruct[T any](r io.Reader) (T, error) {
	var result T
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&result)
	return result, err
}

func StructToReader[T any](v T) (io.Reader, error) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(v) // Note: Encode adds a trailing newline
	if err != nil {
		return nil, err
	}
	return buf, nil
}
