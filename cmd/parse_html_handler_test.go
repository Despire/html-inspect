// Copyright 2021 Matus Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func mockExternalServer() *httptest.Server {
	r := http.NewServeMux()

	r.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		out := []byte(`					<!DOCTYPE html>
			<html>
			
			<head>
				<title>Some title</title>
			</head>
			
			<body>
				<div>
					<div>
						<div>
							<div>
								<h1>
									<p> test </p>
								</h1>
								<a href="/some/relative/path/"><span>link 2</span></a>
							</div>
						</div>
						<div>
							<h1>
								<p> test 2</p>
							</h1>
						</div>
						<div>
							<form>
								<input type="text" name="email">
								<input type="password" name="password">
							</form>
						</div>
					</div>
					<div>
						<div>
							<h3>
								<p> test 3</p>
							</h3>
							<a href="https://www.facebook.com"><span>link 8</span></a>
						</div>
					</div>
				</div>
			</body>
			
			</html>`)

		rw.Write(out)
	})

	r.HandleFunc("/some/relative/path/", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	})

	return httptest.NewServer(r)
}

func TestParseHTML(t *testing.T) {
	externalMockServer := mockExternalServer()
	defer externalMockServer.Close()

	mockServer := httptest.NewServer(parseHtml())
	defer mockServer.Close()

	tests := []struct {
		Name           string
		Request        *http.Request
		wantErr        bool
		wantStatusCode int
		wantBody       []byte
	}{
		{
			Name: "fail-invalid-content-type",
			Request: func() *http.Request {
				req, err := http.NewRequest(http.MethodPost, mockServer.URL, nil)
				if err != nil {
					t.Fatal(err)
				}

				return req
			}(),
			wantErr:        false,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       []byte(`{"err":"Invalid content-type"}`),
		},
		{
			Name: "fail-invalid-url",
			Request: func() *http.Request {
				req, err := http.NewRequest(http.MethodPost, mockServer.URL, strings.NewReader(`{"url": "%%2"}`))
				if err != nil {
					t.Fatal(err)
				}

				req.Header.Set("content-type", "application/json")

				return req
			}(),
			wantErr:        false,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       []byte(`{"err":"parse \"%%2\": invalid URL escape \"%%2\""}`),
		},
		{
			Name: "fail-empty",
			Request: func() *http.Request {
				req, err := http.NewRequest(http.MethodPost, mockServer.URL, strings.NewReader(`{"url": ""}`))
				if err != nil {
					t.Fatal(err)
				}

				req.Header.Set("content-type", "application/json")

				return req
			}(),
			wantErr:        false,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       []byte(`{"err":"empty URL in payload"}`),
		},
		{
			Name: "fail-dead-url",
			Request: func() *http.Request {
				req, err := http.NewRequest(http.MethodPost, mockServer.URL, strings.NewReader(`{"url": "http://www.foobar"}`))
				if err != nil {
					t.Fatal(err)
				}

				req.Header.Set("content-type", "application/json")

				return req
			}(),
			wantErr:        false,
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       []byte(`{"err":"Get \"http://www.foobar\": dial tcp: lookup www.foobar: no such host"}`),
		},
		{
			Name: "ok",
			Request: func() *http.Request {
				req, err := http.NewRequest(http.MethodPost, mockServer.URL, strings.NewReader(fmt.Sprintf(`{"url": "%v"}`, externalMockServer.URL)))
				if err != nil {
					t.Fatal(err)
				}

				req.Header.Set("content-type", "application/json")

				return req
			}(),
			wantErr:        false,
			wantStatusCode: http.StatusOK,
			wantBody:       []byte(fmt.Sprintf(`{"version":"5","title":"Some title","login_form":true,"headings":[{"level":"h1","total":2},{"level":"h3","total":1}],"internal":{"domain":"127.0.0.1","links":["%[1]v/some/relative/path/"],"total":1},"external":[{"domain":"www.facebook.com","links":["https://www.facebook.com"],"total":1}],"inaccessible":[{"domain":"127.0.0.1","links":[{"URL":"%[1]v/some/relative/path/","Reason":"endpoint responded with code: 500"}],"total":1}]}`, externalMockServer.URL)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			resp, err := http.DefaultClient.Do(tt.Request)
			if (err != nil) != tt.wantErr {
				t.Errorf("http request for parseHtml() err = %v, want: %v", err, tt.wantErr)
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("parseHtml() status code = %v, want: %v", resp.StatusCode, tt.wantStatusCode)
				return
			}

			b, err := io.ReadAll(resp.Body)
			if (err != nil) != tt.wantErr {
				t.Errorf("io.ReadAll() err = %v, want: %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(string(b), string(tt.wantBody)); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}
