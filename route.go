// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routing

import (
	"fmt"
	"net/url"
	"strings"
)

// Route represents a URL path pattern that can be used to match requested URLs.
type Route struct {
	group                         *RouteGroup
	schemes, hosts, methods, path string
	query                         map[string]string
	headers                       map[string]string
	name, template                string
	tags                          []interface{}
	routes                        []*Route
}

// Name sets the name of the route.
// This method will update the registration of the route in the router as well.
func (r *Route) Name(name string) *Route {
	r.name = name
	r.group.router.namedRoutes[name] = r
	return r
}

// Tag associates some custom data with the route.
func (r *Route) Tag(value interface{}) *Route {
	if len(r.routes) > 0 {
		// this route is a composite one (a path with multiple methods)
		for _, route := range r.routes {
			route.Tag(value)
		}
		return r
	}
	if r.tags == nil {
		r.tags = []interface{}{}
	}
	r.tags = append(r.tags, value)
	return r
}

// Method returns the HTTP method that this route is associated with.
func (r *Route) Method(methods string) *Route {
	r.methods = methods
	return r
}
func (r *Route) GetMethod() string {
	return r.methods
}

// Schemes returns the request schemes that this route should match.
func (r *Route) Scheme(schemes string) *Route {
	r.schemes = schemes
	return r
}
func (r *Route) GetScheme() string {
	return r.schemes
}

// Hosts returns the request hosts that this route should match.
func (r *Route) Host(hosts string) *Route {
	r.hosts = hosts
	return r
}
func (r *Route) GetHost() string {
	return r.hosts
}

// Headers returns the request headers that this route should match.
func (r *Route) Headers(headers map[string]string) *Route {
	r.headers = headers
	return r
}
func (r *Route) GetHeader() map[string]string {
	return r.headers
}

// Path returns the request path that this route should match.
func (r *Route) Path(path string) *Route {
	r.path = path
	return r
}
func (r *Route) GetPath(path string) string {
	return r.path
}

// Queries returns the request queries that this route should match.
func (r *Route) Query(query map[string]string) *Route {
	r.query = query
	return r
}
func (r *Route) GetQuery() map[string]string {
	return r.query
}

// Tags returns all custom data associated with the route.
func (r *Route) Tags(tags []interface{}) *Route {
	r.tags = tags
	return r
}
func (r *Route) GetTags() []interface{} {
	return r.tags
}

// Get adds the route to the router using the GET HTTP method.
func (r *Route) Get(handlers ...Handler) *Route {
	return r.group.add("GET", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

// Post adds the route to the router using the POST HTTP method.
func (r *Route) Post(handlers ...Handler) *Route {
	return r.group.add("POST", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

// Put adds the route to the router using the PUT HTTP method.
func (r *Route) Put(handlers ...Handler) *Route {
	return r.group.add("PUT", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

// Patch adds the route to the router using the PATCH HTTP method.
func (r *Route) Patch(handlers ...Handler) *Route {
	return r.group.add("PATCH", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

// Delete adds the route to the router using the DELETE HTTP method.
func (r *Route) Delete(handlers ...Handler) *Route {
	return r.group.add("DELETE", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

// Connect adds the route to the router using the CONNECT HTTP method.
func (r *Route) Connect(handlers ...Handler) *Route {
	return r.group.add("CONNECT", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

// Head adds the route to the router using the HEAD HTTP method.
func (r *Route) Head(handlers ...Handler) *Route {
	return r.group.add("HEAD", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

// Options adds the route to the router using the OPTIONS HTTP method.
func (r *Route) Options(handlers ...Handler) *Route {
	return r.group.add("OPTIONS", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

// Trace adds the route to the router using the TRACE HTTP method.
func (r *Route) Trace(handlers ...Handler) *Route {
	return r.group.add("TRACE", r.schemes, r.hosts, r.path, r.query, r.headers, handlers)
}

func (r *Route) To(methods string, handlers ...Handler) *Route {
	return r.group.To(methods, r.schemes, r.hosts, r.path, r.query, r.headers, handlers...)
}

// URL creates a URL using the current route and the given parameters.
// The parameters should be given in the sequence of name1, value1, name2, value2, and so on.
// If a parameter in the route is not provided a value, the parameter token will remain in the resulting URL.
// The method will perform URL encoding for all given parameter values.
func (r *Route) URL(pairs ...interface{}) (s string) {
	s = r.template
	for i := 0; i < len(pairs); i++ {
		name := fmt.Sprintf("<%v>", pairs[i])
		value := ""
		if i < len(pairs)-1 {
			value = url.QueryEscape(fmt.Sprint(pairs[i+1]))
		}
		s = strings.Replace(s, name, value, -1)
	}
	return
}

// String returns the string representation of the route.
func (r *Route) String() string {
	query := ""
	queryNum := len(r.query)
	if queryNum > 0 {
		query += "?"
		for key, val := range r.query {
			queryNum--
			if queryNum > 0 {
				query += key + "=" + val + "&"
			} else {
				query += key + "=" + val
			}
		}
	}
	header := ""
	headerNum := len(r.headers)
	if headerNum > 0 {
		header += " "
		for key, val := range r.headers {
			headerNum--
			if headerNum > 0 {
				header += key + ":" + val + "\n"
			} else {
				header += key + ":" + val
			}
		}
	}
	return r.methods + " " + r.schemes + "://" + r.hosts + r.group.prefix + r.path + query + header
}
