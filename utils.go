package gogi

import (
	"bytes"
	"encoding/json"
	"io"
)

func ReaderToJson(r io.Reader, dest interface{}) error {
	if r == nil {
		return io.EOF
	}

	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	r = bytes.NewBuffer(bodyBytes)

	return json.Unmarshal(bodyBytes, &dest)
}

func JsonToReader(v interface{}) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func FromJSON[T any](r io.Reader) (T, error) {
	var result T
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&result)
	return result, err
}
