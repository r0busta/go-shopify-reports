package utils

import (
	"encoding/json"
	"io"
)

func ParseJSON(r io.Reader, v interface{}) error {
	err := json.NewDecoder(r).Decode(v)
	if err != io.EOF {
		return err
	}
	return nil
}

func WriteJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func WriteFormatedJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
