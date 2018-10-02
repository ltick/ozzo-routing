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

type Proxy struct {
	Host     string
	URL    string
	Upstream string
}

func (p *Proxy) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)
	r := regexp.MustCompile(p.URL)
	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}
	for i, name := range r.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		captures[name] = match[i]
	}
	return captures
}

func (p *Proxy) MatchProxy(r *http.Request) (*url.URL, error) {
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}
	RequestURL := host + r.RequestURI
	HitRule := p.FindStringSubmatchMap(RequestURL)
	if len(HitRule) != 0 {
		//命中
		upstreamURL, err := url.Parse(p.Upstream)
		if err != nil {
			return nil, err
		}
		UpstreamRequestURI := "/"
		//拼接配置文件指定中的uri，$符号分割
		pathArr := strings.Split(strings.TrimLeft(upstreamURL.Path, "/"), "$")
		for _, path := range pathArr {
			if _, ok := HitRule[path]; ok {
				UpstreamRequestURI += strings.TrimLeft(HitRule[path], "/") + "/"
			}
		}
		Upstream := upstreamURL.Scheme + "://" + upstreamURL.Host + strings.TrimRight(UpstreamRequestURI, "/")
		UpstreamURL, err := url.Parse(Upstream)
		if err != nil {
			return nil, err
		}
		return UpstreamURL, nil
	}
	return nil, nil
}

// Proxy returns a handler that calls the Proxy passed to it for every request.
// The LogWriterFunc is provided with the http.Request and LogResponseWriter objects for the
// request, as well as the elapsed time since the request first came through the middleware.
// LogWriterFunc can then do whatever logging it needs to do.
//
//     import (
//         "log"
//         "github.com/ltick/tick-routing"
//         "github.com/ltick/tick-routing/access"
//         "net/http"
//     )
//
//     func myCustomLogger(req http.Context, res access.LogResponseWriter, elapsed int64) {
//         // Do something with the request, response, and elapsed time data here
//     }
//     r := routing.New()
//     r.Use(access.CustomLogger(myCustomLogger))
func ProxyHandler(proxyRoutes []*Proxy) routing.Handler {
	return func(c *routing.Context) error {
		for _, proxy := range proxyRoutes {
			upstreamURL, err := proxy.MatchProxy(c.Request)
			if err != nil {
				return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			if upstreamURL != nil {
				director := func(req *http.Request) {
					req = c.Request
					req.URL.Scheme = upstreamURL.Scheme
					req.URL.Host = upstreamURL.Host
					req.RequestURI = upstreamURL.RequestURI()
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
