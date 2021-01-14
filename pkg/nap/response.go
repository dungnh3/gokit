package nap

import "net/http"

type Raw []byte

type Response struct {
	resp *http.Response
}

func NewResponse(resp *http.Response) *Response {
	return &Response{
		resp: resp,
	}
}

type SuccessDecider func(response *http.Response) bool

func DecodeOnSuccess(resp *http.Response) bool {
	return 200 <= resp.StatusCode && resp.StatusCode <= 299
}
