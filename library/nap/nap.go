package nap

import (
	"context"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	goquery "github.com/google/go-querystring/query"
)

const (
	// plainTextType   = "text/plain; charset=utf-8"
	ContentTypeJSON = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
)

const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
)

const (
	// hdrUserAgentKey       = "User-Agent"
	// hdrAcceptKey          = "Accept"
	HeaderContentTypeKey = "Content-Type"
	// hdrContentLengthKey   = "Content-Length"
	// hdrContentEncodingKey = "Content-Encoding"
	HeaderAuthorizationKey = "Authorization"
)

// Doer executes http requests.  It is implemented by *http.Client.  You can
// wrap *http.Client with layers of Doers to form a stack of client-side
// middleware.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Nap is an HTTP Request builder and sender.
type Nap struct {
	// http Client for doing requests
	httpClient Doer
	// HTTP method (GET, POST, etc.)
	method string
	// raw url string for requests
	rawURL string
	// stores key-values pairs to add to request's Headers
	header http.Header
	// url tagged query structs
	queryStructs []interface{}
	queryParams  map[string]string
	// body provider
	bodyProvider BodyProvider
	// response decoder
	responseDecoder ResponseDecoder

	ctx       context.Context
	isSuccess SuccessDecider
}

var defaultClient = &http.Client{
	Timeout:   10 * time.Second,
	Transport: nil, // FIXME: if you want to trace your http client request, please implement me
}

// New returns a new Nap with an http defaultClient.
func New() *Nap {
	return &Nap{
		httpClient:      defaultClient,
		method:          MethodGet,
		header:          make(http.Header),
		queryStructs:    make([]interface{}, 0),
		queryParams:     make(map[string]string),
		responseDecoder: jsonDecoder{},
		isSuccess:       DecodeOnSuccess,
	}
}

func (s *Nap) New() *Nap {
	// copy Headers pairs into new Header map
	headerCopy := make(http.Header)
	for k, v := range s.header {
		headerCopy[k] = v
	}
	return &Nap{
		httpClient:      s.httpClient,
		method:          s.method,
		rawURL:          s.rawURL,
		header:          headerCopy,
		queryStructs:    append([]interface{}{}, s.queryStructs...),
		bodyProvider:    s.bodyProvider,
		queryParams:     s.queryParams,
		responseDecoder: s.responseDecoder,
		isSuccess:       s.isSuccess,
	}
}

// Http Client

// var defaultHttpClient = &http.Client{}

// Client sets the http Client used to do requests. If a nil client is given,
// the http.defaultClient will be used.
func (s *Nap) Client(httpClient *http.Client) *Nap {
	if httpClient == nil {
		return s.Doer(defaultClient)
	}
	return s.Doer(httpClient)
}

// Doer sets the custom Doer implementation used to do requests.
// If a nil client is given, the http.defaultClient will be used.
func (s *Nap) Doer(doer Doer) *Nap {
	if doer == nil {
		s.httpClient = defaultClient
	} else {
		s.httpClient = doer
	}
	return s
}

