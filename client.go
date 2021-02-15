package lifx

import (
	//"crypto/tls"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultUserAgent = "lifx-go"

type (
	Client struct {
		accessToken string
		userAgent   string
		Client      *http.Client
	}

	Result struct {
		Id     string `json:"id"`
		Label  string `json:"label"`
		Status Status `json:"status"`
	}

	Error struct {
		Field   string   `json:"field"`
		Message []string `json:"message"`
	}

	Warning struct {
		Warning string `json:"warning"`
	}

	RateLimit struct {
		Limit     int
		Remaining int
		Reset     time.Time
	}

	Response struct {
		StatusCode int
		Header     http.Header
		Body       io.ReadCloser
		RateLimit  RateLimit
	}

	LifxResponse struct {
		Error    string    `json:"error"`
		Errors   []Error   `json:"errors"`
		Warnings []Warning `json:"warnings"`
		Results  []Result  `json:"results"`
	}
)

var errorMap = map[int]error{
	http.StatusNotFound:            errors.New("Selector did not match any lights"),
	http.StatusUnauthorized:        errors.New("Bad access token"),
	http.StatusForbidden:           errors.New("Bad OAuth scope"),
	http.StatusUnprocessableEntity: errors.New("Missing or malformed parameters"),
	http.StatusUpgradeRequired:     errors.New("HTTP was used to make the request instead of HTTPS. Repeat the request using HTTPS instead"),
	http.StatusTooManyRequests:     errors.New("The request exceeded a rate limit"),
	http.StatusInternalServerError: errors.New("Something went wrong on LIFX's end"),
	http.StatusBadGateway:          errors.New("Something went wrong on LIFX's end"),
	http.StatusServiceUnavailable:  errors.New("Something went wrong on LIFX's end"),
	523:                            errors.New("Something went wrong on LIFX's end"),
}

func NewClient(accessToken string, options ...func(*Client)) *Client {
	var c *Client
	tr := &http.Transport{
		//TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	c = &Client{
		accessToken: accessToken,
		Client:      &http.Client{Transport: tr},
	}

	for _, option := range options {
		option(c)
	}

	return c
}

func WithUserAgent(userAgent string) func(*Client) {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

func NewClientWithUserAgent(accessToken string, userAgent string) *Client {
	tr := &http.Transport{
		//TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	return &Client{
		accessToken: accessToken,
		userAgent:   userAgent,
		Client:      &http.Client{Transport: tr},
	}
}

func NewResponse(r *http.Response) (*Response, error) {
	resp := Response{
		StatusCode: r.StatusCode,
		Header:     r.Header,
		Body:       r.Body,
	}

	if t := r.Header.Get("X-RateLimit-Limit"); t != "" {
		if n, err := strconv.ParseInt(t, 10, 32); err == nil {
			resp.RateLimit.Limit = int(n)
		} else {
			return nil, err
		}
	}

	if t := r.Header.Get("X-RateLimit-Remaining"); t != "" {
		if n, err := strconv.ParseInt(t, 10, 32); err == nil {
			resp.RateLimit.Remaining = int(n)
		} else {
			return nil, err
		}
	}

	if t := r.Header.Get("X-RateLimit-Reset"); t != "" {
		if n, err := strconv.ParseInt(t, 10, 32); err == nil {
			resp.RateLimit.Reset = time.Unix(n, 0)
		} else {
			return nil, err
		}
	}

	return &resp, nil
}

func (r *Response) IsError() bool {
	return r.StatusCode > 299
}

func (r *Response) GetLifxError() (err error) {
	var (
		s *LifxResponse
	)
	if err = json.NewDecoder(r.Body).Decode(&s); err != nil {
		return nil
	}
	return errors.New(s.Error)
}

func (c *Client) NewRequest(method, url string, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", c.userAgent)
	return
}

func (c *Client) setState(selector string, state State) (*Response, error) {
	var (
		err  error
		j    []byte
		req  *http.Request
		r    *http.Response
		resp *Response
	)

	if j, err = json.Marshal(state); err != nil {
		return nil, err
	}

	if req, err = c.NewRequest("PUT", EndpointState(selector), bytes.NewBuffer(j)); err != nil {
		return nil, err
	}

	if r, err = c.Client.Do(req); err != nil {
		return nil, err
	}

	resp, err = NewResponse(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) setStates(selector string, states States) (*Response, error) {
	var (
		err  error
		j    []byte
		req  *http.Request
		r    *http.Response
		resp *Response
	)

	if j, err = json.Marshal(states); err != nil {
		return nil, err
	}

	if req, err = c.NewRequest("PUT", EndpointStates(), bytes.NewBuffer(j)); err != nil {
		return nil, err
	}

	if r, err = c.Client.Do(req); err != nil {
		return nil, err
	}

	resp, err = NewResponse(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) toggle(selector string, duration float64) (*Response, error) {
	var (
		err  error
		j    []byte
		req  *http.Request
		r    *http.Response
		resp *Response
	)

	if j, err = json.Marshal(&Toggle{Duration: duration}); err != nil {
		return nil, err
	}

	if req, err = c.NewRequest("POST", EndpointToggle(selector), bytes.NewBuffer(j)); err != nil {
		return nil, err
	}

	if r, err = c.Client.Do(req); err != nil {
		return nil, err
	}

	resp, err = NewResponse(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) validateColor(color Color) (*Response, error) {
	var (
		err  error
		req  *http.Request
		r    *http.Response
		resp *Response
		q    url.Values
	)

	if req, err = c.NewRequest("GET", EndpointColor(), nil); err != nil {
		return nil, err
	}

	q = req.URL.Query()
	q.Set("string", color.ColorString())
	req.URL.RawQuery = q.Encode()

	if r, err = c.Client.Do(req); err != nil {
		return nil, err
	}

	resp, err = NewResponse(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) listLights(selector string) (*Response, error) {
	var (
		err  error
		req  *http.Request
		r    *http.Response
		resp *Response
	)

	if req, err = c.NewRequest("GET", EndpointListLights(selector), nil); err != nil {
		return nil, err
	}

	if r, err = c.Client.Do(req); err != nil {
		return nil, err
	}

	resp, err = NewResponse(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) stateDelta(selector string, delta StateDelta) (*Response, error) {
	var (
		err  error
		j    []byte
		req  *http.Request
		r    *http.Response
		resp *Response
	)

	if j, err = json.Marshal(delta); err != nil {
		return nil, err
	}

	if req, err = c.NewRequest("POST", EndpointStateDelta(selector), bytes.NewBuffer(j)); err != nil {
		return nil, err
	}

	if r, err = c.Client.Do(req); err != nil {
		return nil, err
	}

	resp, err = NewResponse(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
