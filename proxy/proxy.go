// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package access provides an access logging handler for the ozzo routing package.
package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	"github.com/ltick/tick-routing"
)

var (
	errMatchProxy = "tick-routing: match proxy error"
)

type Proxy struct {
	MethodRule     string
	HostRule       string
	UriRule        string
	UpstreamURL    *url.URL
	UpstreamHeader *http.Header
}

func (p *Proxy) MatchProxy(r *http.Request) bool {
	// Host
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}
	var hostMatch, methodMatch, uriMatch bool
	if strings.Compare(host, p.HostRule) == 0 {
		hostMatch = true
	} else if regexp.MustCompile(p.HostRule).Match([]byte(host)) {
		hostMatch = true
	}
	if hostMatch {
		// Method
		if strings.Compare(r.Method, p.MethodRule) == 0 {
			methodMatch = true
		} else if regexp.MustCompile(p.MethodRule).Match([]byte(r.Method)) {
			methodMatch = true
		}
		if methodMatch {
			// Uri
			if strings.Compare(r.URL.RequestURI(), p.UriRule) == 0 {
				uriMatch = true
			} else if regexp.MustCompile(p.UriRule).Match([]byte(r.URL.RequestURI())) {
				uriMatch = true
			}
			if uriMatch {
				return true
			}
		}
	}
	return false
}

// ProxyHandler returns a handler that proxy config for every request.
// all request will pass through is match rule, if matched it will proxy request to upstream address.
//
//     import (
//         "log"
//         "github.com/ltick/tick-routing"
//         "github.com/ltick/tick-routing/proxy"
//         "net/http"
//     )
//
//     proxys = []*Proxy{
//         &Proxy{Rule:"http://www.example.com/\w+.html/", Upstream:"http://127.0.0.1:9000/index.html"}
//     }
//     r := routing.New()
//     r.Use(access.Proxy(proxys))
func ProxyHandler(proxys []*Proxy) routing.Handler {
	return func(c *routing.Context) error {
		for _, p := range proxys {
			match  := p.MatchProxy(c.Request)
			if match {
				director := func(req *http.Request) {
					req.URL = p.UpstreamURL
					req.Header = *p.UpstreamHeader
				}
				proxy := &httputil.ReverseProxy{Director: director}
				proxy.ServeHTTP(c.ResponseWriter, c.Request)
				c.Abort()
				return nil
			}
		}
		return nil
	}
}

func HTTPProxyHandler(proxys []*Proxy) routing.Handler {
	return func(c *routing.Context) error {
		for _, p := range proxys {
			match := p.MatchProxy(c.Request)
			if match {
				director := func(req *http.Request) {
					req.URL = p.UpstreamURL
					req.Header = *p.UpstreamHeader
				}
				proxy := &httputil.ReverseProxy{Director: director}
				proxy.ServeHTTP(c.ResponseWriter, c.Request)
				c.Abort()
				return nil
			}
		}
		return nil
	}
}
