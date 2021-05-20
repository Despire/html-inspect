// Copyright 2021 Matus Mrekaj. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package inspect

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/net/html"
)

func TestCombine(t *testing.T) {
	type args struct {
		base     string
		relative string
	}

	tests := []struct {
		Name string
		args args
		want string
	}{
		{
			Name: "ok-1",
			args: args{
				base:     "www.google.com/",
				relative: "#",
			},
			want: "www.google.com/#",
		},
		{
			Name: "ok-2",
			args: args{
				base:     "www.google.com",
				relative: "#",
			},
			want: "www.google.com#",
		},
		{
			Name: "ok-3",
			args: args{
				base:     "www.google.com/",
				relative: "/some/path/",
			},
			want: "www.google.com/some/path/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			have := Combine(tt.args.base, tt.args.relative)

			if diff := cmp.Diff(have, tt.want); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	type args struct {
		node *html.Node
	}
	tests := []struct {
		Name        string
		args        args
		wantVersion string
	}{
		{
			Name: "ok-version5",
			args: args{
				node: &html.Node{
					Type: html.DoctypeNode,
				},
			},
			wantVersion: Version5,
		},
		{
			Name: "ok-version401",
			args: args{
				node: &html.Node{
					Type: html.DoctypeNode,
					Attr: []html.Attribute{
						{
							Val: `PUBLIC "-//W3C//DTD HTML 4.01//EN"`,
						},
					},
				},
			},
			wantVersion: Version4_01,
		},
		{
			Name: "ok-version4",
			args: args{
				node: &html.Node{
					Type: html.DoctypeNode,
					Attr: []html.Attribute{
						{
							Val: `PUBLIC "-//W3C//DTD HTML 4.0//EN"`,
						},
					},
				},
			},
			wantVersion: Version4_0,
		},
		{
			Name: "ok-version32",
			args: args{
				node: &html.Node{
					Type: html.DoctypeNode,
					Attr: []html.Attribute{
						{
							Val: `PUBLIC "-//W3C//DTD HTML 3.2//EN"`,
						},
					},
				},
			},
			wantVersion: Version3_2,
		},
		{
			Name: "ok-version3",
			args: args{
				node: &html.Node{
					Type: html.DoctypeNode,
					Attr: []html.Attribute{
						{
							Val: `PUBLIC "-//W3C//DTD HTML 3.0//EN"`,
						},
					},
				},
			},
			wantVersion: Version3_0,
		},
		{
			Name: "ok-version2",
			args: args{
				node: &html.Node{
					Type: html.DoctypeNode,
					Attr: []html.Attribute{
						{
							Val: `PUBLIC "-//W3C//DTD HTML 2.0//EN"`,
						},
					},
				},
			},
			wantVersion: Version2_0,
		},
		{
			Name: "ok-version-older",
			args: args{
				node: &html.Node{
					Type: html.DoctypeNode,
					Attr: []html.Attribute{
						{
							Val: `PUBLIC "-//W3C//DTD HTML 1.0//EN"`,
						},
					},
				},
			},
			wantVersion: LessThan2_0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			pc := newPageContents()
			pc.extractVersion(tt.args.node)

			if diff := cmp.Diff(pc.Version, tt.wantVersion); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}

func TestTraversePage(t *testing.T) {
	type args struct {
		node *html.Node
	}

	tests := []struct {
		Name         string
		args         args
		wantContents *PageContents
		wantErr      bool
	}{
		{
			Name: "ok-custom-html",
			args: args{
				node: func() *html.Node {
					n, _ := html.Parse(strings.NewReader(`
					<!DOCTYPE html>
<html>

<head>
    <title>Some title</title>
</head>

<body>
    <div>
        <div>
            <a href="#"><span>link 1</span></a>
            <div>
                <div>
                    <h1>
                        <p> test </p>
                    </h1>
                    <a href="/some/relative/path/"><span>link 2</span></a>
                </div>
            </div>
            <div>
                <a href="#test"><span>link 3</span></a>
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
        <a href="/some/relative/path/"><span>link 4</span></a>
        <div>
            <div>
                <a href="https://www.google.com"><span>link 5</span></a>
                <h3>
                    <p> test 3</p>
                </h3>
                <a href="https://www.facebook.com"><span>link 6</span></a>
                <a href="https://www.facebook.com"><span>link 6</span></a>
            </div>
        </div>
    </div>
</body>

</html>
					`))

					return n
				}(),
			},
			wantContents: &PageContents{
				Version: Version5,
				Title:   "Some title",
				Headings: map[string]int{
					"h1": 2,
					"h3": 1,
				},
				Links: map[string]map[string]struct{}{
					// relative links
					"": {
						"#":                    struct{}{},
						"#test":                struct{}{},
						"/some/relative/path/": struct{}{},
					},
					"www.google.com": {
						"https://www.google.com": struct{}{},
					},
					"www.facebook.com": {
						"https://www.facebook.com": struct{}{},
					},
				},
				LoginForm: true,
			},
		},
		{
			Name: "ok-no-header",
			args: args{
				node: func() *html.Node {
					n, _ := html.Parse(strings.NewReader(`
						<!DOCTYPE html>
	<html>
	
	<head>
	</head>
	
	<body>
		<div>
			<div>
				<a href="#"><span>link 1</span></a>
				<div>
					<div>
						<h1>
							<p> test </p>
						</h1>
						<a href="/some/relative/path/"><span>link 2</span></a>
					</div>
				</div>
				<div>
					<a href="#test"><span>link 3</span></a>
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
			<a href="/some/relative/path/"><span>link 4</span></a>
			<div>
				<div>
					<a href="https://www.google.com"><span>link 5</span></a>
					<h3>
						<p> test 3</p>
					</h3>
					<a href="https://www.facebook.com"><span>link 6</span></a>
					<a href="https://www.facebook.com"><span>link 6</span></a>
				</div>
			</div>
		</div>
	</body>
	
	</html>
						`))

					return n
				}(),
			},
			wantContents: &PageContents{
				Version: Version5,
				Title:   "",
				Headings: map[string]int{
					"h1": 2,
					"h3": 1,
				},
				Links: map[string]map[string]struct{}{
					// relative links
					"": {
						"#":                    struct{}{},
						"#test":                struct{}{},
						"/some/relative/path/": struct{}{},
					},
					"www.google.com": {
						"https://www.google.com": struct{}{},
					},
					"www.facebook.com": {
						"https://www.facebook.com": struct{}{},
					},
				},
				LoginForm: true,
			},
		},
		{
			Name: "ok-no-header&form",
			args: args{
				node: func() *html.Node {
					n, _ := html.Parse(strings.NewReader(`
						<!DOCTYPE html>
	<html>
	
	<head>
	</head>
	
	<body>
		<div>
			<div>
				<a href="#"><span>link 1</span></a>
				<div>
					<div>
						<h1>
							<p> test </p>
						</h1>
						<a href="/some/relative/path/"><span>link 2</span></a>
					</div>
				</div>
				<div>
					<a href="#test"><span>link 3</span></a>
					<h1>
						<p> test 2</p>
					</h1>
				</div>
				<div>
				</div>
			</div>
			<a href="/some/relative/path/"><span>link 4</span></a>
			<div>
				<div>
					<a href="https://www.google.com"><span>link 5</span></a>
					<h3>
						<p> test 3</p>
					</h3>
					<a href="https://www.facebook.com"><span>link 6</span></a>
					<a href="https://www.facebook.com"><span>link 6</span></a>
				</div>
			</div>
		</div>
	</body>
	
	</html>
						`))

					return n
				}(),
			},
			wantContents: &PageContents{
				Version: Version5,
				Title:   "",
				Headings: map[string]int{
					"h1": 2,
					"h3": 1,
				},
				Links: map[string]map[string]struct{}{
					// relative links
					"": {
						"#":                    struct{}{},
						"#test":                struct{}{},
						"/some/relative/path/": struct{}{},
					},
					"www.google.com": {
						"https://www.google.com": struct{}{},
					},
					"www.facebook.com": {
						"https://www.facebook.com": struct{}{},
					},
				},
				LoginForm: false,
			},
		},
		{
			Name: "ok-no-header&form&headings",
			args: args{
				node: func() *html.Node {
					n, _ := html.Parse(strings.NewReader(`
							<!DOCTYPE html>
		<html>
		
		<head>
		</head>
		
		<body>
			<div>
				<div>
					<a href="#"><span>link 1</span></a>
					<div>
						<div>
							<a href="/some/relative/path/"><span>link 2</span></a>
						</div>
					</div>
					<div>
						<a href="#test"><span>link 3</span></a>
					</div>
					<div>
					</div>
				</div>
				<a href="/some/relative/path/"><span>link 4</span></a>
				<div>
					<div>
						<a href="https://www.google.com"><span>link 5</span></a>
						<a href="https://www.facebook.com"><span>link 6</span></a>
						<a href="https://www.facebook.com"><span>link 6</span></a>
					</div>
				</div>
			</div>
		</body>
		
		</html>
							`))

					return n
				}(),
			},
			wantContents: &PageContents{
				Version:  Version5,
				Title:    "",
				Headings: map[string]int{},
				Links: map[string]map[string]struct{}{
					// relative links
					"": {
						"#":                    struct{}{},
						"#test":                struct{}{},
						"/some/relative/path/": struct{}{},
					},
					"www.google.com": {
						"https://www.google.com": struct{}{},
					},
					"www.facebook.com": {
						"https://www.facebook.com": struct{}{},
					},
				},
				LoginForm: false,
			},
		},
		{
			Name: "ok-no-header&form&headings&links",
			args: args{
				node: func() *html.Node {
					n, _ := html.Parse(strings.NewReader(`
							<!DOCTYPE html>
		<html>
		
		<head>
		</head>
		
		<body>
			<div>
				<div>
					<div>
						<div>
						</div>
					</div>
					<div>
					</div>
					<div>
					</div>
				</div>
				<div>
					<div>
					</div>
				</div>
			</div>
		</body>
		
		</html>
							`))

					return n
				}(),
			},
			wantContents: &PageContents{
				Version:   Version5,
				Title:     "",
				Headings:  map[string]int{},
				Links:     map[string]map[string]struct{}{},
				LoginForm: false,
			},
		},
		{
			Name: "fail-invalid-link",
			args: args{
				node: func() *html.Node {
					n, _ := html.Parse(strings.NewReader(`
							<!DOCTYPE html>
		<html>
		
		<head>
		</head>
		
		<body>
			<div>
				<div>
				<a href="%%2"></a>
					<div>
						<div>
						</div>
					</div>
					<div>
					</div>
					<div>
					</div>
				</div>
				<div>
					<div>
					</div>
				</div>
			</div>
		</body>
		
		</html>
							`))

					return n
				}(),
			},
			wantContents: &PageContents{
				Version:  Version5,
				Headings: map[string]int{},
				Links:    map[string]map[string]struct{}{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			pc := newPageContents()

			if err := pc.traversePage(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("traversePage() err = %v, want %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(pc, tt.wantContents); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}

func TestPage(t *testing.T) {
	tests := []struct {
		Name         string
		in           io.Reader
		wantContents *PageContents
		wantErr      bool
	}{
		{
			Name: "ok",
			in: strings.NewReader(`
			<!DOCTYPE html>
			<html>
			
			<head>
				<title>Some title</title>
			</head>
			
			<body>
				<div>
					<div>
						<a href="#"><span>link 1</span></a>
						<div>
							<div>
								<h1>
									<p> test </p>
								</h1>
								<a href="/some/relative/path/"><span>link 2</span></a>
							</div>
						</div>
						<div>
							<a href="#test"><span>link 3</span></a>
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
					<a href="/some/relative/path/"><span>link 4</span></a>
					<div>
						<div>
							<a href="https://www.google.com"><span>link 5</span></a>
							<h3>
								<p> test 3</p>
							</h3>
							<a href="https://www.facebook.com"><span>link 6</span></a>
							<a href="https://www.facebook.com"><span>link 6</span></a>
						</div>
					</div>
				</div>
			</body>
			
			</html>`),
			wantContents: &PageContents{
				Version: Version5,
				Title:   "Some title",
				Headings: map[string]int{
					"h1": 2,
					"h3": 1,
				},
				Links: map[string]map[string]struct{}{
					// relative links
					"": {
						"#":                    struct{}{},
						"#test":                struct{}{},
						"/some/relative/path/": struct{}{},
					},
					"www.google.com": {
						"https://www.google.com": struct{}{},
					},
					"www.facebook.com": {
						"https://www.facebook.com": struct{}{},
					},
				},
				LoginForm: true,
			},
		},
		{
			Name: "fail-invalid-link",
			in: strings.NewReader(`
							<!DOCTYPE html>
		<html>
		
		<head>
		</head>
		
		<body>
			<div>
				<div>
				<a href="%%2"></a>
					<div>
						<div>
						</div>
					</div>
					<div>
					</div>
					<div>
					</div>
				</div>
				<div>
					<div>
					</div>
				</div>
			</div>
		</body>
		
		</html>`),
			wantContents: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			have, err := Page(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Page() err = %v, want = %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(have, tt.wantContents); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}

func TestInvalidLinks(t *testing.T) {
	mustParse := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}

		return u
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}))

	defer mockServer.Close()

	type args struct {
		url *url.URL
	}
	tests := []struct {
		Name     string
		args     args
		Contents *PageContents
		Want     map[string][]InvalidLink
	}{
		{
			Name: "ok-empty",
			args: args{
				url: mustParse(""),
			},
			Contents: new(PageContents),
			Want:     map[string][]InvalidLink{},
		},
		{
			Name: "ok-unsuported protocol scheme",
			args: args{
				url: mustParse("www.foobar"),
			},
			Contents: &PageContents{
				Links: map[string]map[string]struct{}{
					"": {
						"/relative": struct{}{},
					},
				},
			},

			Want: map[string][]InvalidLink{
				"": {
					{
						URL:    "www.foobar/relative",
						Reason: `Get "www.foobar/relative": unsupported protocol scheme ""`,
					},
				},
			},
		},
		{
			Name: "ok-no-such-host",
			args: args{
				url: mustParse("http://www.foobar"),
			},
			Contents: &PageContents{
				Links: map[string]map[string]struct{}{
					"": {
						"/relative": struct{}{},
					},
				},
			},

			Want: map[string][]InvalidLink{
				"www.foobar": {
					{
						URL:    "http://www.foobar/relative",
						Reason: "Get \"http://www.foobar/relative\": dial tcp: lookup www.foobar: no such host",
					},
				},
			},
		},
		{
			Name: "ok-server-error",
			args: args{
				url: mustParse(mockServer.URL),
			},
			Contents: &PageContents{
				Links: map[string]map[string]struct{}{
					mustParse(mockServer.URL).Hostname(): {
						mockServer.URL: struct{}{},
					},
				},
			},

			Want: map[string][]InvalidLink{
				mustParse(mockServer.URL).Hostname(): {
					{
						URL:    mockServer.URL,
						Reason: "endpoint responded with code: 500",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			have := tt.Contents.InvalidLinks(*tt.args.url)

			if diff := cmp.Diff(have, tt.Want); diff != "" {
				t.Error(diff)
				return
			}
		})
	}
}
