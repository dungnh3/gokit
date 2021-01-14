package nap

import "net/http"

// Raw is response's raw data
type Raw []byte

// Response is a http response wrapper
type Response struct {
	*http.Response
}

// Response is a http response wrapper
func NewResponse(response *http.Response) *Response {
	return &Response{
		response,
	}
}

// SuccessDecider decide should we decode the response or not
type SuccessDecider func(response *http.Response) bool

// DecodeOnSuccess decide that we should decode on success response (http code 2xx)
func DecodeOnSuccess(resp *http.Response) bool {
	statusCode := resp.StatusCode
	return 200 <= statusCode && statusCode <= 299
}
