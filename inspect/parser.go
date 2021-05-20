// Copyright 2021 Matus Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package inspect

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// HTML versions
const (
	Version5   = "5"
	Version401 = "4.01"
	Version4   = "4.0"
	Version32  = "3.2"
	Version3   = "3.0"
	Version2   = "2.0"
	Older      = "<2.0"
)

// InvalidLink represents an inaccessible link
// from the parsed HTML page.
type InvalidLink struct{ URL, Reason string }

// PageContents contains the basic information
// extracted from a HTML page.
type PageContents struct {
	// HTML version used on the page.
	Version string

	// Title of the page.
	Title string

	// Maps Heading to its occurrence count.
	Headings map[string]int

	// Maps domain names to links within that same domain (using set to remove duplicates).
	// Relative URL will be stored under the empty domain "".
	Links map[string]map[string]struct{}

	// If the page contains a login form.
	LoginForm bool
}

// Page extracts general contents from a HTML page.
func Page(page io.Reader) (*PageContents, error) {
	root, err := html.Parse(page)
	if err != nil {
		return nil, err
	}

	p := newPageContents()
	if err := p.traversePage(root); err != nil {
		return nil, err
	}

	return p, nil
}

// newPageContents creates a default initialized *PageContents.
func newPageContents() *PageContents {
	return &PageContents{
		Version:   "",
		Title:     "",
		Headings:  make(map[string]int),
		Links:     make(map[string]map[string]struct{}),
		LoginForm: false,
	}
}

// traversePage traverses the HTML document from the given node.
func (p *PageContents) traversePage(node *html.Node) error {
	switch node.Type {
	case html.DoctypeNode:
		p.extractVersion(node)
	case html.ElementNode:
		if node.Data == "title" {
			if node.FirstChild != nil {
				p.Title = node.FirstChild.Data
			}
		}

		if isHeading(node.Data) {
			p.Headings[node.Data]++
		}

		if node.Data == "a" {
			for i := 0; i < len(node.Attr); i++ {
				if node.Attr[i].Key == "href" {
					// href values can contain:
					// 1. full urls
					// 2. relative urls
					// 3. and starting with #

					u, err := url.Parse(node.Attr[i].Val)
					if err != nil {
						return err
					}

					if p.Links[u.Hostname()] == nil {
						p.Links[u.Hostname()] = make(map[string]struct{})
					}

					// relative ulrs will have an empty hostname
					p.Links[u.Hostname()][node.Attr[i].Val] = struct{}{}
				}
			}
		}

		// we can check for a login form with an <input type="password">
		if node.Data == "input" {
			for i := 0; i < len(node.Attr); i++ {
				if node.Attr[i].Key == "type" && node.Attr[i].Val == "password" {
					// traverse back the tree to check for a form.
					insideForm := false

					for p := node.Parent; p != nil; p = p.Parent {
						if p.Data == "form" {
							insideForm = true
							break
						}
					}

					if insideForm {
						p.LoginForm = true
					}

					break
				}
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if err := p.traversePage(c); err != nil {
			return err
		}
	}

	return nil
}

// extractVersion extract the HTML version from a DocType HTML Node.
func (p *PageContents) extractVersion(node *html.Node) {
	p.Version = Older

	if len(node.Attr) == 0 {
		// <!DOCTYPE html>
		p.Version = Version5
	} else {
		for i := 0; i < len(node.Attr); i++ {
			// <!DOCTYPE html ... HTML 4.01 ...>
			if re := regexp.MustCompile(" (?i)HTML 4.01"); re.MatchString(node.Attr[i].Val) {
				p.Version = Version401
				break
			}

			// <!DOCTYPE html ... HTML 4.0 ...>
			if re := regexp.MustCompile(" (?i)HTML 4.0"); re.MatchString(node.Attr[i].Val) {
				p.Version = Version4
				break
			}

			// <!DOCTYPE html ... HTML 3.2 ...>
			if re := regexp.MustCompile(" (?i)HTML 3.2"); re.MatchString(node.Attr[i].Val) {
				p.Version = Version32
				break
			}

			// <!DOCTYPE html ... HTML 3.0 ...>
			if re := regexp.MustCompile(" (?i)HTML 3.0"); re.MatchString(node.Attr[i].Val) {
				p.Version = Version3
				break
			}

			// <!DOCTYPE html ... HTML 2.0 ...>
			if re := regexp.MustCompile(" (?i)HTML 2.0"); re.MatchString(node.Attr[i].Val) {
				p.Version = Version2
				break
			}
		}
	}
}

// InvalidLinks checks every link extracted from the HTML page if it is
// accessible. For relative links the baseURL parameter will be use to
// create a full URL.
func (p *PageContents) InvalidLinks(baseURL url.URL) map[string][]InvalidLink {
	type pair struct {
		Domain string
		Link   InvalidLink
	}

	var (
		out = make(map[string][]InvalidLink)
		ch  = make(chan pair)

		wg = new(sync.WaitGroup)
	)

	for domain, links := range p.Links {
		wg.Add(1)

		go func(domain string, links map[string]struct{}) {
			defer wg.Done()

			relative := false
			if domain == "" {
				relative = true
				domain = baseURL.Hostname()
			}

			for link := range links {
				if relative { // relative urls
					link = combine(baseURL.String(), link)
				}

				resp, err := http.Get(link)
				if err != nil {
					ch <- pair{
						Domain: domain,
						Link: InvalidLink{
							URL:    link,
							Reason: err.Error(),
						},
					}

					continue
				}

				defer resp.Body.Close()

				// server is having issues and the page is unreachable
				if resp.StatusCode >= 500 && resp.StatusCode < 600 {
					ch <- pair{
						Domain: domain,
						Link: InvalidLink{
							URL:    link,
							Reason: fmt.Sprintf("endpoint responded with code: %v", resp.StatusCode),
						},
					}
				}
			}
		}(domain, links)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for l := range ch {
		out[l.Domain] = append(out[l.Domain], l.Link)
	}

	return out
}

func isHeading(s string) bool {
	return s == "h1" || s == "h2" || s == "h3" || s == "h4" || s == "h5" || s == "h6"
}

func combine(base, relative string) string {
	if strings.HasPrefix(relative, "/") && strings.HasSuffix(base, "/") {
		return base[:len(base)-1] + relative
	}

	return base + relative
}
