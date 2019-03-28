// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routing

import (
	"strings"
)

// RouteGroup represents a group of routes that share the same path prefix.
type RouteGroup struct {
	prefix   string
	router   *Router
	handlers []Handler
	// 以下预处理Handler
	startupHandlers   []Handler
	anteriorHandlers  []Handler
	posteriorHandlers []Handler
	shutdownHandlers  []Handler
}

// newRouteGroup creates a new RouteGroup with the given path prefix, router, and handlers.
func newRouteGroup(prefix string, router *Router, handlers []Handler, startupHandlers []Handler, anteriorHandlers []Handler, posteriorHandlers []Handler, shutdownHandlers []Handler) *RouteGroup {
	return &RouteGroup{
		prefix:            prefix,
		router:            router,
		handlers:          handlers,
		startupHandlers:   startupHandlers,
		anteriorHandlers:  anteriorHandlers,
		posteriorHandlers: posteriorHandlers,
		shutdownHandlers:  shutdownHandlers,
	}
}

func (rg *RouteGroup) GetStartupHandlers() []Handler {
	return rg.startupHandlers
}

func (rg *RouteGroup) GetAnteriorHandlers() []Handler {
	return rg.anteriorHandlers
}

func (rg *RouteGroup) GetPosteriorHandlers() []Handler {
	return rg.posteriorHandlers
}

func (rg *RouteGroup) GetShutdownHandlers() []Handler {
	return rg.shutdownHandlers
}

// Get adds a GET route to the router with the given route path and handlers.
func (rg *RouteGroup) Get(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("GET", schemes, hosts, path, query, headers, handlers)
}

// Post adds a POST route to the router with the given route path and handlers.
func (rg *RouteGroup) Post(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("POST", schemes, hosts, path, query, headers, handlers)
}

// Put adds a PUT route to the router with the given route path and handlers.
func (rg *RouteGroup) Put(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("PUT", schemes, hosts, path, query, headers, handlers)
}

// Patch adds a PATCH route to the router with the given route path and handlers.
func (rg *RouteGroup) Patch(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("PATCH", schemes, hosts, path, query, headers, handlers)
}

// Delete adds a DELETE route to the router with the given route path and handlers.
func (rg *RouteGroup) Delete(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("DELETE", schemes, hosts, path, query, headers, handlers)
}

// Connect adds a CONNECT route to the router with the given route path and handlers.
func (rg *RouteGroup) Connect(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("CONNECT", schemes, hosts, path, query, headers, handlers)
}

// Head adds a HEAD route to the router with the given route path and handlers.
func (rg *RouteGroup) Head(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("HEAD", schemes, hosts, path, query, headers, handlers)
}

// Options adds an OPTIONS route to the router with the given route path and handlers.
func (rg *RouteGroup) Options(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("OPTIONS", schemes, hosts, path, query, headers, handlers)
}

// Trace adds a TRACE route to the router with the given route path and handlers.
func (rg *RouteGroup) Trace(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.add("TRACE", schemes, hosts, path, query, headers, handlers)
}

// Any adds a route with the given route, handlers, and the HTTP methods as listed in routing.Methods.
func (rg *RouteGroup) Any(schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	return rg.To(strings.Join(Methods, ","), schemes, hosts, path, query, headers, handlers...)
}

// To adds a route to the router with the given HTTP methods, route path, and handlers.
// Multiple HTTP methods should be separated by commas (without any surrounding spaces).
func (rg *RouteGroup) To(methods, schemes, hosts, path string, query, headers map[string]string, handlers ...Handler) *Route {
	mm := strings.Split(methods, ",")
	ss := strings.Split(schemes, ",")
	hh := strings.Split(hosts, ",")
	if len(mm) == 1 && len(ss) == 1 && len(hh) == 1 {
		return rg.add(methods, schemes, hosts, path, query, headers, handlers)
	}
	r := rg.newRoute(methods, schemes, hosts, path, query, headers)
	for _, method := range mm {
		if len(ss) > 0 && len(hh) > 0 {
			for _, scheme := range ss {
				for _, host := range hh {
					r.routes = append(r.routes, rg.add(method, scheme, host, path, query, headers, handlers))
				}
			}
		} else if len(ss) > 0 {
			for _, scheme := range ss {
				r.routes = append(r.routes, rg.add(method, scheme, hosts, path, query, headers, handlers))
			}
		} else if len(hh) > 0 {
			for _, host := range hh {
				r.routes = append(r.routes, rg.add(method, schemes, host, path, query, headers, handlers))
			}
		} else {
			r.routes = append(r.routes, rg.add(method, schemes, hosts, path, query, headers, handlers))
		}
	}
	return r
}

