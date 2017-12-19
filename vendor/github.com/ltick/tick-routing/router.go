// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package routing provides high performance and powerful HTTP routing capabilities.
package routing

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type (
	// Handler is the function for handling HTTP requests.
	Handler func(context.Context, *Context) (context.Context, error)

	// Router manages routes and dispatches HTTP requests to the handlers of the matching routes.
	Router struct {
		RouteGroup
		IgnoreTrailingSlash bool // whether to ignore trailing slashes in the end of the request URL
		pool                sync.Pool
		routes              []*Route
		namedRoutes         map[string]*Route
		stores              map[string]routeStore
		maxParams           int
		notFound            []Handler
		notFoundHandlers    []Handler

		CancelHandlers  []Handler
		TimeoutHandlers []Handler
		TimeoutDuration time.Duration
		Context         context.Context
	}

	// routeStore stores route paths and the corresponding handlers.
	routeStore interface {
		Add(key string, data interface{}) int
		Get(key string, pvalues []string) (data interface{}, pnames []string)
		String() string
	}
)

// Methods lists all supported HTTP methods by Router.
var Methods = []string{
	"CONNECT",
	"DELETE",
	"GET",
	"HEAD",
	"OPTIONS",
	"PATCH",
	"POST",
	"PUT",
	"TRACE",
}

// New creates a new Router object.
func New(ctx context.Context) *Router {
	r := &Router{
		namedRoutes:     make(map[string]*Route),
		stores:          make(map[string]routeStore),
		TimeoutDuration: 0 * time.Second,
		TimeoutHandlers: []Handler{TimeoutHandler},
		Context:         ctx,
	}
	r.RouteGroup = *newRouteGroup("", r, make([]Handler, 0), make([]Handler, 0), make([]Handler, 0))
	r.NotFound(MethodNotAllowedHandler, NotFoundHandler)
	r.pool.New = func() interface{} {
		return &Context{
			pvalues: make([]string, r.maxParams),
			router:  r,
		}
	}
	return r
}

// ServeHTTP handles the HTTP request.
// It is required by http.Handler
func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	c := r.pool.Get().(*Context)
	c.init(res, req)
	// timeout & cancel
	if r.TimeoutDuration != 0*time.Second {
		c.Ctx, c.CancelFunc = context.WithTimeout(c.Ctx, r.TimeoutDuration)
	} else {
		c.Ctx, c.CancelFunc = context.WithCancel(c.Ctx)
	}
	c.Request.WithContext(c.Ctx)
	c.handlers, c.pnames = r.find(req.Method, r.normalizeRequestPath(req.URL.Path), c.pvalues)
	// next
	if err := c.Next(); err != nil {
		r.handleError(c, err)
	}
	r.pool.Put(c)
}

// Route returns the named route.
// Nil is returned if the named route cannot be found.
func (r *Router) Route(name string) *Route {
	return r.namedRoutes[name]
}

// Routes returns all routes managed by the router.
func (r *Router) Routes() []*Route {
	return r.routes
}

// Use appends the specified handlers to the router and shares them with all routes.
func (r *Router) Use(handlers ...Handler) {
	r.RouteGroup.Use(handlers...)
	r.notFoundHandlers = combineHandlers(r.handlers, r.notFound)
}

func (r *Router) PrependAnterior(handlers ...Handler) {
	r.RouteGroup.PrependAnterior(handlers...)
}

func (r *Router) AppendAnterior(handlers ...Handler) {
	r.RouteGroup.AppendAnterior(handlers...)
}

func (r *Router) PrependPosterior(handlers ...Handler) {
	r.RouteGroup.PrependPosterior(handlers...)
}

func (r *Router) AppendPosterior(handlers ...Handler) {
	r.RouteGroup.AppendPosterior(handlers...)
}

// NotFound specifies the handlers that should be invoked when the router cannot find any route matching a request.
// Note that the handlers registered via Use will be invoked first in this case.
func (r *Router) NotFound(handlers ...Handler) {
	r.notFound = handlers
	r.notFoundHandlers = combineHandlers(r.handlers, r.notFound)
}

