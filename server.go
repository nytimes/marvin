package marvin

import (
	"context"
	"errors"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/golang/protobuf/proto"

	"github.com/NYTimes/marvin/internal"
)

type contextKey int

const (
	// ContextKeyInboundAppID is populated in the context by default.
	// It contains the value of the 'X-Appengine-Inbound-Appid' header.
	ContextKeyInboundAppID contextKey = iota
	// key to set/retrieve URL params from a
	// Gorilla request context.
	varsKey
)

var defaultOpts = []httptransport.ServerOption{
	httptransport.ServerBefore(
		// init the App Engine context first
		func(ctx context.Context, r *http.Request) context.Context {
			return context.WithValue(ctx, ContextKeyInboundAppID, r.Header.Get("X-Appengine-Inbound-Appid"))

		},
		// populate context with helpful keys
		httptransport.PopulateRequestContext),
}

// Init will register the Service with a Server
// and register the server with App Engine.
// Call this in an `init()` or `main()` function.
func Init(service Service) {
	http.Handle("/", NewServer(service))
}

// Server manages routing and initiating the request context.
// Users should only need to interact with this struct in testing.
//
// See examples/reading-list/api/service_test.go for example usage.
type Server struct {
	mux Router
	svc Service
}

// NewServer will init the mux and register all endpoints.
// This gets called by Init() and should only be used within
// tests.
//
// See examples/reading-list/api/service_test.go for example usage.
func NewServer(svc Service) Server {
	opts := svc.RouterOptions()
	if len(opts) == 0 {
		// select the default router
		opts = append(opts, RouterSelect(""))
	}
	var r Router
	for _, opt := range opts {
		r = opt(r)
	}
	svr := Server{
		mux: r,
		svc: svc,
	}
	err := svr.register(svc)
	if err != nil {
		panic("unable to register service: " + err.Error())
	}

	return svr
}

// ServeHTTP is the entrypoint for the server. This will initiate
// the app engine context and hand the request off to the router.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(internal.NewContext(r))
	s.svc.HTTPMiddleware(s.mux).ServeHTTP(w, r)
}

// register will accept and register server, JSONService or MixedService implementations.
func (s Server) register(svc Service) error {
	var (
		jseps map[string]map[string]HTTPEndpoint
		peps  map[string]map[string]HTTPEndpoint
	)
	switch svc.(type) {
	case MixedService:
		jseps = svc.(JSONService).JSONEndpoints()
		peps = svc.(ProtoService).ProtoEndpoints()
	case JSONService:
		jseps = svc.(JSONService).JSONEndpoints()
	case ProtoService:
		peps = svc.(ProtoService).ProtoEndpoints()
	default:
		return errors.New("services for servers must implement one of the Service interface extensions")
	}

	opts := defaultOpts
	opts = append(opts, svc.Options()...)

	// so we can add a /_ah/warmup if none provided
	var warmupExists bool

	// register all JSON endpoints with our wrappers & default decoders/encoders
	for path, epMethods := range jseps {
		for method, ep := range epMethods {
			if path == warmupURI && method == http.MethodGet {
				warmupExists = true
			}
			// just pass the http.Request in if no decoder provided
			if ep.Decoder == nil {
				ep.Decoder = func(_ context.Context, r *http.Request) (interface{}, error) {
					return r, nil
				}
			}
			// default to the httptransport helper
			if ep.Encoder == nil {
				ep.Encoder = httptransport.EncodeJSONResponse
			}
			s.mux.Handle(method, path,
				httptransport.NewServer(
					svc.Middleware(ep.Endpoint),
					ep.Decoder,
					ep.Encoder,
					append(opts, ep.Options...)...))
		}
	}

	// register all Protobuf endpoints with our wrappers & default decoders/encoders
	for path, epMethods := range peps {
		for method, ep := range epMethods {
			if path == warmupURI && method == http.MethodGet {
				warmupExists = true
			}
			// just pass the http.Request in if no decoder provided
			if ep.Decoder == nil {
				ep.Decoder = func(_ context.Context, r *http.Request) (interface{}, error) {
					return r, nil
				}
			}
			// default to the a protobuf helper
			if ep.Encoder == nil {
				ep.Encoder = EncodeProtoResponse
			}
			s.mux.Handle(method, path,
				httptransport.NewServer(
					svc.Middleware(ep.Endpoint),
					ep.Decoder,
					ep.Encoder,
					append(opts, ep.Options...)...))
		}
	}

	// add a warmup hook if one doesn't already exist
	if !warmupExists {
		s.mux.HandleFunc("GET", warmupURI,
			func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	}
	return nil
}

const warmupURI = "/_ah/warmup"

// EncodeProtoResponse is an httptransport.EncodeResponseFunc that serializes the response
// as Protobuf. Many Proto-over-HTTP services can use it as a sensible default. If the
// response implements Headerer, the provided headers will be applied to the response.
// If the response implements StatusCoder, the provided StatusCode will be used instead
// of 200.
func EncodeProtoResponse(ctx context.Context, w http.ResponseWriter, pres interface{}) error {
	res, ok := pres.(proto.Message)
	if !ok {
		return errors.New("response does not implement proto.Message")
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	if headerer, ok := w.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}
	code := http.StatusOK
	if sc, ok := pres.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}
	if res == nil {
		return nil
	}
	b, err := proto.Marshal(res)
	if err != nil {
		// maybe log instead? need to avoid a second header write
		return nil
	}
	_, err = w.Write(b)
	if err != nil {
		// maybe log instead? need to avoid a second header write
	}
	return nil
}
