package nap

import (
	"context"
	"net/http"
)

// Define http method constant
const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
)

// Define content-type constant
const (
	ContentTypeJson = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
)

/**
 * Doer executes http requests. It is implemented by *http.Client. You can
 * wrap *http.Client with layers of Doer to form a stack of client-side
 * middleware
 */
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

/**
 * Nap is an implementation HTTP Request builder and sender
 */
type Nap struct {
	// httpClient for doing requests
	httpClient Doer
	// httpMethod (GET, POST, PUT, DELETE,..)
	httpMethod string
	// rawUrl string for requests
	rawUrl string
	// header stores key-values pairs to add to request's header
	header http.Header

	queryStructs    []interface{}
	queryParams     map[string]interface{}
	bodyProvider    BodyProvider
	responseDecoder ResponseDecoder

	ctx       context.Context
	isSuccess SuccessDecider
}

var defaultClient = &http.Client{
}

// New return a new Nap with an http defaultClient
func New() *Nap {
	return &Nap{
		httpClient:      defaultClient,
		httpMethod:      MethodGet,
		header:          make(http.Header),
		queryStructs:    make([]interface{}, 0),
		queryParams:     make(map[string]interface{}),
		responseDecoder: jsonDecoder{},
		isSuccess:       DecodeOnSuccess,
	}
}
