package marvin

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/julienschmidt/httprouter"
)

// Router is an interface to wrap different router implementations.
type Router interface {
	Handle(method string, path string, handler http.Handler)
	HandleFunc(method string, path string, handlerFunc func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	SetNotFoundHandler(handler http.Handler)
}

// RouterOption sets optional Router overrides.
type RouterOption func(Router) Router

// RouterSelect allows users to override the default use of the GorillaRouter with one of
// the other implementations.
//
// The following router names are accepted:
// * `gorilla` which uses github.com/gorilla/mux (on by default)
// * `stdlib` to utilize the standard library's http.ServeMux
// * `httprouter` which uses github.com/julienschmidt/httprouter
//
// If the supplied name does not match any known router, `gorilla` will be used.
// If a user wishes to supply their own router implementation, the `CustomRouter` option
// is available.
func RouterSelect(name string) RouterOption {
	return func(_ Router) Router {
		switch name {
		case "httprouter":
			return &FastRouter{httprouter.New()}
		case "gorilla":
			return &GorillaRouter{mux.NewRouter()}
		case "stdlib":
			return &StdlibRouter{http.NewServeMux()}
		default:
			return &GorillaRouter{mux.NewRouter()}
		}
	}
}

// CustomRouter allows users to inject an alternate Router implementation.
func CustomRouter(r Router) RouterOption {
	return func(_ Router) Router {
		return r
	}
}

// RouterNotFound will set the not found handler of the router.
func RouterNotFound(h http.Handler) RouterOption {
	return func(r Router) Router {
		r.SetNotFoundHandler(h)
		return r
	}
}

// StdlibRouter is a Router implementation for the Stdlib's `http.ServeMux`.
type StdlibRouter struct {
	mux *http.ServeMux
}

// Handle will call the Stdlib's HandleFunc() methods with a check for the incoming
// HTTP method. To allow for multiple methods on a single route, use 'ANY'.
func (g *StdlibRouter) Handle(method, path string, h http.Handler) {
	g.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method || method == "ANY" {
			h.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}

// HandleFunc will call the Stdlib's HandleFunc() methods with a check for the incoming
// HTTP method. To allow for multiple methods on a single route, use 'ANY'.
func (g *StdlibRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	g.Handle(method, path, http.HandlerFunc(h))
}

// SetNotFoundHandler will do nothing as we cannot override the stdlib not found.
func (g *StdlibRouter) SetNotFoundHandler(h http.Handler) {
}

// ServeHTTP will call Stdlib's ServeMux.ServerHTTP directly.
func (g *StdlibRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

// GorillaRouter is a Router implementation for the Gorilla web toolkit's `mux.Router`.
type GorillaRouter struct {
	mux *mux.Router
}

// Handle will call the Gorilla web toolkit's Handle().Method() methods.
func (g *GorillaRouter) Handle(method, path string, h http.Handler) {
	g.mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// copy the route params into a shared location
		// duplicating memory, but allowing Gizmo to be more flexible with
		// router implementations.
		r = SetRouteVars(r, mux.Vars(r))
		h.ServeHTTP(w, r)
	})).Methods(method)
}

// HandleFunc will call the Gorilla web toolkit's HandleFunc().Method() methods.
func (g *GorillaRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	g.Handle(method, path, http.HandlerFunc(h))
}

// SetNotFoundHandler will set the Gorilla mux.Router.NotFoundHandler.
func (g *GorillaRouter) SetNotFoundHandler(h http.Handler) {
	g.mux.NotFoundHandler = h
}

// ServeHTTP will call Gorilla mux.Router.ServerHTTP directly.
func (g *GorillaRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

// FastRouter is a Router implementation for `julienschmidt/httprouter`.
type FastRouter struct {
	mux *httprouter.Router
}

// Handle will call the `httprouter.METHOD` methods and pass the httprouter.Params into
// the request context. The params will be available via the `Vars` function.
func (f *FastRouter) Handle(method, path string, h http.Handler) {
	f.mux.Handle(method, path, httpToFastRoute(h))
}

// HandleFunc will call the `httprouter.METHOD` methods and pass the httprouter.Params into
// the request context. The params will be available via the `Vars` function.
func (f *FastRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	f.Handle(method, path, http.HandlerFunc(h))
}

// SetNotFoundHandler will set httprouter.Router.NotFound.
func (f *FastRouter) SetNotFoundHandler(h http.Handler) {
	f.mux.NotFound = h
}

// ServeHTTP will call httprouter.ServerHTTP directly.
func (f *FastRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mux.ServeHTTP(w, r)
}

// httpToFastRoute will convert an http.Handler to a httprouter.Handle
// by stuffing any route parameters into the request context.
// To access the request parameters within the endpoint,
// use the `Vars` function.
func httpToFastRoute(fh http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if len(params) > 0 {
			vars := map[string]string{}
			for _, param := range params {
				vars[param.Key] = param.Value
			}
			r = SetRouteVars(r, vars)
		}
		fh.ServeHTTP(w, r)
	}
}

// Vars is a helper function for accessing route
// parameters from any server.Router implementation. This is the equivalent
// of using `mux.Vars(r)` with the Gorilla mux.Router.
func Vars(r *http.Request) map[string]string {
	if rv := r.Context().Value(varsKey); rv != nil {
		vars, _ := rv.(map[string]string)
		return vars
	}
	return nil
}

// SetRouteVars will set the given value into into the request context
// with the shared 'vars' storage key.
func SetRouteVars(r *http.Request, val map[string]string) *http.Request {
	if val != nil {
		r = r.WithContext(context.WithValue(r.Context(), varsKey, val))
	}
	return r
}
