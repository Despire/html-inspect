// Copyright 2021 Matus Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Despire/htmlinspect/inspect"
)

type (
	Heading struct {
		Level string `json:"level"`
		Total int    `json:"total"`
	}

	Link struct {
		Domain string   `json:"domain"`
		Links  []string `json:"links"`
		Total  int      `json:"total"`
	}

	InvalidLink struct {
		Domain string                `json:"domain"`
		Links  []inspect.InvalidLink `json:"links"`
		Total  int                   `json:"total"`
	}
)

type ParseHTMLRequest struct {
	URL string `json:"url"`
}

type ParseHTMLResponse struct {
	Version   string `json:"version"`
	Title     string `json:"title"`
	LoginForm bool   `json:"login_form"`

	Headings     []Heading     `json:"headings"`
	Internal     *Link         `json:"internal"`
	External     []Link        `json:"external"`
	Inaccessible []InvalidLink `json:"inaccessible"`
}

func parseHtml() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("content-type") != "application/json" {
			JSONError(w, "Invalid content-type", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read request body: %v\n", err)
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		payload := ParseHTMLRequest{}
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("failed to unmarshal request body: %v\n", err)
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if payload.URL == "" {
			log.Printf("empty URL in payload")
			JSONError(w, "empty URL in payload", http.StatusBadRequest)
			return
		}

		u, err := url.Parse(payload.URL)
		if err != nil {
			log.Printf("failed to parse url")
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := http.Get(u.String())
		if err != nil {
			log.Printf("failed to fetch page for url:%v", payload.URL)
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("failed to read response body: %v\n", err)
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		contents, err := inspect.Page(bytes.NewReader(body))
		if err != nil {
			log.Printf("failed to extract page contents: %v", err)
			JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		out := ParseHTMLResponse{
			Version:   contents.Version,
			LoginForm: contents.LoginForm,
		}

		if contents.Title == "" {
			contents.Title = u.String() // if there was no title element default to the url of the page.
		}

		out.Title = contents.Title

		for level, count := range contents.Headings {
			out.Headings = append(out.Headings, Heading{
				Level: level,
				Total: count,
			})
		}

		for domain, links := range contents.InvalidLinks(*u) {
			out.Inaccessible = append(out.Inaccessible, InvalidLink{
				Domain: domain,
				Links:  links,
				Total:  len(links),
			})
		}

		if internal := consumeInternalLinks(u, contents.Links); len(internal) > 0 {
			out.Internal = &Link{
				Domain: u.Hostname(),
				Links:  internal,
				Total:  len(internal),
			}
		}

		// rest of the links are all external.
		for domain, links := range contents.Links {
			link := Link{Domain: domain}

			for l := range links {
				link.Links = append(link.Links, l)
			}

			link.Total = len(link.Links)
			out.External = append(out.External, link)
		}

		JSON(w, &out, http.StatusOK)
	}
}

// consumeInternalLinks extracts relative links and links for the urls domain.
// This function will delete the links from the map.
func consumeInternalLinks(u *url.URL, links map[string]map[string]struct{}) []string {
	var out []string

	// relative links are internal.
	for l := range links[""] {
		if u := u.String(); strings.HasSuffix(u, "/") && strings.HasPrefix(l, "/") {
			out = append(out, u[:len(u)-1]+l)
		} else {
			out = append(out, u+l)
		}
	}

	// check also for links using the full URL.
	for l := range links[u.Hostname()] {
		out = append(out, l)
	}

	delete(links, "")
	delete(links, u.Hostname())

	return out
}

// JSON marshals the payload and writes it to the output.
func JSON(out http.ResponseWriter, payload interface{}, status int) {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	out.Header().Set("content-type", "application/json")
	out.WriteHeader(status)

	if _, err := out.Write(b); err != nil {
		panic(err)
	}
}

// JSONError marhals the err msg and sends it to the output.
func JSONError(out http.ResponseWriter, msg string, status int) {
	out.Header().Set("content-type", "application/json")
	out.WriteHeader(status)

	s := struct {
		Err string `json:"err"`
	}{Err: msg}

	b, err := json.Marshal(&s)
	if err != nil {
		panic(err)
	}

	if _, err := out.Write(b); err != nil {
		panic(err)
	}
}
