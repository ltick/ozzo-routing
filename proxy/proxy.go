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
	Host           []string
	Group          string
	Path           string
	Upstream       string
	UpstreamHeader http.Header
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
			captures := make(map[string]string)
			r := regexp.MustCompile("<:(\\w+)>")
			match := r.FindStringSubmatch(p.Upstream)
			if match == nil {
				return errors.New("ltick: upstream not match")
			}
			for i, name := range r.SubexpNames() {
				if i == 0 || name == "" {
					continue
				}
				if name != "group" && name != "path" && name != "fragment" && name != "query" {
					captures[name] = c.Param(name)
				}
			}
			captures["group"] = p.Group
			path := c.Request.URL.Path
			if path == "" {
				path = c.Request.URL.RawPath
			}
			captures["fragment"] = c.Request.URL.Fragment
			captures["query"] = c.Request.URL.RawQuery
			captures["path"] = strings.TrimLeft(path, p.Group)
			captures["group"] = p.Group
			if len(captures) != 0 {
				upstream := p.Upstream
				//拼接配置文件指定中的uri，$符号分割
				for name, capture := range captures {
					upstream = strings.Replace(upstream, "<:"+name+">", capture, -1)
				}
				upstreamURL, err := url.Parse(upstream)
				if err != nil {
					return err
				}
				if upstreamURL != nil {
					director := func(req *http.Request) {
						req.URL.Scheme = upstreamURL.Scheme
						req.URL.Host = upstreamURL.Host
						req.Host = upstreamURL.Host
						req.RequestURI = upstreamURL.RequestURI()
						req.Header = p.UpstreamHeader
					}
					proxy := &httputil.ReverseProxy{Director: director}
					proxy.ServeHTTP(c.ResponseWriter, c.Request)
					c.Abort()
				}
			}
		}
		return nil
	}
}