// Group creates a RouteGroup with the given route path prefix and handlers.
// The new group will combine the existing path prefix with the new one.
// If no handler is provided, the new group will inherit the handlers registered
// with the current group.
func (rg *RouteGroup) Group(prefix string, handlers ...Handler) *RouteGroup {
	if len(handlers) == 0 {
		handlers = make([]Handler, len(rg.handlers))
		copy(handlers, rg.handlers)
	}
	return newRouteGroup(rg.prefix+prefix, rg.router, handlers,
		rg.startupHandlers, rg.anteriorHandlers, rg.posteriorHandlers, rg.shutdownHandlers)
}

// Startup registers one or multiple handlers to the current route group.
// These handlers will be shared by all routes belong to this group and its subgroups.
func (rg *RouteGroup) AppendStartupHandler(handlers ...Handler) {
	rg.startupHandlers = append(rg.startupHandlers, handlers...)
}

func (rg *RouteGroup) AppendShutdownHandler(handlers ...Handler) {
	rg.shutdownHandlers = append(rg.shutdownHandlers, handlers...)
}

func (rg *RouteGroup) AppendAnteriorHandler(handlers ...Handler) {
	rg.anteriorHandlers = append(rg.anteriorHandlers, handlers...)
}

func (rg *RouteGroup) AppendPosteriorHandler(handlers ...Handler) {
	rg.posteriorHandlers = append(rg.posteriorHandlers, handlers...)
}

// Use registers one or multiple handlers to the current route group.
// These handlers will be shared by all routes belong to this group and its subgroups.
func (rg *RouteGroup) Use(handlers ...Handler) {
	rg.handlers = append(rg.handlers, handlers...)
}

func (rg *RouteGroup) add(methods, schemes, hosts, path string, query, headers map[string]string, handlers []Handler) *Route {
	r := rg.newRoute(methods, schemes, hosts, path, query, headers)
	rg.router.addRoute(r, combineHandlers(rg.handlers, handlers))
	return r
}

// newRoute creates a new Route with the given route path and route group.
func (rg *RouteGroup) newRoute(methods, schemes, hosts, path string, query, headers map[string]string) *Route {
	return &Route{
		group:    rg,
		schemes:  schemes,
		hosts:    hosts,
		methods:  methods,
		path:     path,
		headers:  headers,
		query:    query,
		template: buildURLTemplate(methods, schemes, hosts, rg.prefix+path, query, headers),
	}
}

// combineHandlers merges multiple lists of handlers into a new list.
func combineHandlers(h ...[]Handler) []Handler {
	var (
		l       int = 0
		handler []Handler
	)
	for _, handler = range h {
		l += len(handler)
	}
	var hh []Handler = make([]Handler, l)
	l = 0
	for _, handler = range h {
		copy(hh[l:], handler)
		l += len(handler)
	}
	return hh
}

// buildURLTemplate converts a route pattern into a URL template by removing regular expressions in parameter tokens.
func buildURLTemplate(methods, schemes, hosts, path string, query, headers map[string]string) string {
	path = strings.TrimRight(path, "*")
	template, start, end := "", -1, -1
	template += methods + " " + schemes + "://" + hosts
	for i := 0; i < len(path); i++ {
		if path[i] == '<' && start < 0 {
			start = i
		} else if path[i] == '>' && start >= 0 {
			name := path[start+1 : i]
			for j := start + 1; j < i; j++ {
				if path[j] == ':' {
					name = path[start+1 : j]
					break
				}
			}
			template += path[end+1:start] + "<" + name + ">"
			end = i
			start = -1
		}
	}
	if end < 0 {
		template += path
	} else if end < len(path)-1 {
		template += path[end+1:]
	}
	queryNum := len(query)
	if queryNum > 0 {
		template += "?"
		for key, val := range query {
			queryNum--
			if queryNum > 0 {
				template += key + "=" + val + "&"
			} else {
				template += key + "=" + val
			}
		}
	}

	headerNum := len(headers)
	if headerNum > 0 {
		template += " "
		for key, val := range headers {
			headerNum--
			if headerNum > 0 {
				template += key + ":" + val + "\n"
			} else {
				template += key + ":" + val
			}
		}
	}

	return template
}
