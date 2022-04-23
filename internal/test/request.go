// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type Request struct {
	t           *testing.T
	headers     http.Header
	Response    interface{}
	urlParams   url.Values
	HTTPVerb    string
	ContentType string
	Endpoint    string
	HTTPBody    []byte
	Cookies     []*http.Cookie
}

func New(t *testing.T) *Request {
	t.Helper()

	return &Request{
		t:         t,
		urlParams: url.Values{},
		headers:   http.Header{},
	}
}

func (r *Request) newRequest(httpVerb string, endpoint string) *Request {
	r.t.Helper()
	r.HTTPVerb = httpVerb
	r.Endpoint = endpoint

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

func (r *Request) Query(key, value string) *Request {
	r.t.Helper()
	r.urlParams.Set(key, value)
	return r
}

func (r *Request) Header(key, value string) *Request {
	r.t.Helper()
	r.headers.Set(key, value)

	return r
}

func (r *Request) JSON(v interface{}) *Request {
	r.t.Helper()
	require.Empty(r.t, r.ContentType, "content-type should not be empty")

	var err error
	r.HTTPBody, err = json.Marshal(v)
	require.NoError(r.t, err)

	r.ContentType = fiber.MIMEApplicationJSON

	return r
}

func (r *Request) StdRequest() *http.Request {
	r.t.Helper()
	var body io.ReadCloser = http.NoBody
	if r.HTTPBody != nil {
		body = io.NopCloser(bytes.NewBuffer(r.HTTPBody))
	}

	path := r.Endpoint
	if len(r.urlParams) > 0 {
		var sep = "?"
		if strings.Contains(r.Endpoint, "?") {
			sep = "&"
		}

		path = r.Endpoint + sep + r.urlParams.Encode()
	}

	req := httptest.NewRequest(r.HTTPVerb, path, body)
	req.Header = r.headers

	return req
}

func (r *Request) Execute(app *fiber.App, msTimeout ...int) *Response {
	r.t.Helper()

	resp, err := app.Test(r.StdRequest(), msTimeout...)
	require.NoError(r.t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(r.t, err)

	return &Response{
		t:          r.t,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
	}
}

type Response struct {
	t          *testing.T
	Header     http.Header
	Body       []byte
	StatusCode int // e.g. 200
}

func (r *Response) JSON(v interface{}) *Response {
	r.t.Helper()

	if strings.HasPrefix(r.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		require.NoError(r.t, json.Unmarshal(r.Body, v))
	}

	return r
}

func (r *Response) BodyString() string {
	return string(r.Body)
}

func (r *Response) ExpectCode(t int) *Response {
	r.t.Helper()

	require.Equalf(r.t, t, r.StatusCode, "expecting http response status code %d", t)

	return r
}
