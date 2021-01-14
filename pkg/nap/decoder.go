package nap

import (
	"encoding/json"
	"net/http"
)

/**
 * ResponseDecoder decodes http responses into struct assign
 */
type ResponseDecoder interface {
	// Decode will decode the response into the value pointed to by v
	Decode(resp *http.Response, v interface{}) error
}

/**
 * jsonDecoder is an implementation of ResponseDecoder
 * it decodes http response JSON into a JSON-tagged struct value
 */
type jsonDecoder struct {
}

func (d jsonDecoder) Decode(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}
