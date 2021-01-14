package nap

import (
	"encoding/json"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"net/http"
)

// ResponseDecoder decodes http responses into struct values
type ResponseDecoder interface {
	Decode(resp *http.Response, v interface{}) error
}

// jsonDecoder decodes http response JSON into a JSON-tagged struct value
type jsonDecoder struct {
}

// Decode decodes the Response Body into the value pointed to by v
// Caller must provide a non-nil v and close the resp.Body
func (d jsonDecoder) Decode(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// jsonDecoder decodes http response JSON into a JSON-tagged struct value
type JsonPbDecoder struct {
}

// Decode decodes the Response Body into the value pointed to by v
// Caller must provide a non-nil v and close the resp.Body
func (d JsonPbDecoder) Decode(resp *http.Response, v interface{}) error {
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	return unmarshaler.Unmarshal(resp.Body, v.(proto.Message))
}
