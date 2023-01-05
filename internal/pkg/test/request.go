// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type Request struct {
	t           *testing.T
	headers     http.Header
	urlQuery    url.Values
	formData    url.Values
	cookies     map[string]string
	httpVerb    string
	contentType string
	endpoint    string
	httpBody    []byte
}

func New(t *testing.T) *Request {
	t.Helper()

	return &Request{
		t:        t,
		urlQuery: url.Values{},
		cookies:  make(map[string]string),
		formData: url.Values{},
		headers:  http.Header{http.CanonicalHeaderKey("user-agent"): {"chii-test-client"}},
	}
}

func (r *Request) newRequest(httpVerb string, endpoint string) *Request {
	r.t.Helper()
	r.httpVerb = httpVerb
	r.endpoint = endpoint

	return r
}

func (r *Request) Get(path string) *Request {
	r.t.Helper()
	return r.newRequest(http.MethodGet, path)
}

func (r *Request) Post(path string) *Request {
	r.t.Helper()
	return r.newRequest(http.MethodPost, path)
}

func (r *Request) Put(path string) *Request {
	r.t.Helper()
	return r.newRequest(http.MethodPut, path)
}

func (r *Request) Patch(path string) *Request {
	r.t.Helper()
	return r.newRequest(http.MethodPatch, path)
}

func (r *Request) Delete(path string) *Request {
	r.t.Helper()
	return r.newRequest(http.MethodDelete, path)
}

func (r *Request) Cookie(key, value string) *Request {
	r.t.Helper()

	r.cookies[key] = value

	return r
}

func (r *Request) Query(key, value string) *Request {
	r.t.Helper()
	r.urlQuery.Set(key, value)
	return r
}

func (r *Request) Header(key, value string) *Request {
	r.t.Helper()
	r.headers.Set(key, value)

	return r
}

func (r *Request) Form(key, value string) *Request {
	r.t.Helper()
	if r.contentType == "" {
		r.contentType = echo.MIMEApplicationForm
	}

	if r.contentType != echo.MIMEApplicationForm {
		r.t.Error("content-type should be empty or 'application/x-www-form-urlencoded'," +
			" can't mix .Form(...) with .JSON(...)")
		r.t.FailNow()
	}

	r.formData.Set(key, value)
	r.httpBody = []byte(r.formData.Encode())

	return r
}

func (r *Request) JSON(v any) *Request {
	r.t.Helper()
	require.Empty(r.t, r.contentType, "content-type should not be empty")

	var err error
	r.httpBody, err = sonic.Marshal(v)
	require.NoError(r.t, err)

	r.contentType = echo.MIMEApplicationJSON

	return r
}

func (r *Request) StdRequest() *http.Request {
	r.t.Helper()
	var body io.ReadCloser = http.NoBody
	if r.httpBody != nil {
		r.headers.Set(echo.HeaderContentLength, strconv.Itoa(len(r.httpBody)))
		if r.headers.Get(echo.HeaderContentType) == "" {
			r.headers.Set(echo.HeaderContentType, r.contentType)
		}

		body = io.NopCloser(bytes.NewBuffer(r.httpBody))
	}

	path := r.endpoint
	if len(r.urlQuery) > 0 {
		var sep = "?"
		if strings.Contains(r.endpoint, "?") {
			sep = "&"
		}

		path = r.endpoint + sep + r.urlQuery.Encode()
	}

	req := httptest.NewRequest(r.httpVerb, path, body)
	req.Header = r.headers
	for name, value := range r.cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	req.RemoteAddr = "0.0.0.0"

	return req
}

func (r *Request) Execute(app *echo.Echo) *Response {
	r.t.Helper()

	resp := httptest.NewRecorder()

	app.ServeHTTP(resp, r.StdRequest())

	return &Response{
		t:          r.t,
		StatusCode: resp.Code,
		Header:     resp.Header(),
		Body:       resp.Body.Bytes(),
		cookies:    parseCookies(r.t, resp.Header().Get(echo.HeaderSetCookie)),
	}
}

func parseCookies(t *testing.T, rawCookies string) []*http.Cookie {
	t.Helper()

	if rawCookies == "" {
		return nil
	}

	r := http.Response{
		Header: http.Header{echo.HeaderSetCookie: {rawCookies}},
	}

	return r.Cookies()
}

type Response struct {
	t          *testing.T
	Header     http.Header
	Body       []byte
	cookies    []*http.Cookie
	StatusCode int
}

func (r *Response) JSON(v any) *Response {
	r.t.Helper()

	if strings.HasPrefix(r.Header.Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
		require.NoError(r.t, sonic.Unmarshal(r.Body, v))
	}

	return r
}

func (r *Response) BodyString() string {
	return string(r.Body)
}

func (r *Response) ExpectCode(t int) *Response {
	r.t.Helper()

	require.Equalf(r.t, t, r.StatusCode, "expecting http response status code %d %s", t, r.BodyString())

	return r
}

func (r *Response) Cookies() []*http.Cookie {
	r.t.Helper()

	return r.cookies
}

type PagedResponse struct {
	Data   json.RawMessage `json:"data"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}
