package michi

import (
	"fmt"
	"net/http"
)

// Router is a http.Handler
type Router struct {
	path string
	// handlerMiddlewares are the middlewares to be applied to the final handler
	// if the final handler is not found, the middleware is not executed
	handlerMiddlewares []func(http.Handler) http.Handler
	// subRouterMiddlewares are the middlewares to be applied to the sub router
	// even if the final handler is not found, the middlewares are executed if the path of the sub router is matched
	subRouterMiddlewares []func(http.Handler) http.Handler
	// serveMux is the http.ServeMux for Router
	serveMux *http.ServeMux
	// executedRouteOrHandle is true if the Route or Handle is executed
	executedRouteOrHandle bool
	// inGroupOrWith is true if the router is in Group or With
	inGroupOrWith bool
}

// NewRouter creates a new Router
func NewRouter() *Router {
	return newRouter("")
}

func newRouter(path string) *Router {
	return &Router{
		path:                  path,
		handlerMiddlewares:    nil,
		subRouterMiddlewares:  nil,
		serveMux:              http.NewServeMux(),
		executedRouteOrHandle: false,
		inGroupOrWith:         false,
	}
}

func (r *Router) cloneForWith() *Router {
	return &Router{
		path:                 r.path,
		subRouterMiddlewares: r.subRouterMiddlewares,
		handlerMiddlewares:   r.handlerMiddlewares,
		serveMux:             r.serveMux,
		// After executing With and Group, set executedRouteOrHandle to false. Otherwise, it will panic with r.Group -> r.Use
		executedRouteOrHandle: false,
		inGroupOrWith:         true,
	}
}

// ServeHTTP is the single method of the http.Handler interface that makes
// Mux interoperable with the standard library.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	chain(r.subRouterMiddlewares, r.serveMux).ServeHTTP(w, req)
}

// Use appends a middleware handler to the Mux middleware stack.
//
// The middleware stack for any Mux will execute before searching for a matching
// route to a specific handler, which provides opportunity to respond early,
// change the course of the request execution, or set request-scoped values for
// the next http.Handler.
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	if r.executedRouteOrHandle {
		panic("michi: all middlewares must be defined before routes on a mux")
	}
	if r.inGroupOrWith {
		r.handlerMiddlewares = append(r.handlerMiddlewares, middlewares...)
	} else {
		r.subRouterMiddlewares = append(r.subRouterMiddlewares, middlewares...)
	}
}

// With adds inline middlewares for an endpoint handler.
func (r *Router) With(middlewares ...func(http.Handler) http.Handler) *Router {
	withRouter := r.cloneForWith()
	withRouter.handlerMiddlewares = append(withRouter.handlerMiddlewares, middlewares...)
	return withRouter
}

// Group creates a new inline-Mux with a copy of middleware stack. It's useful
// for a group of handlers along the same routing path that use an additional
func (r *Router) Group(fn func(sub *Router)) {
	r2 := r.With()
	if fn != nil {
		fn(r2)
	}
}

// Route creates a new Mux and mounts it along the `pattern` as a subrouter.
func (r *Router) Route(pattern string, fn func(sub *Router)) {
	if fn == nil {
		panic(fmt.Errorf("michi: sub router function cannot be nil on '%s'", pattern))
	}
	if string(pattern[len(pattern)-1]) != "/" {
		pattern += "/"
	}

	subRouter := newRouter(joinPathAndPattern(r.path, pattern))
	fn(subRouter)
	r.serveMux.Handle(subRouter.path, subRouter)

	r.executedRouteOrHandle = true
}

// HandleFunc adds the route `pattern` that matches any http method to
// execute the `handlerFn` http.HandlerFunc.
func (r *Router) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	r.Handle(pattern, handlerFunc)
}

// Handle adds the route `pattern` that matches any http method to
// execute the `handler` http.Handler.
func (r *Router) Handle(pattern string, handler http.Handler) {
	method, path := methodAndPath(pattern)
	fullPath := joinPathAndPattern(r.path, path)
	// The chain of handlerMiddlewares is done in Handle, not ServeHTTP.
	// This is because it does not work correctly when Handle is executed after With.
	// The reason it doesn't work correctly is that a different Router is created with With,
	// and the handlerMiddlewares registered with With are not applied when ServeHTTP is executed.
	r.serveMux.Handle(joinMethodAndPath(method, fullPath), chain(r.handlerMiddlewares, handler))
	r.executedRouteOrHandle = true
}

func chain(middlewares []func(http.Handler) http.Handler, handler http.Handler) http.Handler {
	for i := range middlewares {
		handler = middlewares[len(middlewares)-1-i](handler)
	}
	return handler
}