// Context method returns the Context if its already set in request
// otherwise it creates new one using `context.Background()`.
func (s *Nap) Context() context.Context {
	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

//func (s *Nap) AutoRetry(opts ...RetryOption) *Nap {
//	s.httpClient = NewRetryDoer(s.httpClient, opts...)
//	return s
//}

// SetContext method sets the context.Context for current Request. It allows
// to interrupt the request execution if ctx.Done() channel is closed.
// See https://blog.golang.org/context article and the "context" package
// documentation.
func (s *Nap) SetContext(ctx context.Context) *Nap {
	s.ctx = ctx
	return s
}

// Method

// Head sets the Nap method to HEAD and sets the given pathURL.
func (s *Nap) Head(pathURL string) *Nap {
	s.method = MethodHead
	return s.Path(pathURL)
}

// Get sets the Nap method to GET and sets the given pathURL.
func (s *Nap) Get(pathURL string) *Nap {
	s.method = MethodGet
	return s.Path(pathURL)
}

// Post sets the Nap method to POST and sets the given pathURL.
func (s *Nap) Post(pathURL string) *Nap {
	s.method = MethodPost
	return s.Path(pathURL)
}

// Put sets the Nap method to PUT and sets the given pathURL.
func (s *Nap) Put(pathURL string) *Nap {
	s.method = MethodPut
	return s.Path(pathURL)
}

// Patch sets the Nap method to PATCH and sets the given pathURL.
func (s *Nap) Patch(pathURL string) *Nap {
	s.method = MethodPatch
	return s.Path(pathURL)
}

// Delete sets the Nap method to DELETE and sets the given pathURL.
func (s *Nap) Delete(pathURL string) *Nap {
	s.method = MethodDelete
	return s.Path(pathURL)
}

// Options sets the Nap method to OPTIONS and sets the given pathURL.
func (s *Nap) Options(pathURL string) *Nap {
	s.method = MethodOptions
	return s.Path(pathURL)
}

// Header

func (s *Nap) AddHeader(key, value string) *Nap {
	s.header.Add(key, value)
	return s
}

func (s *Nap) SetHeader(key, value string) *Nap {
	s.header.Set(key, value)
	return s
}

func (s *Nap) SetHeaders(headers map[string]string) *Nap {
	for h, v := range headers {
		s.header.Set(h, v)
	}
	return s
}

func (s *Nap) SetBasicAuth(username, password string) *Nap {
	return s.SetHeader(HeaderAuthorizationKey, "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
}

func (s *Nap) SetAuthToken(token string) *Nap {
	return s.SetHeader(HeaderAuthorizationKey, "Bearer "+token)
}

func (s *Nap) WithSuccessDecider(isSuccess SuccessDecider) *Nap {
	s.isSuccess = isSuccess
	return s
}

// Url

// Base sets the rawURL. If you intend to extend the url with Path,
// baseUrl should be specified with a trailing slash.
func (s *Nap) Base(rawURL string) *Nap {
	s.rawURL = rawURL
	return s
}

// Path extends the rawURL with the given path by resolving the reference to
// an absolute URL. If parsing errors occur, the rawURL is left unmodified.
func (s *Nap) Path(path string) *Nap {
	baseURL, baseErr := url.Parse(s.rawURL)
	pathURL, pathErr := url.Parse(path)
	if baseErr == nil && pathErr == nil {
		s.rawURL = baseURL.ResolveReference(pathURL).String()
		if strings.HasSuffix(path, "/") && !strings.HasSuffix(s.rawURL, "/") {
			s.rawURL += "/"
		}
		return s
	}
	return s
}

// QueryStruct appends the queryStruct to the Nap's queryStructs. The value
// pointed to by each queryStruct will be encoded as url query parameters on
// new requests (see Request()).
// The queryStruct argument should be a pointer to a url tagged struct. See
// https://godoc.org/github.com/google/go-querystring/query for details.
func (s *Nap) QueryStruct(queryStruct interface{}) *Nap {
	if queryStruct != nil {
		s.queryStructs = append(s.queryStructs, queryStruct)
	}
	return s
}

func (s *Nap) QueryParams(params map[string]string) *Nap {
	if params != nil {
		s.queryParams = params
	}
	return s
}

// Body

// Body sets the Nap's body. The body value will be set as the Body on new
// requests (see Request()).
// If the provided body is also an io.Closer, the request Body will be closed
// by http.Client methods.
func (s *Nap) Body(body io.Reader) *Nap {
	if body == nil {
		return s
	}
	return s.BodyProvider(bodyProvider{body: body})
}

// BodyProvider sets the Nap's body provider.
func (s *Nap) BodyProvider(body BodyProvider) *Nap {
	if body == nil {
		return s
	}
	s.bodyProvider = body

	ct := body.ContentType()
	if ct != "" {
		s.SetHeader(HeaderContentTypeKey, ct)
	}

	return s
}

// BodyJSON sets the Nap's bodyJSON. The value pointed to by the bodyJSON
// will be JSON encoded as the Body on new requests (see Request()).
// The bodyJSON argument should be a pointer to a JSON tagged struct. See
// https://golang.org/pkg/encoding/json/#MarshalIndent for details.
func (s *Nap) BodyJSON(bodyJSON interface{}) *Nap {
	if bodyJSON == nil {
		return s
	}
	return s.BodyProvider(jsonBodyProvider{payload: bodyJSON})
}

// BodyForm sets the Nap's bodyForm. The value pointed to by the bodyForm
// will be url encoded as the Body on new requests (see Request()).
// The bodyForm argument should be a pointer to a url tagged struct. See
// https://godoc.org/github.com/google/go-querystring/query for details.
func (s *Nap) BodyForm(bodyForm interface{}) *Nap {
	if bodyForm == nil {
		return s
	}
	return s.BodyProvider(formBodyProvider{payload: bodyForm})
}

// Requests

// Request returns a new http.Request created with the Nap properties.
// Returns any errors parsing the rawURL, encoding query structs, encoding
// the body, or creating the http.Request.
func (s *Nap) Request() (*http.Request, error) {
	reqURL, err := url.Parse(s.rawURL)
	if err != nil {
		return nil, err
	}

	err = buildQueryParamUrl(reqURL, s.queryStructs, s.queryParams)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if s.bodyProvider != nil {
		body, err = s.bodyProvider.Body()
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(s.Context(), s.method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}
	addHeaders(req, s.header)
	return req, err
}

// buildQueryParamUrl parses url tagged query structs using go-querystring to
// encode them to url.Values and format them onto the url.RawQuery. Any
// query parsing or encoding errors are returned.
func buildQueryParamUrl(reqURL *url.URL, queryStructs []interface{}, queryParams map[string]string) error {
	urlValues, err := url.ParseQuery(reqURL.RawQuery)
	if err != nil {
		return err
	}
	// encodes query structs into a url.Values map and merges maps
	for _, queryStruct := range queryStructs {
		queryValues, err := goquery.Values(queryStruct)
		if err != nil {
			return err
		}
		for key, values := range queryValues {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
	}
	for k, v := range queryParams {
		urlValues.Add(k, v)
	}
	// url.Values format to a sorted "url encoded" string, e.g. "key=val&foo=bar"
	reqURL.RawQuery = urlValues.Encode()
	return nil
}

// addHeaders adds the key, value pairs from the given http.Header to the
// request. Values for existing keys are appended to the keys values.
func addHeaders(req *http.Request, header http.Header) {
	for key, values := range header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

// Sending

// ResponseDecoder sets the Nap's response decoder.
func (s *Nap) ResponseDecoder(decoder ResponseDecoder) *Nap {
	if decoder == nil {
		return s
	}
	s.responseDecoder = decoder
	return s
}

// ReceiveSuccess creates a new HTTP request and returns the response. Success
// responses (2XX) are JSON decoded into the value pointed to by successV.
// Any error creating the request, sending it, or decoding a 2XX response
// is returned.
func (s *Nap) ReceiveSuccess(successV interface{}) (*Response, error) {
	return s.Receive(successV, nil)
}

// Receive creates a new HTTP request and returns the response. Success
// responses (2XX) are JSON decoded into the value pointed to by successV and
// other responses are JSON decoded into the value pointed to by failureV.
// If the status code of response is 204(no content), decoding is skipped.
// Any error creating the request, sending it, or decoding the response is
// returned.
// Receive is shorthand for calling Request and Do.
func (s *Nap) Receive(successV, failureV interface{}) (*Response, error) {
	req, err := s.Request()
	if err != nil {
		return nil, err
	}
	return s.Do(req, successV, failureV)
}

// Do sends an HTTP request and returns the response. Success responses (2XX)
// are JSON decoded into the value pointed to by successV and other responses
// are JSON decoded into the value pointed to by failureV.
// If the status code of response is 204(no content), decoding is skipped.
// Any error sending the request or decoding the response is returned.
func (s *Nap) Do(req *http.Request, successV, failureV interface{}) (*Response, error) {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return NewResponse(resp), err
	}
	// when err is nil, resp contains a non-nil resp.Body which must be closed
	defer resp.Body.Close()

	// The default HTTP client's Transport may not
	// reuse HTTP/1.x "keep-alive" TCP connections if the Body is
	// not read to completion and closed.
	// See: https://golang.org/pkg/net/http/#Response
	//nolint:errcheck
	defer io.Copy(ioutil.Discard, resp.Body)

	// Don't try to decode on 204s
	if resp.StatusCode == http.StatusNoContent {
		return NewResponse(resp), nil
	}

	// Decode from json
	if successV != nil || failureV != nil {
		err = decodeResponse(resp, s.isSuccess, s.responseDecoder, successV, failureV)
	}
	return NewResponse(resp), err
}

// decodeResponse decodes response Body into the value pointed to by successV
// if the response is a success (2XX) or into the value pointed to by failureV
// otherwise. If the successV or failureV argument to decode into is nil,
// decoding is skipped.
// Caller is responsible for closing the resp.Body.
func decodeResponse(resp *http.Response, isSuccess SuccessDecider, decoder ResponseDecoder, successV, failureV interface{}) error {
	if isSuccess(resp) {
		switch sv := successV.(type) {
		case nil:
			return nil
		case *Raw:
			respBody, err := ioutil.ReadAll(resp.Body)
			*sv = respBody
			return err
		default:
			return decoder.Decode(resp, successV)
		}
	} else {
		switch fv := failureV.(type) {
		case nil:
			return nil
		case *Raw:
			respBody, err := ioutil.ReadAll(resp.Body)
			*fv = respBody
			return err
		default:
			return decoder.Decode(resp, failureV)
		}
	}
}