func (r *Router) Timeout(timeoutDuration time.Duration, handlers ...Handler) {
	r.TimeoutDuration = timeoutDuration
	if len(handlers) > 0 {
		r.TimeoutHandlers = handlers
	}
}

func (r *Router) Cancel(handlers ...Handler) {
	r.CancelHandlers = handlers
}

// handleError is the error handler for handling any unhandled errors.
func (r *Router) handleError(c *Context, err error) {
	if httpError, ok := err.(HTTPError); ok {
		http.Error(c.Response, httpError.Error(), httpError.StatusCode())
	} else {
		http.Error(c.Response, err.Error(), http.StatusInternalServerError)
	}
}

func (r *Router) addRoute(route *Route, handlers []Handler) {
	path := route.group.prefix + route.path

	r.routes = append(r.routes, route)

	store := r.stores[route.method]
	if store == nil {
		store = newStore()
		r.stores[route.method] = store
	}

	// an asterisk at the end matches any number of characters
	if strings.HasSuffix(path, "*") {
		path = path[:len(path)-1] + "<:.*>"
	}

	if n := store.Add(path, handlers); n > r.maxParams {
		r.maxParams = n
	}
}

func (r *Router) find(method, path string, pvalues []string) (handlers []Handler, pnames []string) {
	var hh interface{}
	if store := r.stores[method]; store != nil {
		hh, pnames = store.Get(path, pvalues)
	}
	if hh != nil {
		return hh.([]Handler), pnames
	}
	return r.notFoundHandlers, pnames
}

func (r *Router) findAllowedMethods(path string) map[string]bool {
	methods := make(map[string]bool)
	pvalues := make([]string, r.maxParams)
	for m, store := range r.stores {
		if handlers, _ := store.Get(path, pvalues); handlers != nil {
			methods[m] = true
		}
	}
	return methods
}

func (r *Router) normalizeRequestPath(path string) string {
	if r.IgnoreTrailingSlash && len(path) > 1 && path[len(path)-1] == '/' {
		for i := len(path) - 2; i > 0; i-- {
			if path[i] != '/' {
				return path[0 : i+1]
			}
		}
		return path[0:1]
	}
	return path
}

// TimeoutHandler returns a 408 HTTP error indicating a request execute timeout.
func TimeoutHandler(ctx context.Context, c *Context) error {
	return ctx, NewHTTPError(http.StatusRequestTimeout)
}

// NotFoundHandler returns a 404 HTTP error indicating a request has no matching route.
func NotFoundHandler(ctx context.Context, c *Context) error {
	return ctx, NewHTTPError(http.StatusNotFound)
}

// MethodNotAllowedHandler handles the situation when a request has matching route without matching HTTP method.
// In this case, the handler will respond with an Allow HTTP header listing the allowed HTTP methods.
// Otherwise, the handler will do nothing and let the next handler (usually a NotFoundHandler) to handle the problem.
func MethodNotAllowedHandler(ctx context.Context, c *Context) error {
	methods := c.Router().findAllowedMethods(c.Request.URL.Path)
	if len(methods) == 0 {
		return ctx, nil
	}
	methods["OPTIONS"] = true
	ms := make([]string, len(methods))
	i := 0
	for method := range methods {
		ms[i] = method
		i++
	}
	sort.Strings(ms)
	c.Response.Header().Set("Allow", strings.Join(ms, ", "))
	if c.Request.Method != "OPTIONS" {
		c.Response.WriteHeader(http.StatusMethodNotAllowed)
	}
	c.Abort()
	return ctx, nil
}

// HTTPHandlerFunc adapts a http.HandlerFunc into a routing.Handler.
func HTTPHandlerFunc(h http.HandlerFunc) Handler {
	return func(ctx context.Context, c *Context) error {
		h(c.Response, c.Request)
		return ctx, nil
	}
}

// HTTPHandler adapts a http.Handler into a routing.Handler.
func HTTPHandler(h http.Handler) Handler {
	return func(ctx context.Context, c *Context) error {
		h.ServeHTTP(c.Response, c.Request)
		return ctx, nil
	}
}
