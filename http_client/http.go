package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dungnh3/gokit/log/level"
	"github.com/jinzhu/copier"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	ContentTypeJson = "application/json"
)

// Option ..
type Option struct {
	Header map[string]string
	// TimeoutMs is a number of time before TCP connection stop (unit: miliseconds)
	TimeoutMs int
}

// Request store data needed to do a http request
type Request struct {
	Path       string
	RawQuery   string
	Payload    interface{}
	RawPayload []byte
	Headers    http.Header
}

// Response store extra metada of the response
type Response struct {
	StatusCode int
	RawBody    []byte
}

// Unmarshal decode json rawBody into data struct
func (r *Response) Unmarshal(result interface{}) error {
	return json.Unmarshal(r.RawBody, result)
}

// HttpClient implement basic Http client behavior
type HttpClient interface {
	Get(ctx context.Context, request Request) (*Response, error)
	Post(ctx context.Context, request Request) (*Response, error)
	Put(ctx context.Context, request Request) (*Response, error)
}

// HttpFancyClient is an implement of HttpClient
type HttpFancyClient struct {
	client *http.Client
	url    *url.URL
}

// NewHttpFancyClient return real implemented HttpClient
func NewHttpFancyClient(host string, option *Option) (HttpClient, error) {
	if option == nil {
		option = &Option{
			TimeoutMs: 5000, // default time out of request
		}
	}

	parseUrl, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	return &HttpFancyClient{
		client: &http.Client{
			Timeout: time.Duration(option.TimeoutMs) * time.Millisecond,
		},
		url: parseUrl,
	}, nil
}

// Get ..
func (c *HttpFancyClient) Get(ctx context.Context, request Request) (*Response, error) {
	return c.doAction(ctx, http.MethodGet, request)
}

// Post ..
func (c *HttpFancyClient) Post(ctx context.Context, request Request) (*Response, error) {
	return c.doAction(ctx, http.MethodPost, request)
}

// Put ..
func (c *HttpFancyClient) Put(ctx context.Context, request Request) (*Response, error) {
	return c.doAction(ctx, http.MethodPut, request)
}

// scopedUrl ..
func (c *HttpFancyClient) scopedUrl(req Request) (*url.URL, error) {
	var res url.URL
	err := copier.Copy(&res, c.url)
	if err != nil {
		return nil, err
	}
	res.Path += req.Path
	res.RawQuery = req.RawQuery
	return &res, nil
}

// scopedRequest ..
func (c *HttpFancyClient) scopedRequest(method string, req Request) (*http.Request, error) {
	var reqBody *bytes.Reader
	if len(req.RawPayload) > 0 {
		reqBody = bytes.NewReader(req.RawPayload)
	} else if req.Payload != nil {
		requestBody, err := json.Marshal(req.Payload)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(requestBody)
	} else {
		reqBody = bytes.NewReader(make([]byte, 0))
	}

	scopedUrl, err := c.scopedUrl(req)
	if err != nil {
		return nil, err
	}
	scopeUrlStr := scopedUrl.String()
	request, err := http.NewRequest(method, scopeUrlStr, reqBody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", ContentTypeJson)
	for key, header := range req.Headers {
		for _, h := range header {
			request.Header.Set(key, h)
		}
	}
	return request, nil
}

func (c *HttpFancyClient) doAction(ctx context.Context, method string, req Request) (*Response, error) {
	scopedRequest, err := c.scopedRequest(method, req)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(scopedRequest)
	if err != nil {
		level.Error(ctx).F("%v %v failed: %v", method, req, err)
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			level.Warn(ctx).F("error while closing conncection, error: %v", err)
		}
	}()

	// parse response http to Response struct
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Response{
		StatusCode: resp.StatusCode,
		RawBody:    raw,
	}, nil
}
